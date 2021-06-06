package main

import (
	"encoding/json"
	"fmt"

	schema "github.com/Kangaroux/go-map-schema"
)

// Person is the model we are using.
type Person struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
}

// Response is used to generate a JSON response.
type Response struct {
	Errors interface{} `json:"errors,omitempty"`
	OK     bool        `json:"ok"`
}

func main() {
	input := `{
		"first_name": "Jessie",
		"age": 26
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

	resp := &Response{
		OK: true,
	}

	// Notify the client of any errors.
	if len(r.MissingFields) > 0 {
		m := make(map[string]string)

		for _, f := range r.MissingFields {
			m[f.String()] = "this field is required"
		}

		resp.OK = false
		resp.Errors = m
	}

	respJson, _ := json.Marshal(resp)

	fmt.Println(string(respJson))
}
