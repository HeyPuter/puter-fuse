package localutil

import (
	"fmt"
	"reflect"
)

func Printvar(value interface{}, tag string) {
	if value == nil {
		fmt.Printf("PRINTVAR:%s->nil\n", tag)
		return
	}

	val := reflect.ValueOf(value)

	// Handle pointers by getting the value they point to
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Print type
	fmt.Printf("PRINTVAR:%s->%s\n", tag, val.Type())

	// Handling based on kind
	switch val.Kind() {
	case reflect.Struct:
		// Iterate over struct fields
		fmt.Println("Content: {")
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			fmt.Printf("  %s: %v\n", field.Name, val.Field(i).Interface())
		}
		fmt.Println("}")
	case reflect.Slice, reflect.Array:
		// Handle slices and arrays
		fmt.Print("Content: [")
		for i := 0; i < val.Len(); i++ {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(val.Index(i).Interface())
		}
		fmt.Println("]")
	case reflect.Map:
		// Handle maps
		fmt.Println("Content: {")
		for _, key := range val.MapKeys() {
			fmt.Printf("  %v: %v\n", key.Interface(), val.MapIndex(key).Interface())
		}
		fmt.Println("}")
	default:
		// Fallback for basic types
		fmt.Printf("Content: %v\n", val.Interface())
	}
}
