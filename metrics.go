// Copyright 2023 Infrable. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Generate Prometheus metrics from configuration.
//
// For documentation on the Go client library for Prometheus, see the
// following:
//   - https://pkg.go.dev/github.com/prometheus/client_golang/prometheus
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

// PrometheusMetrics represents a collection of Allocation field name -> metric
// mappings.
type PrometheusMetrics map[string]*prometheus.GaugeVec

// Create new PrometheusMetrics from configuration.
func NewPrometheusMetrics(v *viper.Viper) PrometheusMetrics {
	// Create a map (PrometheusMetrics) to associate Allocation field names with
	// Prometheus metrics.
	names := GetPrometheusMetricsNames(v)
	labels := GetPrometheusMetricsLabelNames(v)
	metrics := make(PrometheusMetrics, len(names))
	for _, n := range names {
		// NOTE: Do NOT use promauto.NewGaugeVec. This function automatically
		// registers the GaugeVec with the prometheus.DefaultRegisterer. We do not
		// want to use DefaultRegisterer, since we register Collectors with a new
		// vanilla Registry explicitly.
		metrics[n["field"]] = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: v.GetString("metrics.namespace"),
				Subsystem: v.GetString("metrics.subsystem"),
				Name:      n["name"],
			}, labels,
		)
	}
	return metrics
}

// Get Prometheus metric names as slice of mappings.
//
// It is not possible to retrieve a slice of map[string]string values (ex.
// []map[string]string) using Viper's utility functions. Instead, GetStringMap
// is used to retrieve the value associated with the key as a map of
// interfaces. A slice of map[string]string elements is constructed by
// asserting the type of the interface values, which should always be of type
// map[string]string.
func GetPrometheusMetricsNames(v *viper.Viper) []map[string]string {
	metrics := v.GetStringMap("metrics")
	ns := metrics["names"].([]any)
	names := make([]map[string]string, len(ns))
	for i, n := range ns {
		names[i] = map[string]string{
			"name":  GetElementOrZeroValue[string]("name", n.(map[string]any)),
			"field": GetElementOrZeroValue[string]("field", n.(map[string]any)),
		}
	}
	return names
}

// Get Prometheus metric labels as slice of mappings.
//
// It is not possible to retrieve a slice of map[string]string values (ex.
// []map[string]string) using Viper's utility functions. Instead, GetStringMap
// is used to retrieve the value associated with the key as a map of
// interfaces. A slice of map[string]string elements is constructed by
// asserting the type of the interface values, which should always be of type
// map[string]string.
func GetPrometheusMetricsLabels(v *viper.Viper) []map[string]string {
	metrics := v.GetStringMap("metrics")
	ls := metrics["labels"].([]any)
	labels := make([]map[string]string, len(ls))
	for i, l := range ls {
		labels[i] = map[string]string{
			"name": GetElementOrZeroValue[string]("name", l.(map[string]any)),
			"key":  GetElementOrZeroValue[string]("key", l.(map[string]any)),
		}
	}
	return labels
}

// Get Prometheus metric label names as slice of strings.
func GetPrometheusMetricsLabelNames(c *viper.Viper) []string {
	labels := GetPrometheusMetricsLabels(c)
	names := make([]string, len(labels))
	for i, l := range labels {
		names[i] = l["name"]
	}
	return names
}

// Create new prometheus.Labels from configuration and values.
//
// /!\ WARNING /!\
// Label cardinality must be consistent when retrieving a Gauge with
// GetMetricWith.
//
//	> An error is returned if the number and names of the Labels are
//	  inconsistent with those of the variable labels in Desc (minus any
//	  curried labels).
func NewPrometheusLabelsFromValues(v *viper.Viper, m map[string]any) prometheus.Labels {
	// Default to comma-separated string of elements.
	const sep = ","
	// Only labels defined via configuration are used to construct new
	// prometheus.Labels.
	labels := prometheus.Labels{}
	for _, l := range GetPrometheusMetricsLabels(v) {
		var v string
		elem := GetElementFromKey(l["key"], m)
		switch elem.(type) {
		case string:
			v = elem.(string)
		case []any:
			v = SortAndJoinSlice(elem.([]any), sep)
		case map[string]any:
			v = SortAndJoinMap(elem.(map[string]any), sep)
		}
		labels[l["name"]] = v
	}
	return labels
}
