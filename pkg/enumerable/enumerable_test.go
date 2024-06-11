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

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestMap(t *testing.T) {
	tests := map[string]struct {
		input    []any
		expected []any
		action   func(any) any
	}{
		"Nil slice": {
			input:    nil,
			expected: []any{},
			action: func(v any) any {
				return v
			},
		},
		"Empty slice": {
			input:    []any{},
			expected: []any{},
			action:   func(v any) any { return v },
		},
		"Single element": {
			input:    []any{2},
			expected: []any{4},
			action:   func(v any) any { return v.(int) * v.(int) },
		},
		"Multiple elements": {
			input:    []any{1, 2, 3, 4, 5},
			expected: []any{1, 4, 9, 16, 25},
			action:   func(v any) any { return v.(int) * v.(int) },
		},
		"Type conversion": {
			input:    []any{1, 2, 3, 4, 5},
			expected: []any{"Number: 1", "Number: 2", "Number: 3", "Number: 4", "Number: 5"},
			action: func(v any) any {
				return fmt.Sprintf("Number: %d", v)
			},
		},
		"Mixed types": {
			input:    []any{1, "hello", true},
			expected: []any{2, "HELLO", false},
			action: func(v any) any {
				switch t := v.(type) {
				case int:
					return t + 1
				case string:
					return strings.ToUpper(t)
				case bool:
					return !t
				default:
					return v
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := Map(test.input, test.action)

			if !reflect.DeepEqual(test.expected, actual) {
				t.Errorf("Select(%v, action) = expected %v, actual %v", test.input, test.expected, actual)
			}
		})
	}
}
