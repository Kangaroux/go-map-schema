package compare

import (
	"fmt"
	"reflect"
)

type MyStruct struct {
	S string
	N int
}

type CompareResults struct {
	dst           interface{}
	src           map[string]interface{}
	MissingFields []string
}

func (r *CompareResults) Dst() reflect.Type {
	return reflect.ValueOf(r.dst).Elem().Type()
}

func (r *CompareResults) String() string {
	return fmt.Sprintf("CompareResults<missing=%v, dst=%s, src=%v>", r.MissingFields, r.Dst().Name(), r.src)
}

func Compare(dst interface{}, src map[string]interface{}) *CompareResults {
	v := reflect.ValueOf(dst)
	t := v.Elem().Type()
	results := &CompareResults{
		dst:           dst,
		src:           src,
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
