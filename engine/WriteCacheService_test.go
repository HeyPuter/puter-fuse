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
package engine_test

import (
	"testing"

	"github.com/HeyPuter/puter-fuse/engine"
)

func TestWriteMutation(t *testing.T) {
	t.Run("WriteMutation->ApplyToBuffer", func(t *testing.T) {
		t.Run("read offset 2, mutate offset 4", func(t *testing.T) {
			// __ab1234

			mut := &engine.WriteMutation{
				Data:   []byte("cdef"),
				Offset: 4,
			}
			buffer := []byte("ab1234")
			mut.ApplyToBuffer(buffer, int64(2))
			if string(buffer) != "abcdef" {
				t.Errorf("Expected 'abcdef', got '%s'", buffer)
			}
		})

		t.Run("read offset 4, mutate offset 2", func(t *testing.T) {
			// __1234ef

			mut := &engine.WriteMutation{
				Data:   []byte("abcd"),
				Offset: 2,
			}
			buffer := []byte("34ef")
			mut.ApplyToBuffer(buffer, int64(4))
			if string(buffer) != "cdef" {
				t.Errorf("Expected 'cdef', got '%s'", buffer)
			}
		})

		type testCase struct {
			label      string
			readOffset int64
			readLength int64
			mutOffset  int64
			mutLength  int64
			expected   string
		}

		// test source: 0123456789
		// test write:  ABCD
		testCases := map[string]testCase{
			"D-L": {
				label:      "Disjointed (L)",
				readOffset: 0,
				readLength: 4,
				mutOffset:  6,
				mutLength:  4,
				expected:   "0123",
			},
			"D-R": {
				label:      "Disjointed (R)",
				readOffset: 6,
				readLength: 4,
				mutOffset:  0,
				mutLength:  4,
				expected:   "6789",
			},
			"A-L": {
				label:      "Adjacent (L)",
				readOffset: 0,
				readLength: 4,
				mutOffset:  4,
				mutLength:  4,
				expected:   "0123",
			},
			"A-R": {
				label:      "Adjacent (R)",
				readOffset: 4,
				readLength: 4,
				mutOffset:  0,
				mutLength:  4,
				expected:   "4567",
			},
			"O-L": {
				label:      "Overlap (L)",
				readOffset: 0,
				readLength: 4,
				mutOffset:  2,
				mutLength:  4,
				expected:   "01AB",
			},
			"O-R": {
				label:      "Overlap (R)",
				readOffset: 2,
				readLength: 4,
				mutOffset:  0,
				mutLength:  4,
				expected:   "CD45",
			},
			"EQ": {
				label:      "Equal",
				readOffset: 0,
				readLength: 4,
				mutOffset:  0,
				mutLength:  4,
				expected:   "ABCD",
			},
			"C-L": {
				label:      "Contained (L)",
				readOffset: 0,
				readLength: 4,
				mutOffset:  1,
				mutLength:  2,
				expected:   "0AB3",
			},
			"C-R": {
				label:      "Contained (R)",
				readOffset: 1,
				readLength: 2,
				mutOffset:  0,
				mutLength:  4,
				expected:   "BC",
			},
		}

		for label, tc := range testCases {
			t.Run(label, func(t *testing.T) {
				// 0123456789

				mut := &engine.WriteMutation{
					Data:   []byte("ABCD")[0:tc.mutLength],
					Offset: tc.mutOffset,
				}
				buffer := []byte("0123456789")[tc.readOffset : tc.readOffset+tc.readLength]
				mut.ApplyToBuffer(buffer, tc.readOffset)

				if string(buffer) != tc.expected {
					t.Errorf("Expected '%s', got '%s'", tc.expected, buffer)
				}
			})
		}

	})
}
