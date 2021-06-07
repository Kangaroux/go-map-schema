This example shows how you might generate a JSON response if the client's request has some type errors.

## Model
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

## Input
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

## Output
```
missing fields:    [last_name address.address_line]
mismatched fields: [expected "age" to be an int but it's a string expected "address.city" to be a string but it's null]
```

```json
{
    "errors": {
        "address": {
            "city": "expected a string but it's null"
        },
        "age": "expected an int but it's a string"
    },
    "ok": false
}
```
