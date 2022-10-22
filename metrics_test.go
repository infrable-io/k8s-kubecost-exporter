// Copyright 2022 Infrable. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func NewTestConfig(c []byte) *viper.Viper {
	v := viper.New()
	v.SetConfigType("yaml")
	v.ReadConfig(bytes.NewBuffer(c))
	return v
}

func TestNewPrometheusMetrics(t *testing.T) {
	cases := []struct {
		config []byte
		want   PrometheusMetrics
	}{
		{
			config: []byte(`metrics:
  namespace: namespace
  subsystem: subsystem
  names:
    - name: name
      field: field
  labels:
    - name: label
      key: "key"
`),
			want: PrometheusMetrics{
				"field": prometheus.NewGaugeVec(
					prometheus.GaugeOpts{
						Namespace: "namespace",
						Subsystem: "subsystem",
						Name:      "name",
					}, []string{"label"},
				),
			},
		},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			v := NewTestConfig(tc.config)
			// Just make sure function does not panic.
			_ = NewPrometheusMetrics(v)
		})
	}
}

func TestGetPrometheusMetricsNames(t *testing.T) {
	cases := []struct {
		config []byte
		want   []map[string]string
	}{
		{
			config: []byte(`metrics:
  namespace: namespace
  subsystem: subsystem
  names:
    - name: metric_a
      field: field1
    - name: metric_b
      field: field2
    - name: metric_c
      field: field3
  labels:
    - name: label
      key: "key"
`),
			want: []map[string]string{
				{"name": "metric_a", "field": "field1"},
				{"name": "metric_b", "field": "field2"},
				{"name": "metric_c", "field": "field3"},
			},
		},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			v := NewTestConfig(tc.config)
			ret := GetPrometheusMetricsNames(v)
			assert.Equal(t, tc.want, ret)
		})
	}
}

func TestGetPrometheusMetricsLabels(t *testing.T) {
	cases := []struct {
		config []byte
		want   []map[string]string
	}{
		{
			config: []byte(`metrics:
  namespace: namespace
  subsystem: subsystem
  names:
    - name: name
      field: field
  labels:
    - name: label_a
      key: "key1"
    - name: label_b
      key: "key2"
    - name: label_c
      key: "key3"
`),
			want: []map[string]string{
				{"name": "label_a", "key": "key1"},
				{"name": "label_b", "key": "key2"},
				{"name": "label_c", "key": "key3"},
			},
		},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			v := NewTestConfig(tc.config)
			ret := GetPrometheusMetricsLabels(v)
			assert.Equal(t, tc.want, ret)
		})
	}
}

func TestGetPrometheusMetricsLabelNames(t *testing.T) {
	cases := []struct {
		config []byte
		want   []string
	}{
		{
			config: []byte(`metrics:
  namespace: namespace
  subsystem: subsystem
  names:
    - name: name
      field: field
  labels:
    - name: label_a
      key: "key1"
    - name: label_b
      key: "key2"
    - name: label_c
      key: "key3"
`),
			want: []string{"label_a", "label_b", "label_c"},
		},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			v := NewTestConfig(tc.config)
			ret := GetPrometheusMetricsLabelNames(v)
			assert.Equal(t, tc.want, ret)
		})
	}
}

func TestNewPrometheusLabelsFromValues(t *testing.T) {
	cases := []struct {
		config []byte
		want   prometheus.Labels
		m      map[string]any
	}{
		{
			config: []byte(`metrics:
  namespace: namespace
  subsystem: subsystem
  names:
    - name: name
      field: field
  labels:
    - name: label_a
      key: "key1"
    - name: label_b
      key: "key2"
    - name: label_c
      key: "key3"
`),
			m: map[string]any{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			want: prometheus.Labels{
				"label_a": "value1",
				"label_b": "value2",
				"label_c": "value3",
			},
		},
		{
			config: []byte(`metrics:
  namespace: namespace
  subsystem: subsystem
  names:
    - name: name
      field: field
  labels:
    - name: label_a
      key: "key1"
    - name: label_b
      key: "key2"
    - name: label_c
      key: "key3"
`),
			m: map[string]any{
				"key1": "value1",
				"key2": []any{"value2a", "value2b", "value2c"},
				"key3": map[string]any{
					"key3a": "value3a",
					"key3b": "value3b",
					"key3c": "value3c",
				},
			},
			want: prometheus.Labels{
				"label_a": "value1",
				"label_b": "value2a,value2b,value2c",
				"label_c": "key3a:value3a,key3b:value3b,key3c:value3c",
			},
		},
		{
			config: []byte(`metrics:
  namespace: namespace
  subsystem: subsystem
  names:
    - name: name
      field: field
  labels:
    - name: label_a
      key: "key1.key2"
    - name: label_b
      key: "key2"
    - name: label_c
      key: "key3"
`),
			m: map[string]any{
				"key1": map[string]any{"key2": "value1"},
				"key2": []any{"value2a", "value2b", "value2c"},
				"key3": map[string]any{
					"key3a": "value3a",
					"key3b": "value3b",
					"key3c": "value3c",
				},
			},
			want: prometheus.Labels{
				"label_a": "value1",
				"label_b": "value2a,value2b,value2c",
				"label_c": "key3a:value3a,key3b:value3b,key3c:value3c",
			},
		},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			v := NewTestConfig(tc.config)
			ret := NewPrometheusLabelsFromValues(v, tc.m)
			assert.Equal(t, tc.want, ret)
		})
	}
}
