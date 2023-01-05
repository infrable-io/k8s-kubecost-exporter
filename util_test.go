// Copyright 2023 Infrable. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetElementFromKey(t *testing.T) {
	cases := []struct {
		k    string
		m    map[string]any
		want any
	}{
		{
			k:    "key",
			m:    map[string]any{"key": "value"},
			want: "value",
		},
		{
			k:    "x",
			m:    map[string]any{"key": "value"},
			want: nil,
		},
		// string
		{
			k:    "key1.key2",
			m:    map[string]any{"key1": map[string]any{"key2": "value"}},
			want: "value",
		},
		// int
		{
			k:    "key1.key2",
			m:    map[string]any{"key1": map[string]any{"key2": 1}},
			want: 1,
		},
		// float64
		{
			k:    "key1.key2",
			m:    map[string]any{"key1": map[string]any{"key2": 13.37}},
			want: 13.37,
		},
		{
			k:    "key1.key2.key3",
			m:    map[string]any{"key1": map[string]any{"key2": "value"}},
			want: nil,
		},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			ret := GetElementFromKey(tc.k, tc.m)
			assert.Equal(t, tc.want, ret)
		})
	}
}

type Testable interface {
	Test(t *testing.T)
}

type GetElementOrZeroValueTestCase[T BasicTypes] struct {
	k    string
	m    map[string]any
	want T
}

func (tc GetElementOrZeroValueTestCase[T]) Test(t *testing.T) {
	t.Run("", func(t *testing.T) {
		ret := GetElementOrZeroValue[T](tc.k, tc.m)
		assert.Equal(t, tc.want, ret)
	})
}

func TestGetElementOrZeroValue(t *testing.T) {
	cases := []Testable{
		GetElementOrZeroValueTestCase[bool]{
			k:    "key",
			m:    map[string]any{},
			want: false,
		},
		GetElementOrZeroValueTestCase[string]{
			k:    "key",
			m:    map[string]any{},
			want: "",
		},
		GetElementOrZeroValueTestCase[int]{
			k:    "key",
			m:    map[string]any{},
			want: 0,
		},
		GetElementOrZeroValueTestCase[float64]{
			k:    "key",
			m:    map[string]any{},
			want: 0,
		},
	}
	for _, tc := range cases {
		tc.Test(t)
	}
}

func TestSortAndJoinSlice(t *testing.T) {
	cases := []struct {
		x    []any
		sep  string
		want string
	}{
		// string
		{
			x:    []any{"a", "c", "b"},
			sep:  ",",
			want: "a,b,c",
		},
		// int
		{
			x:    []any{1, 3, 2},
			sep:  ",",
			want: "1,2,3",
		},
		// float64
		{
			x:    []any{1.1, 3.3, 2.2},
			sep:  ",",
			want: "1.1,2.2,3.3",
		},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			ret := SortAndJoinSlice(tc.x, tc.sep)
			assert.Equal(t, tc.want, ret)
		})
	}
}

func TestSortAndJoinMap(t *testing.T) {
	cases := []struct {
		m    map[string]any
		sep  string
		want string
	}{
		// string
		{
			m:    map[string]any{"a": "1", "c": "3", "b": "2"},
			sep:  ",",
			want: "a:1,b:2,c:3",
		},
		// int
		{
			m:    map[string]any{"a": 1, "c": 3, "b": 2},
			sep:  ",",
			want: "a:1,b:2,c:3",
		},
		// float64
		{
			m:    map[string]any{"a": 1.1, "c": 3.3, "b": 2.2},
			sep:  ",",
			want: "a:1.1,b:2.2,c:3.3",
		},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			ret := SortAndJoinMap(tc.m, tc.sep)
			assert.Equal(t, tc.want, ret)
		})
	}
}
