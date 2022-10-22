// Copyright 2022 Infrable. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Utility functions.
package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Retrieve an element from a map using a dot-separated key.
//
// NOTE: Assumes map keys do not contain dots.
//
// # Example
//
//	func main() {
//		m := map[string]any{
//			"key1": map[string]any{
//				"key2": "value",
//			},
//		}
//		fmt.Printf("%v\n", GetElementFromKey("key1.key2", m))
//	}
//
//	Output:
//	value
//
// See: https://go.dev/play/p/-4AcBjUnPdh
func GetElementFromKey(k string, m map[string]any) any {
	ks := strings.Split(k, ".")
	elem, _ := m[ks[0]]
	if len(ks) == 1 {
		return elem
	} else {
		switch elem.(type) {
		case map[string]any:
			// Remove first key, then recurse.
			k = strings.Join(ks[1:], ".")
			return GetElementFromKey(k, elem.(map[string]any))
		default:
			return nil
		}
	}
}

// A non-exhaustive union of basic Go types.
type BasicTypes interface {
	bool | string | int | float64
}

// Retrieve an element from a map of interfaces. If the interface is nil,
// return the zero value of the type argument.
//
// # Example
//
//	func main() {
//		m := map[string]any{
//			"key1": map[string]any{
//				"key2": "value",
//			},
//		}
//		fmt.Printf("%v\n", GetElementOrZeroValue[string]("x", m))
//	}
//
//	Output:
//	""
//
// See: https://go.dev/play/p/Uee1TmAuc32
func GetElementOrZeroValue[T BasicTypes](k string, m map[string]any) T {
	elem := m[k]
	if elem == nil {
		// Variables declared without an explicit initial value are given their
		// zero value.
		var v T
		return v
	}
	return elem.(T)
}

// Sort elements in a slice and return concatenation of elements.
//
// TODO: Convert all elements in slice to string type. Note, this is an
// expensive operation, but would account for lists of variable types (ex.
// ["1", 2, 3.0]).
//
// # Example
//
//	func main() {
//	  x := []any{"a", "c", "b"}
//		fmt.Printf("%v\n", SortAndJoinSlice(x, ","))
//	}
//
//	Output:
//	a,b,c
//
// See: https://go.dev/play/p/1vIBE_t4EJD
func SortAndJoinSlice(x []any, sep string) string {
	elems := []string{}
	switch x[0].(type) {
	case string:
		sort.Slice(x, func(i, j int) bool {
			return x[i].(string) < x[j].(string)
		})
	case int:
		sort.Slice(x, func(i, j int) bool {
			return x[i].(int) < x[j].(int)
		})
	case float64:
		sort.Slice(x, func(i, j int) bool {
			return x[i].(float64) < x[j].(float64)
		})
	}
	for _, v := range x {
		elems = append(elems, fmt.Sprint(v))
	}
	return strings.Join(elems, sep)
}

// Sort elements in a map by key and return concatenation of elements.
//
// # Example
//
//	func main() {
//	  m := map[string]any{"a": "1", "c": "3", "b": "2"}
//		fmt.Printf("%v\n", SortAndJoinMap(m, ","))
//	}
//
//	Output:
//	a:1,b:2,c:3
//
// See: https://go.dev/play/p/GgAGQCUWCOf
func SortAndJoinMap(m map[string]any, sep string) string {
	elems := []string{}
	ks := make([]string, 0)
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		switch m[k].(type) {
		case string:
			elems = append(elems, fmt.Sprintf("%s:%s", k, m[k]))
		case int:
			elems = append(elems, fmt.Sprintf("%s:%d", k, m[k]))
		case float64:
			// The special precision -1 uses the smallest number of digits necessary
			// such that ParseFloat will return f exactly.
			s := strconv.FormatFloat(m[k].(float64), 'f', -1, 64)
			elems = append(elems, fmt.Sprintf("%s:%s", k, s))
		}
	}
	return strings.Join(elems, sep)
}
