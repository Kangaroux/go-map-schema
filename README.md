# go-map-schema

[![Go Reference](https://img.shields.io/badge/Go-Docs-blue?style=for-the-badge)](https://pkg.go.dev/github.com/Kangaroux/go-map-schema) ![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/Kangaroux/go-map-schema?style=for-the-badge&label=Latest&color=green)

## Table of Contents

- [Overview](#overview)
- [Use Case](#use-case)
- [Do I Really Need This?](#do-i-really-need-this)
- [Examples](#examples)
    - [Usage](#usage)
    - [Full Code](#full-code)
    - [Output](#output)
- [Universal Type Names](#universal-type-names)

## Overview

`go-map-schema` is a tiny library that's useful for comparing a map (usually from JSON) to a struct, and finding any fields that are missing or that have incompatible types.

## Use Case
The most common usage would be for an API that accepts JSON from clients.

Before we can fulfill the request we need to know if the JSON matches what we expect. By verifying the JSON before we even try to `json.Unmarshal` it into a struct, we can be sure the JSON will be safely converted with no loss of data or swallowed errors.

As a result, you end up with an API that has

1. strict type checking, and
2. can give the client helpful error messages when the request is invalid

## Do I Really Need This?

For something like an API, it's worth asking who is going to be the client that is using the API. Is this an API for your single page app, or are you providing a service and it will be used by other developers?

If the API is "internal" and only used within the context of your site, it's probably not necessary. But if someone is trying to use your API, it can be very useful. API documentation is rarely perfect, and giving the client a readable error can help them debug when they get a `400`.

# Examples

Read below for a quick look at how you can use `go-map-schema`.

For more examples, check out the [examples/](examples/) directory.

## Usage

Suppose we have a `Person` model that contains a nested `Address`

```go
type Address struct {
    Country     string `json:"country"`
    City        string `json:"city"`
    AddressLine string `json:"address_line"`
}

type Person struct {
    FirstName string  `json:"first_name"`
    LastName  string  `json:"last_name"`
    Age       int     `json:"age"`
    Address   Address `json:"address"`
}
```

and a client comes along and makes a `POST` request with this JSON.

```json
{
    "first_name": "Jessie",
    "age": "26",
    "address": {
      "country": "US",
      "city": null
    }
}
```

We can unmarshal the JSON into a map to make it easier to work with, and then compare it with the `Person` model.

```go
src := make(map[string]interface{})
json.Unmarshal(payload, &src)

dst := Person{}
results, err := schema.CompareMapToStruct(&dst, src, nil)
```

After comparing we now have a `CompareResults` instance stored in `results`.

```go
type CompareResults struct {
	MismatchedFields []FieldMismatch
	MissingFields    []FieldMissing
}
```

With this, we can quickly see which fields have mismatched types, as well as any fields that are in the `Person` struct but not the JSON.

Check out [examples/type-errors](examples/type-errors) for the complete example.

# Universal Type Names

By default, `CompareMapToStruct` will use the `DetailedTypeName` func when reporting a type mismatch. The detailed type name includes some extra information that you may not want a client to see.

For example, trying to convert a `float64` into an `int32` will report a mismatch between `float64` and `int32`.

If you don't want to include bit size, you can use `SimpleTypeName` which converts `floatX -> float`, `intX -> int`, `uintX -> uint`.

```go
opts := &schema.CompareOpts{
    TypeNameFunc: schema.SimpleTypeName,
}

schema.CompareMapToStruct(dst, src, opts)
```
