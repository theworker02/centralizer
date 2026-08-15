#!/usr/bin/env python3
# Centralizer protocol shim v2. Minimal logic. Do not duplicate runtime internals.
from __future__ import annotations

import importlib
import importlib.util
import inspect
import json
import os
import struct
import sys
import traceback
from pathlib import Path


def decode(w):
    if w is None:
        return None
    k = w.get("k")
    if k == "null":
        return None
    if k == "boolean":
        return bool(w.get("b"))
    if k in ("int",):
        return int(w.get("i", 0))
    if k == "uint":
        return int(w.get("u", 0))
    if k == "float":
        return float(w.get("f", 0.0))
    if k in ("string", "decimal"):
        return w.get("s", "")
    if k == "bytes":
        import base64

        x = w.get("x") or ""
        if isinstance(x, str):
            return base64.b64decode(x)
        return bytes(x)
    if k in ("array", "tuple"):
        return [decode(i) for i in w.get("a") or []]
    if k in ("map", "struct"):
        return {e["k"]: decode(e["v"]) for e in w.get("m") or []}
    if k == "optional":
        if w.get("p") and w.get("a"):
            return decode(w["a"][0])
        return None
    if k == "error":
        raise RuntimeError(w.get("s") or "error")
    return w.get("s")


def encode(v):
    if v is None:
        return {"k": "null"}
    if isinstance(v, bool):
        return {"k": "boolean", "b": v}
    if isinstance(v, int) and not isinstance(v, bool):
        return {"k": "int", "i": v}
    if isinstance(v, float):
        return {"k": "float", "f": v}
    if isinstance(v, str):
        return {"k": "string", "s": v}
    if isinstance(v, (bytes, bytearray)):
        import base64

        return {"k": "bytes", "x": base64.b64encode(bytes(v)).decode("ascii")}
    if isinstance(v, dict):
        return {"k": "map", "m": [{"k": str(k), "v": encode(val)} for k, val in v.items()]}
    if isinstance(v, (list, tuple)):
        return {"k": "array", "a": [encode(i) for i in v]}
    return {"k": "string", "s": str(v)}


def load_target(path: str, entry: str):
    root = Path(path)
    sys.path.insert(0, str(root))
    if entry:
        spec_path = root / entry
        if spec_path.is_file() and spec_path.suffix == ".py":
            name = spec_path.stem
            spec = importlib.util.spec_from_file_location(name, spec_path)
            mod = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(mod)
            return mod
        return importlib.import_module(entry)
    if (root / "__init__.py").exists():
        return importlib.import_module(root.name)
    py = sorted(root.glob("*.py"))
    py = [p for p in py if p.name != "__init__.py"]
    if len(py) == 1:
        spec = importlib.util.spec_from_file_location(py[0].stem, py[0])
        mod = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(mod)
        return mod
    raise RuntimeError("unable to infer python entry; set entry in the manifest")


def public_functions(mod):
    out = {}
    for name, obj in inspect.getmembers(mod, inspect.isfunction):
        if name.startswith("_"):
            continue
        if getattr(obj, "__module__", None) != mod.__name__:
            continue
        out[name] = obj
    return out


def call_fn(fn, args):
    sig = inspect.signature(fn)
    kwargs = {}
    positional = []
    params = list(sig.parameters.values())
    if len(params) == 1 and params[0].kind in (
        inspect.Parameter.POSITIONAL_ONLY,
        inspect.Parameter.POSITIONAL_OR_KEYWORD,
    ):
        if args:
            if params[0].name in args:
                return fn(args[params[0].name])
            if len(args) == 1:
                return fn(next(iter(args.values())))
        return fn()
    for p in params:
        if p.name in args:
            kwargs[p.name] = args[p.name]
        elif p.default is inspect.Parameter.empty and p.kind != inspect.Parameter.VAR_KEYWORD:
            if p.kind in (inspect.Parameter.VAR_POSITIONAL,):
                continue
            raise TypeError("missing argument %s" % p.name)
    return fn(*positional, **kwargs)


def describe(fns):
    lines = ["service: python_target", "inferred: true", "functions:"]
    for name, fn in fns.items():
        lines.append("  %s:" % name)
        lines.append("    args:")
        for p in inspect.signature(fn).parameters.values():
            lines.append("      %s: string" % p.name)
        lines.append("    returns:")
        lines.append("      type: string")
    return "\n".join(lines) + "\n"


def is_streamable(v):
    if inspect.isgenerator(v) or inspect.isasyncgen(v):
        return True
    if isinstance(v, (str, bytes, bytearray, dict)):
        return False
    return hasattr(v, "__iter__")


class IO:
    def __init__(self, sock=None):
        self.sock = sock

    def send(self, msg):
        body = json.dumps(msg, separators=(",", ":")).encode("utf-8")
        if self.sock is None:
            sys.stdout.buffer.write(body + b"\n")
            sys.stdout.buffer.flush()
            return
        self.sock.sendall(struct.pack(">I", len(body)) + body)

    def messages(self):
        if self.sock is None:
            for line in sys.stdin.buffer:
                line = line.strip()
                if not line:
                    continue
                yield json.loads(line)
            return
        while True:
            hdr = recvall(self.sock, 4)
            if not hdr:
                return
            n = struct.unpack(">I", hdr)[0]
            if n <= 0 or n > 16 * 1024 * 1024:
                raise RuntimeError("invalid frame length")
            body = recvall(self.sock, n)
            if not body:
                return
            yield json.loads(body)


def recvall(sock, n):
    buf = b""
    while len(buf) < n:
        chunk = sock.recv(n - len(buf))
        if not chunk:
            return b""
        buf += chunk
    return buf


def handle(io, mod, fns, handles, state, msg):
    mid = msg.get("id", "")
    typ = msg.get("type")
    payload = msg.get("payload") or {}
    if typ == "HELLO":
        io.send(
            {
                "v": 1,
                "id": mid,
                "type": "HELLO",
                "payload": {
                    "protocol": "1.0",
                    "name": "centralizer-python",
                    "features": ["call", "handles", "describe", "stream"],
                    "runtime": "CPython",
                    "version": sys.version.split()[0],
                },
            }
        )
        return False
    if typ == "DESCRIBE":
        io.send({"v": 1, "id": mid, "type": "DESCRIBE_OK", "payload": {"schema": describe(fns)}})
        return False
    if typ == "CALL":
        name = payload.get("function")
        if name not in fns:
            raise KeyError("unknown function %s" % name)
        args = {k: decode(v) for k, v in (payload.get("args") or {}).items()}
        result = call_fn(fns[name], args)
        io.send({"v": 1, "id": mid, "type": "RESULT", "payload": {"value": encode(result)}})
        return False
    if typ == "STREAM_OPEN":
        name = payload.get("name")
        sid = payload.get("stream") or mid
        if name not in fns:
            raise KeyError("unknown function %s" % name)
        args = {k: decode(v) for k, v in (payload.get("args") or {}).items()}
        io.send({"v": 1, "id": mid, "type": "STREAM_OPEN", "payload": {"stream": sid}})
        result = call_fn(fns[name], args)
        if is_streamable(result):
            for item in result:
                io.send(
                    {
                        "v": 1,
                        "id": sid,
                        "type": "STREAM_DATA",
                        "payload": {"stream": sid, "value": encode(item)},
                    }
                )
        else:
            io.send(
                {
                    "v": 1,
                    "id": sid,
                    "type": "STREAM_DATA",
                    "payload": {"stream": sid, "value": encode(result)},
                }
            )
        io.send({"v": 1, "id": sid, "type": "STREAM_CLOSE", "payload": {"stream": sid}})
        return False
    if typ == "HANDLE_CREATE":
        tname = payload.get("type")
        cls = getattr(mod, tname, None)
        if cls is None:
            raise KeyError("unknown type %s" % tname)
        args = {k: decode(v) for k, v in (payload.get("args") or {}).items()}
        obj = cls(**args) if args else cls()
        state["hid"] += 1
        hid_s = "py-%d" % state["hid"]
        handles[hid_s] = obj
        io.send({"v": 1, "id": mid, "type": "RESULT", "payload": {"value": {"k": "handle", "s": hid_s}}})
        return False
    if typ == "HANDLE_RELEASE":
        handles.pop(payload.get("handle"), None)
        io.send({"v": 1, "id": mid, "type": "OK"})
        return False
    if typ == "HEARTBEAT":
        io.send({"v": 1, "id": mid, "type": "OK"})
        return False
    if typ == "CANCEL":
        io.send({"v": 1, "id": mid, "type": "OK"})
        return False
    if typ == "SHUTDOWN":
        io.send({"v": 1, "id": mid, "type": "OK"})
        return True
    io.send(
        {
            "v": 1,
            "id": mid,
            "type": "ERROR",
            "payload": {"code": "protocol", "message": "unsupported %s" % typ},
        }
    )
    return False


def connect_tcp(addr):
    import socket

    host, port = addr.rsplit(":", 1)
    sock = socket.create_connection((host, int(port)), timeout=15)
    sock.settimeout(None)
    return sock


def main():
    path = os.environ.get("CENTRALIZER_TARGET", ".")
    entry = os.environ.get("CENTRALIZER_ENTRY", "")
    transport = os.environ.get("CENTRALIZER_TRANSPORT", "stdio")
    sock = None
    if transport == "tcp":
        addr = os.environ.get("CENTRALIZER_ADDR", "")
        if not addr:
            sys.stderr.write("CENTRALIZER_ADDR required for tcp\n")
            return 1
        sock = connect_tcp(addr)
    io = IO(sock)
    try:
        mod = load_target(path, entry)
        fns = public_functions(mod)
    except Exception as exc:
        io.send(
            {
                "v": 1,
                "id": "0",
                "type": "ERROR",
                "payload": {"code": "adapter", "message": str(exc)},
            }
        )
        return 1

    handles = {}
    state = {"hid": 0}
    try:
        for msg in io.messages():
            mid = msg.get("id", "")
            try:
                if handle(io, mod, fns, handles, state, msg):
                    return 0
            except Exception as exc:
                io.send(
                    {
                        "v": 1,
                        "id": mid,
                        "type": "ERROR",
                        "payload": {"code": "adapter", "message": str(exc), "retry": False},
                    }
                )
                if os.environ.get("CENTRALIZER_DEBUG"):
                    traceback.print_exc(file=sys.stderr)
    finally:
        if sock is not None:
            sock.close()
    return 0


if __name__ == "__main__":
    sys.exit(main())
