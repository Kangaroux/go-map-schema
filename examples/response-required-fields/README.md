This example shows how you might generate a JSON response if the client's request omits some required fields.

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
    "age": 26
}
```

## Output
```json
{
    "errors": {
        "last_name": "this field is required"
    },
    "ok": false
}
```