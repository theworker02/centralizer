# Centralizer Protocol 1.x

Handshake negotiates `protocol: "1.0"`. Major version 1 is required. Unknown message types return `ERROR`.

## Transports

| Transport | Framing |
| --- | --- |
| stdio | NDJSON, one message per line |
| TCP / Unix | 4-byte big-endian length + JSON body |
| memory (tests) | same as TCP |

Maximum frame size: 16 MiB (`protocol.MaxFrameBytes`).

## Envelope

```json
{"v":1,"id":"corr-id","type":"CALL","payload":{}}
```

`id` is a correlation identifier. Responses reuse it.

## Messages

| Type | Direction | Purpose |
| --- | --- | --- |
| HELLO | both | version and feature negotiation |
| CAPABILITIES | peer → host | optional capability list |
| DESCRIBE / DESCRIBE_OK | both | schema YAML |
| CALL | host → peer | function or method |
| RESULT | peer → host | CIR value |
| ERROR | either | structured error |
| STREAM_OPEN / STREAM_DATA / STREAM_CLOSE | both | streams |
| HANDLE_CREATE / HANDLE_RELEASE | both | opaque objects |
| GET / SET | host → peer | properties |
| HEARTBEAT | both | liveness |
| CANCEL | host → peer | abort `id` |
| SHUTDOWN | host → peer | graceful exit |
| OK | peer → host | empty success |

## HELLO

```json
{"v":1,"id":"1","type":"HELLO","payload":{"protocol":"1.0","name":"centralizer","features":["call","stream","handles"]}}
```

## CALL

Arguments are CIR wire values, not raw JSON types.

```json
{"v":1,"id":"2","type":"CALL","payload":{"function":"calculate","args":{"value":{"k":"int","i":42}}}}
```

## ERROR

```json
{"v":1,"id":"2","type":"ERROR","payload":{"code":"schema","message":"unknown function","retry":false}}
```

Codes: `schema`, `conversion`, `handle`, `timeout`, `cancel`, `adapter`, `protocol`, `frame`.

## Streaming

`STREAM_OPEN` names a function or event. The peer replies with `STREAM_OPEN` or `OK`, then sends one or more `STREAM_DATA` frames using the stream id as `id`, and finishes with `STREAM_CLOSE`.

Python generators and Node iterables are streamed by the generated shims.

## Compatibility

A 1.x peer may ignore unknown optional fields. A 2.x peer must not claim 1.x compatibility without a documented translation. Never assume both sides run identical binaries.
