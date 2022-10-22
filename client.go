// Copyright 2022 Infrable. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// An HTTP client for interacting with the Kubecost Allocation API.
//
// For documentation on the Go standard library net/http package, see the
// following:
//   - https://pkg.go.dev/net/http
//
// For documentation on the Kubecost Allocation API, see the following:
//   - https://docs.kubecost.com/apis/apis/allocation
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"reflect"
	"time"
)

// Used for mocking time.Now() in testing.
var Now = time.Now

// Using an interface, we define a set of method signatures for making HTTP
// requests to the Kubecost Allocation API. This allows for the flexibility to
// mock certain functions in testing.
type AllocationAPI interface {
	GetURL(string, int, string, map[string]any) string
	GetAllocation(string) ([]Allocation, error)
}

// HTTPClient is a common interface that specifies the method signatures on the
// http.Client struct that are to be mocked.
//
// See: net/http/client.go
type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}

// AllocationAPIClient is an application-specific HTTP client. It implements
// the AllocationAPI interface. It also holds an http.Client, which can be
// overridden for testing purposes.
type AllocationAPIClient struct {
	Client HTTPClient
}

// ErrFailedAllocationAPICall is returned when an error or bad response is
// returned from the Kubecost Allocation API.
var ErrFailedAllocationAPICall = errors.New("Failed to retrieve cost allocation data from Allocation API")

// Kubecost API response.
//
// For the Kubecost `Response` struct, see pkg/costmodel/router.go in the
// OpenCost GitHub repository:
//   - https://github.com/opencost/opencost
type Response struct {
	Code    int                     `json:"code"`
	Status  string                  `json:"status"`
	Data    []map[string]Allocation `json:"data"`
	Message string                  `json:"message,omitempty"`
	Warning string                  `json:"warning,omitempty"`
}

// Kubecost Allocation.
//
// For the Kubecost `Allocation` struct, see pkg/kubecost/allocation.go in the
// OpenCost GitHub repository:
//   - https://github.com/opencost/opencost
type Allocation struct {
	Name                       string         `json:"name"`
	Properties                 map[string]any `json:"properties"`
	Window                     map[string]any `json:"window"`
	Start                      string         `json:"start"` // TODO: Use time.Time.
	End                        string         `json:"end"`   // TODO: Use time.Time.
	Minutes                    float64        `json:"minutes"`
	CPUCores                   float64        `json:"cpuCores"`
	CPUCoreRequestAverage      float64        `json:"cpuCoreRequestAverage"`
	CPUCoreUsageAverage        float64        `json:"cpuCoreUsageAverage"`
	CPUCoreHours               float64        `json:"cpuCoreHours"`
	CPUCost                    float64        `json:"cpuCost"`
	CPUCostAdjustment          float64        `json:"cpuCostAdjustment"`
	CPUEfficiency              float64        `json:"cpuEfficiency"`
	GPUCount                   float64        `json:"gpuCount"`
	GPUHours                   float64        `json:"gpuHours"`
	GPUCost                    float64        `json:"gpuCost"`
	GPUCostAdjustment          float64        `json:"gpuCostAdjustment"`
	NetworkTransferBytes       float64        `json:"networkTransferBytes"`
	NetworkReceiveBytes        float64        `json:"networkReceiveBytes"`
	NetworkCost                float64        `json:"networkCost"`
	NetworkCostAdjustment      float64        `json:"networkCostAdjustment"`
	LoadBalancerCost           float64        `json:"loadBalancerCost"`
	LoadBalancerCostAdjustment float64        `json:"loadBalancerCostAdjustment"`
	PVBytes                    float64        `json:"pvBytes"`
	PVByteHours                float64        `json:"pvByteHours"`
	PVCost                     float64        `json:"pvCost"`
	PVs                        map[string]any `json:"pvs"`
	PVCostAdjustment           float64        `json:"pvCostAdjustment"`
	RAMBytes                   float64        `json:"ramBytes"`
	RAMByteRequestAverage      float64        `json:"ramByteRequestAverage"`
	RAMByteUsageAverage        float64        `json:"ramByteUsageAverage"`
	RAMByteHours               float64        `json:"ramByteHours"`
	RAMCost                    float64        `json:"ramCost"`
	RAMCostAdjustment          float64        `json:"ramCostAdjustment"`
	RAMEfficiency              float64        `json:"ramEfficiency"`
	SharedCost                 float64        `json:"sharedCost"`
	ExternalCost               float64        `json:"externalCost"`
	TotalCost                  float64        `json:"totalCost"`
	TotalEfficiency            float64        `json:"totalEfficiency"`
	RawAllocationOnly          map[string]any `json:"rawAllocationOnly"`
}

// Get the value of the struct's field by name.
//
// This method leverages reflection to examine its own structure.
func (a Allocation) GetValueByFieldNameFloat(name string) float64 {
	v := reflect.ValueOf(a)
	// FieldByName returns the zero Value (Value{}) if no field was found.
	fv := v.FieldByName(name)
	// Check whether fv is the zero Value (Value{}).
	// The zero Value represents no value. Its IsValid method returns false.
	if !fv.IsValid() {
		var v float64
		return v
	}
	return fv.Float()
}

// Generate Kubecost Allocation API URL.
func (c AllocationAPIClient) GetURL(host string, port int, path string, params map[string]any) string {
	url := urlpkg.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   path,
	}
	query := url.Query()
	for k, v := range params {
		query.Add(k, v.(string))
	}
	// Durations (such as 30m, 12h, 7d) are calculated as a precise start and end
	// time and added to the query as a comma-separated RFC3339 date pair for the
	// previous minute:
	//
	//   Example:
	//
	//     Given the current time of 2006-01-02T15:04:05Z07:00, a window of 1m,
	//     30m, and 1h would yield the following date pairs:
	//
	//       * 1m: 2006-01-02T15:03:00,2006-01-02T15:04:00
	//       * 30m: 2006-01-02T14:34:00,2006-01-02T15:04:00
	//       * 1h: 2006-01-02T14:04:00,2006-01-02T15:04:00
	//
	// This ensures that the window is the exact specified duration, since the
	// Kubecost Allocation API uses an end time of when the request was made when
	// the 'window' parameter contains a duration.
	dur, err := time.ParseDuration(params["window"].(string))
	if err != nil {
		logger.Printf("Error parsing 'window' config: %v. Defaulting to 1m", err)
		dur, _ = time.ParseDuration("1m")
	}
	now := Now()
	end := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
	start := end.Add(-dur)
	window := fmt.Sprintf("%s,%s", start.Format(time.RFC3339), end.Format(time.RFC3339))
	query.Set("window", window)
	url.RawQuery = query.Encode()
	return url.String()
}

// Retrieve cost allocation data from the Kubecost Allocation API.
func (c AllocationAPIClient) GetAllocation(url string) ([]Allocation, error) {
	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedAllocationAPICall, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"%w: unexpected status code: %d", ErrFailedAllocationAPICall, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: unable to read response body: %v", ErrFailedAllocationAPICall, err)
	}
	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf(
			"%w: unable to unmarshal response JSON: %v", ErrFailedAllocationAPICall, err)
	}
	as := []Allocation{}
	// Data is grouped by aggregation, that is, there is an Allocation for each
	// unique value for the aggregation.
	for _, aggregation := range r.Data {
		// The key of the Allocation is the unique value for the Allocation. The
		// key is discarded and its value (an Allocation) is appended to the slice
		// of Allocations.
		for _, a := range aggregation {
			as = append(as, a)
		}
	}
	return as, nil
}
