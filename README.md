[![Go Reference](https://img.shields.io/badge/Go-Docs-blue?style=for-the-badge)](https://pkg.go.dev/github.com/Kangaroux/go-map-schema) ![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/Kangaroux/go-map-schema?style=for-the-badge&label=Latest&color=green)

`go-map-schema` is a tiny library that's useful for comparing a map (usually from JSON) to a struct, and finding any fields that are missing or that have incompatible types.

## Use Case
The most common usage would be for an API that accepts JSON from clients.

Before we can fulfill the request we need to know if the JSON matches what we expect. By verifying the JSON before we even try to `json.Unmarshal` it into a struct, we can be sure the JSON will be safely converted with no loss of data or swallowed errors.

As a result, you end up with an API that has

1. strict type checking, and
2. can give the client helpful error messages when the request is invalid

## Example

Suppose we have a `Person` model

```go
type Person struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Age       int    `json:"age"`
}
```

and a client comes along and makes a `POST` request with this JSON.

```json
{
    "first_name": "Jessie",
    "age": "26"
}
```

We can unmarshal the JSON into a map to make it easier to work with, and then compare it with the `Person` model.

```go
src := make(map[string]interface{})
json.Unmarshal(payload, &src)

dst := Person{}
results, err := schema.CompareMapToStruct(&dst, src)
```

After comparing we now have a `CompareResults` instance stored in `results`.

```go
type CompareResults struct {
	MismatchedFields []FieldMismatch
	MissingFields    []string
}
```

With this, we can quickly see which fields have mismatched types, as well as any fields that are in the `Person` struct but not the JSON.

### Full Example
```go
package main

import (
    "encoding/json"
    "fmt"

    schema "github.com/Kangaroux/go-map-schema"
)

type Person struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Age       int    `json:"age"`
}

func main() {
    // The JSON payload (src).
    payload := `{
        "first_name": "Jessie",
        "age": "26"
    }`
    m := make(map[string]interface{})
    json.Unmarshal([]byte(payload), &m)

    // The struct we want to load the payload into (dst).
    p := Person{}
    results, _ := schema.CompareMapToStruct(&p, m)

    fmt.Println("missing fields:   ", results.MissingFields)
    fmt.Println("mismatched fields:", results.MismatchedFields)
}
```

### Output

```
missing fields:    [last_name]
mismatched fields: [expected "age" to be a int but it's a string]
```

`FieldMismatch.String()` returns a simple user friendly message describing what the issue is. You can of course use your own custom message instead.