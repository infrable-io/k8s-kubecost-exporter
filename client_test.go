// Copyright 2022 Infrable. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	urlpkg "net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Returns a pointer to the passed value.
func Ptr[T any](v T) *T {
	return &v
}

// Discard log output.
func DisableLogger() {
	// TODO: 'logger' should be mocked and its output tested.
	logger.SetOutput(io.Discard)
}

func TestGetURL(t *testing.T) {
	DisableLogger()
	// Mock time.Now.
	//
	// /!\ WARNING /!\
	// Remember to set 'Now' back to time.Now when no longer mocking time.Now.
	Now = func() time.Time {
		t, _ := time.Parse(time.RFC3339, "1970-01-01T01:33:07Z")
		return t
	}
	cases := []struct {
		host   string
		port   int
		path   string
		params map[string]any
		want   string
	}{
		{
			host:   "localhost",
			port:   9003,
			path:   "/allocation/compute",
			params: map[string]any{"window": "1m", "aggregate": "pod"},
			want: Ptr(urlpkg.URL{
				Scheme: "http",
				Host:   "localhost:9003",
				Path:   "/allocation/compute",
				RawQuery: urlpkg.Values{
					// 'window' should be the previous full 1 minute, that is:
					//   * 1970-01-01T01:32:00+07:00 -> 1970-01-01T01:33:00+07:00
					"window":    []string{"1970-01-01T01:32:00Z,1970-01-01T01:33:00Z"},
					"aggregate": []string{"pod"},
				}.Encode(),
			}).String(),
		},
		{
			host:   "localhost",
			port:   9003,
			path:   "/allocation/compute",
			params: map[string]any{"window": "30m", "aggregate": "pod"},
			want: Ptr(urlpkg.URL{
				Scheme: "http",
				Host:   "localhost:9003",
				Path:   "/allocation/compute",
				RawQuery: urlpkg.Values{
					// 'window' should be the previous full 30 minutes, that is:
					//   * 1970-01-01T01:03:00+07:00 -> 1970-01-01T01:33:00+07:00
					"window":    []string{"1970-01-01T01:03:00Z,1970-01-01T01:33:00Z"},
					"aggregate": []string{"pod"},
				}.Encode(),
			}).String(),
		},
		{
			host:   "localhost",
			port:   9003,
			path:   "/allocation/compute",
			params: map[string]any{"window": "1h", "aggregate": "pod"},
			want: Ptr(urlpkg.URL{
				Scheme: "http",
				Host:   "localhost:9003",
				Path:   "/allocation/compute",
				RawQuery: urlpkg.Values{
					// 'window' should be the previous full 1 hour, that is:
					//   * 1970-01-01T00:33:00+07:00 -> 1970-01-01T01:33:00+07:00
					"window":    []string{"1970-01-01T00:33:00Z,1970-01-01T01:33:00Z"},
					"aggregate": []string{"pod"},
				}.Encode(),
			}).String(),
		},
		{
			host:   "localhost",
			port:   9003,
			path:   "/allocation/compute",
			params: map[string]any{"window": "bad value", "aggregate": "pod"},
			want: Ptr(urlpkg.URL{
				Scheme: "http",
				Host:   "localhost:9003",
				Path:   "/allocation/compute",
				RawQuery: urlpkg.Values{
					// 'window' should default to 1m.
					//   * 1970-01-01T01:32:00+07:00 -> 1970-01-01T01:33:00+07:00
					"window":    []string{"1970-01-01T01:32:00Z,1970-01-01T01:33:00Z"},
					"aggregate": []string{"pod"},
				}.Encode(),
			}).String(),
		},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			c := AllocationAPIClient{}
			ret := c.GetURL(tc.host, tc.port, tc.path, tc.params)
			assert.Equal(t, tc.want, ret)
		})
	}
}

func TestGetValueByFieldNameFloat(t *testing.T) {
	cases := []struct {
		name string
		a    Allocation
		want float64
	}{
		{
			name: "CPUCores",
			a:    Allocation{CPUCores: 1337.0},
			want: 1337.0,
		},
		{
			name: "x",
			a:    Allocation{CPUCores: 1337.0},
			want: 0.0,
		},
		{
			name: "x",
			a:    Allocation{},
			want: 0.0,
		},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			ret := tc.a.GetValueByFieldNameFloat(tc.name)
			assert.Equal(t, tc.want, ret)
		})
	}
}

type MockHTTPClient struct {
	// MockGetFunc is a field on the MockHTTPClient struct that holds the
	// function to be called by the `Get` method.
	MockGetFunc func(url string) (resp *http.Response, err error)
}

func (m MockHTTPClient) Get(url string) (resp *http.Response, err error) {
	return m.MockGetFunc(url)
}

func TestGetAllocation(t *testing.T) {
	// Since we are attempting to test the scenario in which the client fails to
	// make a connection with the server, we forgo using the test server and
	// instead create a mock HTTP client, which simply returns an error for its
	// Get method.
	t.Run("connection refused", func(t *testing.T) {
		mockHTTPClient := &MockHTTPClient{
			MockGetFunc: func(url string) (resp *http.Response, err error) {
				return nil, fmt.Errorf("connection refused")
			},
		}
		c := AllocationAPIClient{
			Client: mockHTTPClient,
		}
		resp, err := c.GetAllocation("")
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrFailedAllocationAPICall)
		assert.ErrorContains(t, err, "connection refused")
	})
	// Define httptest test cases
	//
	// For example tests, see:
	//   * https://go.dev/src/net/http/httptest/example_test.go
	cases := []struct {
		Name             string
		StatusCode       int
		Header           http.Header
		Body             string
		WantResponseData []Allocation
		WantErr          error
	}{
		{
			Name:             "unexpected status code: 400",
			StatusCode:       http.StatusNotFound,
			Header:           nil,
			Body:             "404 page not found",
			WantResponseData: nil,
			WantErr:          fmt.Errorf("unexpected status code: 404"),
		},
		{
			Name:             "unexpected status code: 500",
			StatusCode:       http.StatusInternalServerError,
			Header:           nil,
			Body:             "",
			WantResponseData: nil,
			WantErr:          fmt.Errorf("unexpected status code: 500"),
		},
		{
			Name:       "unable to read response body",
			StatusCode: http.StatusOK,
			// This is a cheeky way of causing io.ReadAll to return an error without
			// having to create a custom io.Reader.
			//
			// The handler is instructed that the body has a length of 1 byte,
			// however, the response body is empty.
			Header:           map[string][]string{"Content-Length": {"1"}},
			Body:             "",
			WantResponseData: nil,
			WantErr:          fmt.Errorf("unable to read response body: unexpected EOF"),
		},
		{
			Name:             "unable to unmarshal response JSON",
			StatusCode:       http.StatusOK,
			Header:           nil,
			Body:             "",
			WantResponseData: nil,
			WantErr:          fmt.Errorf("unable to unmarshal response JSON: unexpected end of JSON input"),
		},
		{
			Name:       "success - data",
			StatusCode: http.StatusOK,
			Header:     nil,
			// NOTE: Formatting uses 2 tab space.
			Body: `{
"code": 200,
"status": "success",
"data": [
  {
    "my-cluster": {
      "name": "my-cluster",
      "properties": {
        "cluster": "my-cluster",
        "node": "minikube"
      },
      "window": {
        "start": "1970-01-01T01:32:00Z",
        "end": "1970-01-01T01:33:00Z"
      },
      "start": "1970-01-01T01:32:00Z",
      "end": "1970-01-01T01:33:00Z",
      "minutes": 0,
      "cpuCores": 0,
      "cpuCoreRequestAverage": 0,
      "cpuCoreUsageAverage": 0,
      "cpuCoreHours": 0,
      "cpuCost": 0,
      "cpuCostAdjustment": 0,
      "cpuEfficiency": 0,
      "gpuCount": 0,
      "gpuHours": 0,
      "gpuCost": 0,
      "gpuCostAdjustment": 0,
      "networkTransferBytes": 0,
      "networkReceiveBytes": 0,
      "networkCost": 0,
      "networkCostAdjustment": 0,
      "loadBalancerCost": 0,
      "loadBalancerCostAdjustment": 0,
      "pvBytes": 0,
      "pvByteHours": 0,
      "pvCost": 0,
      "pvs": {
        "cluster=my-cluster:name=my-pvc": {
          "byteHours": 0,
          "cost": 0
        }
      },
      "pvCostAdjustment": 0,
      "ramBytes": 0,
      "ramByteRequestAverage": 0,
      "ramByteUsageAverage": 0,
      "ramByteHours": 0,
      "ramCost": 0,
      "ramCostAdjustment": 0,
      "ramEfficiency": 0,
      "sharedCost": 0,
      "externalCost": 0,
      "totalCost": 0,
      "totalEfficiency": 0,
      "rawAllocationOnly": {
        "cpuCoreUsageMax": 0,
        "ramByteUsageMax": 0
      }
    }
  }
]}`,
			WantResponseData: []Allocation{{
				Name: "my-cluster",
				Properties: map[string]any{
					"cluster": "my-cluster",
					"node":    "minikube",
				},
				Window: map[string]any{
					"start": "1970-01-01T01:32:00Z",
					"end":   "1970-01-01T01:33:00Z",
				},
				Start:                      "1970-01-01T01:32:00Z",
				End:                        "1970-01-01T01:33:00Z",
				Minutes:                    0.0,
				CPUCores:                   0.0,
				CPUCoreRequestAverage:      0.0,
				CPUCoreUsageAverage:        0.0,
				CPUCoreHours:               0.0,
				CPUCost:                    0.0,
				CPUCostAdjustment:          0.0,
				CPUEfficiency:              0.0,
				GPUCount:                   0.0,
				GPUHours:                   0.0,
				GPUCost:                    0.0,
				GPUCostAdjustment:          0.0,
				NetworkTransferBytes:       0.0,
				NetworkReceiveBytes:        0.0,
				NetworkCost:                0.0,
				NetworkCostAdjustment:      0.0,
				LoadBalancerCost:           0.0,
				LoadBalancerCostAdjustment: 0.0,
				PVBytes:                    0.0,
				PVByteHours:                0.0,
				PVCost:                     0.0,
				PVs: map[string]any{
					"cluster=my-cluster:name=my-pvc": map[string]any{
						"byteHours": 0.0,
						"cost":      0.0,
					},
				},
				PVCostAdjustment:      0.0,
				RAMBytes:              0.0,
				RAMByteRequestAverage: 0.0,
				RAMByteUsageAverage:   0.0,
				RAMByteHours:          0.0,
				RAMCost:               0.0,
				RAMCostAdjustment:     0.0,
				RAMEfficiency:         0.0,
				SharedCost:            0.0,
				ExternalCost:          0.0,
				TotalCost:             0.0,
				TotalEfficiency:       0.0,
				RawAllocationOnly: map[string]any{
					"cpuCoreUsageMax": 0.0,
					"ramByteUsageMax": 0.0,
				},
			},
			},
			WantErr: nil,
		},
		{
			Name:             "success - no data",
			StatusCode:       http.StatusOK,
			Header:           nil,
			Body:             `{"code":200,"status":"success","data":[]}`,
			WantResponseData: []Allocation{},
			WantErr:          nil,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			// A Server is an HTTP server listening on a system-chosen port on the
			// local loopback interface, for use in end-to-end HTTP tests.
			//
			// The URL of the HTTP server is of the form http://ipaddr:port with no
			// trailing slash.
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != "GET" {
					t.Errorf("Expected request method 'GET', got: %s", req.Method)
				}
				for k, vs := range tc.Header {
					for _, v := range vs {
						w.Header().Set(k, v)
					}
				}
				w.WriteHeader(tc.StatusCode)
				w.Write([]byte(tc.Body))
			}))
			defer ts.Close()
			c := AllocationAPIClient{
				Client: &http.Client{},
			}
			// HTTP requests must be made to the URL of the test server.
			data, err := c.GetAllocation(ts.URL)
			if err != nil {
				assert.ErrorIs(t, err, ErrFailedAllocationAPICall)
				assert.ErrorContains(t, err, tc.WantErr.Error())
			}
			assert.Equal(t, tc.WantResponseData, data)
		})
	}
}
