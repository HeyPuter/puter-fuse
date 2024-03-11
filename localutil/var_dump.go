/*
 * Copyright (C) 2024  Puter Technologies Inc.
 *
 * This file is part of puter-fuse.
 *
 * puter-fuse is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
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
