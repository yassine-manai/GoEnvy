# astroenv — Library Specification

> **Purpose:** Language-agnostic reference for re-implementing the astroenv env-to-struct loader pattern.
> **Go implementation:** https://github.com/yassine-manai/astroenv

---

## Core Concept

Given a struct/object with tagged fields and an optional `.env` file, populate the fields from environment variables. Tags define the mapping key, optional default, and optional time format.

---

## Tag System

### Primary Tag: `env`

| Format | Behavior |
|--------|----------|
| `env:"KEY"` | **Required.** Look up `KEY` in env vars. Error if not found and no default. |
| `env:"KEY,default"` | **Optional.** Look up `KEY`. If missing, use `default`. Both env value and default are whitespace-trimmed. |

#### Tag Parsing Algorithm

```
Input:  tag string (e.g. "PORT,8080")
Output: key, default_val, has_default

1. Split on first comma (max 2 parts)
2. part[0] = key (trimmed)
3. if part[1] exists → has_default = true, default_val = trimmed(part[1])
4. else → has_default = false, default_val = ""
```

### Companion Tag: `env_format` (for time types only)

Used exclusively with `time.Time` fields. Specifies the expected date/time layout.

```
env_format:"2006-01-02"
```

If absent, the implementation must default to RFC3339 (`"2006-01-02T15:04:05Z07:00"`).

---

## Public API

### `LoadEnvVariable(cfg: object) -> Error`

| Step | Behavior |
|------|----------|
| 1 | Attempt to load `.env` file (godotenv equivalent). Failure is a non-fatal warning. |
| 2 | Validate `cfg` is a non-null pointer/reference to a struct. Return error if not. |
| 3 | Walk all fields. For each field with an `env` tag, resolve and set the value. |
| 4 | Nested structs are recursed into (unless the field is a time type). |
| 5 | Return first error encountered, or nil on success. |

### `MustLoadEnvVariable(cfg: object)`

Same as `LoadEnvVariable` but panics/crashes on error instead of returning it.

---

## Value Resolution Algorithm

```
resolveValue(key, default_val, has_default, field_name) -> (string, error)

1. val = get_env(key)
2. if val is not empty and not whitespace-only:
     return trimmed(val)
3. if has_default is true:
     return trimmed(default_val)
4. return error: missing required env variable
```

**Critical detail:** Step 1 must NOT distinguish between "variable not set" and "variable set to empty string". An empty string after trimming is treated as "not set", falling through to default or error. This means `KEY=` in a `.env` file produces `""` → trimmed to `""` → falls to default/error.

---

## Supported Types

### Type Dispatch Order

The implementation must check **named types** before **kind-based dispatch**:

1. If field type is `time.Duration` → duration parser
2. If field type is `time.Time` → time parser (uses `env_format` tag)
3. Otherwise, dispatch by kind:

### Type Table

| Kind / Type | Input Format | Parsing Method | Whitespace Handling |
|-------------|-------------|----------------|---------------------|
| **String** | Raw text | Identity (no parse) | N/A (raw value already trimmed) |
| **Int** (int, int8, int16, int32, int64) | Decimal digits, optional sign | `parseInt(s, base=10, bits=64)` | Value already trimmed |
| **Bool** | `true`, `false`, `1`, `0`, `t`, `f` | Standard bool parser | Value already trimmed |
| **Float** (float32, float64) | Decimal number | `parseFloat(s, bits=64)` | Value already trimmed |
| **Duration** (time.Duration) | Go-style: `Ns`, `Nms`, `Ns`, `Nm`, `Nh` (e.g. `5s`, `1m30s`, `2h`) | Dedicated duration parser (`time.ParseDuration`) | Value already trimmed |
| **Slice of string** | Comma-separated: `a,b,c` | Split on `,`, trim each element | Per-element trim inside splitter |
| **Slice of int** | Comma-separated: `1,2,3` | Split on `,`, trim, `parseInt` each | Per-element trim inside splitter |
| **Slice of float64** | Comma-separated: `1.5,2.3` | Split on `,`, trim, `parseFloat` each | Per-element trim inside splitter |
| **Map[string]string** | Comma-separated: `k1=v1,k2=v2` | Split on `,`, split each on first `=`, trim key+val | Per-element trim inside splitter |

### Empty Value Behavior

| Type | Empty Raw Value (`""`) |
|------|-----------------------|
| string | Set to `""` |
| int / bool / float / duration | Falls through to default/error (empty string is not a valid number) |
| slice | Initialize empty slice |
| map | Initialize empty map |
| time.Time | Falls through to default/error (empty string is not a valid time) |

---

## Nested Struct Handling

During field iteration:

1. If field is a struct AND field type is NOT a time type → **recurse** into it (do not look for `env` tag at this level)
2. If field is a struct AND field type IS a time type → process as a time field (look for `env` tag)
3. If field is a struct AND field has no `env` tag → skip (fields inside will be checked during recursion)

---

## Error Messages (exact strings)

All error messages use the format prefix `[astroenv] `. Re-implementations should match for parity:

```
[astroenv] Warning: could not load .env file: <underlying error>

[astroenv] expected a pointer to a struct, got <actual type>

[astroenv] missing required env variable "<key>" (for field "<field_name>")

[astroenv] field "<field_name>": unsupported type <reflect.Kind>

[astroenv] field "<field_name>": cannot parse "<raw_val>" as int: <underlying error>

[astroenv] field "<field_name>": cannot parse "<raw_val>" as bool (use true/false/1/0/t/f): <underlying error>

[astroenv] field "<field_name>": cannot parse "<raw_val>" as float: <underlying error>

[astroenv] field "<field_name>": cannot parse "<raw_val>" as duration (use format like 5s, 1m30s, 2h): <underlying error>

[astroenv] field "<field_name>": cannot parse "<raw_val>" as time with format "<format>": <underlying error>

[astroenv] field "<field_name>": unsupported slice element type <element_type>

[astroenv] field "<field_name>": unsupported map type, only map[string]string is supported

[astroenv] failed to load environment: <inner error>   [from MustLoadEnvVariable panic]
```

---

## Internal Architecture

```
LoadEnvVariable(cfg)
  ├── godotenv.Load()           [side effect: load .env file]
  ├── validate pointer to struct
  └── parseStruct(struct_value)
        └── for each field:
              ├── is nested struct (not time) → recurse
              ├── no env tag → skip
              ├── parseTag(tag) → (key, default, has_default)
              ├── resolveValue(key, default, has_default, field_name) → (raw_string, error)
              ├── setField(field, field_type, raw_string)
              │     ├── is durationType → setDurationField
              │     ├── is timeType → setTimeField (reads env_format tag)
              │     └── dispatch by kind:
              │           ├── String → setStringField
              │           ├── Int family → setIntField
              │           ├── Bool → setBoolField
              │           ├── Float family → setFloatField
              │           ├── Slice → setSliceField
              │           └── Map → setMapField
              └── error ∈ {resolve, set} → abort and return
```

---

## Edge Cases Summary

1. **Whitespace in env values**: Both env lookup results and tag default values are trimmed before parsing.
2. **Missing `.env` file**: Non-fatal warning, continues with actual env vars.
3. **Empty env value** (`KEY=`): Treated as unset (falls to default or error).
4. **Empty slice/map input**: Initialized as empty slice/map, not null/nil.
5. **Malformed map entries** (no `=`): Silently skipped.
6. **Unsupported types**: Returns descriptive error naming the type.
7. **Non-struct / non-pointer input**: Returns descriptive error.
8. **Nested time.Time fields**: Not recursed into; processed as leaf values.
9. **Slice of unsupported element type**: Returns descriptive error.
10. **Map of non-string key/value**: Returns descriptive error (only `map[string]string` supported).
