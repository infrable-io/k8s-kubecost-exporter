// Copyright 2023 Infrable. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// An HTTP server for exposing cost allocation metrics retrieved from Kubecost.
//
// Metrics are exposed via an HTTP metrics endpoint. Applications that provide
// a Prometheus OpenMetrics integration can gather cost allocation metrics from
// this endpoint to store and visualize the data.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

var logger = log.New(os.Stdout, "", log.Lshortfile)

// Instead of initializing the configuration in main(), a package-level
// variable is declared and initialized via init().
var Config *viper.Viper

// See: https://go.dev/doc/effective_go#init
func init() {
	var err error
	Config, err = NewConfig()
	if err != nil {
		logger.Printf("Error during configuration: %s\n", err)
	}
}

// Retrieve cost allocation data and update metrics.
func RecordMetrics(c AllocationAPI, metrics PrometheusMetrics) {
	host, port, path, params := Config.GetString("api.host"), Config.GetInt("api.port"),
		Config.GetString("api.path"), Config.GetStringMap("api.parameters")
	url := c.GetURL(host, port, path, params)
	i, err := time.ParseDuration(Config.GetString("server.update_interval"))
	if err != nil {
		logger.Printf("Error parsing 'update_interval' config: %v. Defaulting to 1m", err)
		i, _ = time.ParseDuration("1m")
	}
	ticker := time.NewTicker(i)
	go func() {
		for {
			as, err := c.GetAllocation(url)
			if err != nil {
				logger.Printf("%s\n", err)
			}
			for _, a := range as {
				// Allocation properties are used to set Prometheus metric label
				// values.
				lvs := a.Properties
				// 'name' is the Allocation struct field name for the corresponding
				// Prometheus metric.
				for name, metric := range metrics {
					ls := NewPrometheusLabelsFromValues(Config, lvs)
					m, err := metric.GetMetricWith(ls)
					if err != nil {
						logger.Printf(
							"Number of label values is not the same as the number of "+
								"variable labels in Desc: %s\n", err)
						continue
					}
					m.Set(a.GetValueByFieldNameFloat(name))
				}
			}
			<-ticker.C
		}
	}()
}

func main() {
	c := AllocationAPIClient{
		Client: &http.Client{},
	}
	// NewRegistry creates a new vanilla Registry without any Collectors
	// pre-registered (see promhttp.Handler() for default Collectors).
	r := prometheus.NewRegistry()
	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})
	// Generate Prometheus metrics from configuration.
	metrics := NewPrometheusMetrics(Config)
	for _, m := range metrics {
		r.MustRegister(m)
	}
	// Retrieve data from the Kubecost Allocation API and update metrics.
	RecordMetrics(c, metrics)
	// Register metrics HTTP endpoint and handle requests on incoming
	// connections.
	pattern, port := Config.GetString("server.path"), fmt.Sprintf(":%s", Config.GetString("server.port"))
	http.Handle(pattern, handler)
	log.Fatal(http.ListenAndServe(port, nil))
}
