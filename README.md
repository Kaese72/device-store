# Huemie Device store

This repository contains all code related to the huemie device store

## Filtering

Endpoints that support filtering (currently devices, groups, and device attribute
audits) accept a single `filters` query parameter containing a JSON array of
filter objects:

```
GET /device-store/v0/devices?filters=[{"key":"bridge-identifier","op":"eq","value":"abc-123"}]
```

Each filter object has three fields:

- `key` — which field (or attribute, see below) to filter on.
- `op` — the comparison operator to apply.
- `value` — the value to compare against, always passed as a string.

Every `key` only supports a specific, explicitly allow-listed set of `op`
values. Requesting an unknown key, or an operator that isn't supported for a
given key, results in an HTTP 400 error describing the problem. Values are
validated against the expected type of the field they're compared against
(for example, an `id` filter must be numeric) — there is no implicit type
coercion; a type mismatch is a 400 error, not a best-effort conversion.

The `id` field additionally supports an `in` operator for matching multiple
IDs at once, taking a comma-separated list of integers as its value, e.g.
`{"key":"id","op":"in","value":"1,2,3"}`.

### Filtering on device attributes

Devices carry a dynamic, per-device set of attributes (e.g. `active`,
`color-ct`) that isn't known ahead of time and isn't fixed across devices —
different devices can expose different attributes, and an attribute's value
is stored as exactly one of three types: boolean, numeric, or text.

To filter on an attribute, use a `key` of the form `attribute.<name>`, where
`<name>` is the attribute's name (e.g. `attribute.active`,
`attribute.color-ct`). Because the attribute's type isn't implied by its
name alone, the type is encoded as a prefix on the `op` instead of on the
key — this keeps the attribute name free-form (it may itself contain dots
or other characters) since only the key's first `.` is treated as the
`attribute.` marker.

Supported type-prefixed operators:

- `bool-eq` — boolean equality. Valid `value`s are `true` and `false`.
- `numeric-eq` — exact numeric equality.
- `numeric-aeq` — approximate numeric equality, for tolerating floating
  point imprecision.
- `numeric-lt` / `numeric-gt` — numeric less-than / greater-than.
- `text-eq` — exact text equality.

Requesting a type/comparison combination that doesn't exist (for example
`bool-lt`, which isn't meaningful) is a 400 error, the same as requesting an
unsupported operator on any other field.

Example:

```
GET /device-store/v0/devices?filters=[{"key":"attribute.active","op":"bool-eq","value":"true"},{"key":"attribute.color-ct","op":"numeric-aeq","value":"0.123"}]
```

Matching semantics for attribute filters:

- If a device doesn't have the named attribute at all, it does not match.
- If a device has the named attribute, but its value isn't stored as the
  type implied by the operator prefix (e.g. filtering with `numeric-eq` on
  an attribute whose value is text), it does not match — there is no
  cross-type coercion.

### Attribute statistics

`GET /device-store/v0/attributes/statistics` returns, for each attribute
name in use across all devices, how many devices store that attribute's
value as each of the three possible types:

```
GET /device-store/v0/attributes/statistics
[
  {"name": "active", "n-boolean": 12, "n-text": 0, "n-numeric": 0},
  {"name": "color-ct", "n-boolean": 0, "n-text": 0, "n-numeric": 8}
]
```

The optional `names` query parameter takes a comma-separated list of
attribute names to restrict the results to, e.g.
`?names=active,color-ct`. Unrecognized names simply produce no matching
entry — this endpoint doesn't validate `names` against a fixed set, since
attribute names are dynamic.

