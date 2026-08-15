#!/usr/bin/env node
// Centralizer protocol shim v2. Minimal logic. Do not duplicate runtime internals.
"use strict";

const fs = require("fs");
const net = require("net");
const path = require("path");
const readline = require("readline");

function decode(w) {
  if (!w) return null;
  switch (w.k) {
    case "null":
      return null;
    case "boolean":
      return !!w.b;
    case "int":
      return Number(w.i || 0);
    case "uint":
      return Number(w.u || 0);
    case "float":
      return Number(w.f || 0);
    case "string":
    case "decimal":
      return w.s || "";
    case "bytes":
      return Buffer.from(w.x || "", "base64");
    case "array":
    case "tuple":
      return (w.a || []).map(decode);
    case "map":
    case "struct": {
      const o = {};
      for (const e of w.m || []) o[e.k] = decode(e.v);
      return o;
    }
    case "optional":
      return w.p && w.a && w.a[0] ? decode(w.a[0]) : null;
    default:
      return w.s;
  }
}

function encode(v) {
  if (v === null || v === undefined) return { k: "null" };
  if (typeof v === "boolean") return { k: "boolean", b: v };
  if (typeof v === "number" && Number.isInteger(v)) return { k: "int", i: v };
  if (typeof v === "number") return { k: "float", f: v };
  if (typeof v === "string") return { k: "string", s: v };
  if (Buffer.isBuffer(v)) return { k: "bytes", x: v.toString("base64") };
  if (Array.isArray(v)) return { k: "array", a: v.map(encode) };
  if (typeof v === "object") {
    return {
      k: "map",
      m: Object.keys(v).map((key) => ({ k: key, v: encode(v[key]) })),
    };
  }
  return { k: "string", s: String(v) };
}

function loadTarget(root, entry) {
  const abs = path.resolve(root);
  if (entry) {
    return require(path.isAbsolute(entry) ? entry : path.join(abs, entry));
  }
  const pkgPath = path.join(abs, "package.json");
  if (fs.existsSync(pkgPath)) {
    const pkg = JSON.parse(fs.readFileSync(pkgPath, "utf8"));
    const main = pkg.main || "index.js";
    return require(path.join(abs, main));
  }
  const index = path.join(abs, "index.js");
  if (fs.existsSync(index)) return require(index);
  const files = fs.readdirSync(abs).filter((f) => f.endsWith(".js"));
  if (files.length === 1) return require(path.join(abs, files[0]));
  throw new Error("unable to infer node entry; set entry in the manifest");
}

function publicFunctions(mod) {
  const out = {};
  if (typeof mod === "function") {
    out.default = mod;
    return out;
  }
  for (const [k, v] of Object.entries(mod || {})) {
    if (k.startsWith("_")) continue;
    if (typeof v === "function") out[k] = v;
  }
  return out;
}

function describe(fns) {
  const lines = ["service: node_target", "inferred: true", "functions:"];
  for (const name of Object.keys(fns)) {
    lines.push("  " + name + ":");
    lines.push("    args: {}");
    lines.push("    returns:");
    lines.push("      type: string");
  }
  return lines.join("\n") + "\n";
}

function isStreamable(v) {
  if (v == null) return false;
  if (typeof v === "string" || Buffer.isBuffer(v)) return false;
  if (typeof v[Symbol.asyncIterator] === "function") return true;
  if (typeof v[Symbol.iterator] === "function" && typeof v !== "string") return true;
  return false;
}

function callFn(fn, args) {
  const values = Object.values(args);
  return values.length === 1 ? fn(values[0]) : fn(args);
}

const target = process.env.CENTRALIZER_TARGET || ".";
const entry = process.env.CENTRALIZER_ENTRY || "";
const transport = process.env.CENTRALIZER_TRANSPORT || "stdio";
let fns = {};
try {
  fns = publicFunctions(loadTarget(target, entry));
} catch (err) {
  writeStdio({ v: 1, id: "0", type: "ERROR", payload: { code: "adapter", message: String(err) } });
  process.exit(1);
}

let send = writeStdio;

function writeStdio(msg) {
  process.stdout.write(JSON.stringify(msg) + "\n");
}

function writeFrame(sock, msg) {
  const body = Buffer.from(JSON.stringify(msg), "utf8");
  const hdr = Buffer.alloc(4);
  hdr.writeUInt32BE(body.length, 0);
  sock.write(Buffer.concat([hdr, body]));
}

async function emitStream(sid, result) {
  const out = await Promise.resolve(result);
  if (out && typeof out[Symbol.asyncIterator] === "function") {
    for await (const item of out) {
      send({ v: 1, id: sid, type: "STREAM_DATA", payload: { stream: sid, value: encode(item) } });
    }
  } else if (isStreamable(out)) {
    for (const item of out) {
      send({ v: 1, id: sid, type: "STREAM_DATA", payload: { stream: sid, value: encode(item) } });
    }
  } else {
    send({ v: 1, id: sid, type: "STREAM_DATA", payload: { stream: sid, value: encode(out) } });
  }
  send({ v: 1, id: sid, type: "STREAM_CLOSE", payload: { stream: sid } });
}

function onMessage(msg) {
  const mid = msg.id || "";
  const typ = msg.type;
  const payload = msg.payload || {};
  try {
    if (typ === "HELLO") {
      send({
        v: 1,
        id: mid,
        type: "HELLO",
        payload: {
          protocol: "1.0",
          name: "centralizer-node",
          features: ["call", "describe", "stream"],
          runtime: "Node.js",
          version: process.versions.node,
        },
      });
      return;
    }
    if (typ === "DESCRIBE") {
      send({ v: 1, id: mid, type: "DESCRIBE_OK", payload: { schema: describe(fns) } });
      return;
    }
    if (typ === "CALL") {
      const name = payload.function;
      if (!fns[name]) throw new Error("unknown function " + name);
      const args = {};
      for (const [k, v] of Object.entries(payload.args || {})) args[k] = decode(v);
      Promise.resolve(callFn(fns[name], args)).then(
        (out) => send({ v: 1, id: mid, type: "RESULT", payload: { value: encode(out) } }),
        (err) => send({ v: 1, id: mid, type: "ERROR", payload: { code: "adapter", message: String(err) } })
      );
      return;
    }
    if (typ === "STREAM_OPEN") {
      const name = payload.name;
      const sid = payload.stream || mid;
      if (!fns[name]) throw new Error("unknown function " + name);
      const args = {};
      for (const [k, v] of Object.entries(payload.args || {})) args[k] = decode(v);
      send({ v: 1, id: mid, type: "STREAM_OPEN", payload: { stream: sid } });
      Promise.resolve(callFn(fns[name], args)).then(
        (out) => emitStream(sid, out),
        (err) => send({ v: 1, id: mid, type: "ERROR", payload: { code: "adapter", message: String(err) } })
      );
      return;
    }
    if (typ === "HEARTBEAT" || typ === "CANCEL") {
      send({ v: 1, id: mid, type: "OK" });
      return;
    }
    if (typ === "SHUTDOWN") {
      send({ v: 1, id: mid, type: "OK" });
      process.exit(0);
    }
    send({ v: 1, id: mid, type: "ERROR", payload: { code: "protocol", message: "unsupported " + typ } });
  } catch (err) {
    send({ v: 1, id: mid, type: "ERROR", payload: { code: "adapter", message: String(err) } });
  }
}

function serveStdio() {
  const rl = readline.createInterface({ input: process.stdin, crlfDelay: Infinity });
  rl.on("line", (line) => {
    if (!line.trim()) return;
    let msg;
    try {
      msg = JSON.parse(line);
    } catch (err) {
      send({ v: 1, id: "0", type: "ERROR", payload: { code: "frame", message: String(err) } });
      return;
    }
    onMessage(msg);
  });
}

function serveTCP(addr) {
  const last = addr.lastIndexOf(":");
  const host = addr.slice(0, last);
  const port = Number(addr.slice(last + 1));
  const sock = net.connect({ host, port }, () => {});
  send = (msg) => writeFrame(sock, msg);
  let buf = Buffer.alloc(0);
  sock.on("data", (chunk) => {
    buf = Buffer.concat([buf, chunk]);
    while (buf.length >= 4) {
      const n = buf.readUInt32BE(0);
      if (buf.length < 4 + n) break;
      const body = buf.subarray(4, 4 + n);
      buf = buf.subarray(4 + n);
      try {
        onMessage(JSON.parse(body.toString("utf8")));
      } catch (err) {
        send({ v: 1, id: "0", type: "ERROR", payload: { code: "frame", message: String(err) } });
      }
    }
  });
  sock.on("error", (err) => {
    process.stderr.write(String(err) + "\n");
    process.exit(1);
  });
}

if (transport === "tcp") {
  const addr = process.env.CENTRALIZER_ADDR || "";
  if (!addr) {
    process.stderr.write("CENTRALIZER_ADDR required for tcp\n");
    process.exit(1);
  }
  serveTCP(addr);
} else {
  serveStdio();
}
