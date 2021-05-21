This example shows how you might generate a JSON response if the client's request has some type errors.

## Model
```go
type Person struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
}
```

## Input
```json
{
    "first_name": "Jessie",
    "age": "26"
}
```

## Output
```json
{
    "errors": {
        "age": "expected an int but it's a string"
    },
    "ok": false
}
```