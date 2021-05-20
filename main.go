package main

import (
	"fmt"
	"reflect"
)

type MyStruct struct {
	S string
	N int
}

type CompareResults struct {
	MissingFields []string
}

func (r *CompareResults) String() string {
	return fmt.Sprintf("CompareResults<Missing=%v>", r.MissingFields)
}

func Compare(dst interface{}, src map[string]interface{}) *CompareResults {
	v := reflect.ValueOf(dst)
	t := v.Elem().Type()
	results := &CompareResults{
		MissingFields: []string{},
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if srcVal, ok := src[f.Name]; ok {
			fmt.Printf("payload[%s]: %v\n", f.Name, srcVal)
		} else {
			results.MissingFields = append(results.MissingFields, f.Name)
		}
	}

	return results
}

func main() {
	o := MyStruct{}
	m := make(map[string]interface{})
	m["S"] = "hello"
	m["N"] = 4

	r := Compare(&o, m)
	fmt.Println(r)
}
