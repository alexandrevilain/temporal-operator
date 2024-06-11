// Licensed to Alexandre VILAIN under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Alexandre VILAIN licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package enumerable

// Map applies a given function to each element of an input slice and returns a new slice
// containing the results. It allows for transforming the elements of a collection into another form.
//
// The function is generic and can handle slices of any type.
//
// Parameters:
//   - items ([]T): The input slice of any type T.
//   - f (func(T) TR): The function to apply to each element of the input slice. The function
//     takes an element of type T and returns a value of type TR.
//
// Returns:
//   - []TR: A new slice containing the results of applying the function f to each element of the input slice.
//
// Example:
//
//	input := []int{1, 2, 3, 4}
//	result := Map(input, func(v int) string {
//	    return fmt.Sprintf("Number: %d", v)
//	})
//	fmt.Println(result) // Output: [Number: 1 Number: 2 Number: 3 Number: 4]
func Map[T any, TR any](items []T, f func(T) TR) []TR {
	if len(items) == 0 {
		return []TR{}
	}

	re := make([]TR, len(items))

	for i, v := range items {
		re[i] = f(v)
	}

	return re
}
