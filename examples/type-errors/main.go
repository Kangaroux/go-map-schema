package main

import (
	"encoding/json"
	"fmt"

	schema "github.com/Kangaroux/go-map-schema"
)

// Address is the model we are using as nested for Person.
type Address struct {
	Country     string `json:"country"`
	City        string `json:"city"`
	AddressLine string `json:"address_line"`
}

// Person is the model we are using.
type Person struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Age       int     `json:"age"`
	Address   Address `json:"address"`
}

// Response is used to generate a JSON response.
type Response struct {
	Errors error `json:"errors,omitempty"`
	OK     bool  `json:"ok"`
}

func main() {
	input := `{
		"first_name": "Jessie",
		"age": "26",
		"address": {
			"country": "US",
			"city": null
		}
	}`
	m := make(map[string]interface{})
	p := Person{}

	// Load the JSON into a map.
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		panic(err)
	}

	// Check the types.
	r, err := schema.CompareMapToStruct(&p, m, nil)

	if err != nil {
		panic(err)
	}

	typeErrors := r.Errors()
	resp := &Response{
		OK:     typeErrors == nil,
		Errors: typeErrors,
	}
	respJson, _ := json.MarshalIndent(resp, "", "    ")

	fmt.Println("missing fields:   ", r.MissingFields)
	fmt.Println("mismatched fields:", r.MismatchedFields)
	fmt.Println(string(respJson))
}
