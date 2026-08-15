# CIR — Centralizer Intermediate Representation

CIR is the semantic boundary between supported languages. Adapters convert native values into CIR and back. They do not invent a second type system.

## Kinds

| Kind | Storage | Notes |
| --- | --- | --- |
| null | — | |
| boolean | `num` 0/1 | |
| int | `num` as int64 bits | overflow checked on convert |
| uint | `num` | |
| float | `num` as float64 bits | non-finite rejected by schema validate |
| decimal | `str` | exact decimal text |
| string | `str` | |
| bytes | `raw` | |
| array / tuple | `items` | |
| map / struct | `keys` + `items` | |
| enum | `str` + discriminant | |
| union | tag + inner | |
| optional | flag + inner | |
| result | ok flag + inner | |
| error | code + message | |
| timestamp | unix nanos | UTC |
| duration | nanos | |
| UUID | 16 bytes | |
| stream | id | |
| handle | id | never a foreign pointer |
| opaque | tag + payload | language-specific |

Optional `Meta` carries type name, language, and schema id.

## Wire JSON

```json
{"k":"int","i":42}
{"k":"map","m":[{"k":"value","v":{"k":"int","i":42}}]}
```

See `pkg/cir.Wire`. Do not decode CIR by guessing JSON types.

## Adapter mapping

| Language | Native → CIR |
| --- | --- |
| Go | `cir.From` / `Value.Native` |
| Python shim | `int`→int, `float`→float, `str`→string, `bytes`→bytes, `dict`→map, `list`→array |
| Node shim | `number` integer→int else float; `Buffer`→bytes; object→map |
| Rust example | integer/float fields in CALL args → float result |

Document new mappings in the adapter README when you add them.

## Schema

Schemas sit above CIR and describe functions, objects, events, and streams. Inferred schemas do not enforce argument types; explicit schemas do.
