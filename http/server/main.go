// Copyright 2018, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"

	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
)

var (
	upstreamURI   string
	downstreamURI string
	serviceName   string
	cpuWorkMs     time.Duration //CPU intensive busywork in ms
)

// busyWork does meaningless work for the specified duration,
// so we can observe CPU usage.
func busyWork(d time.Duration) int {
	var n int
	afterCh := time.After(d)
	for {
		select {
		case <-afterCh:
			return n
		default:
			n++
		}
	}
}

func homeHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello:%s", serviceName)

	if cpuWorkMs != 0 {
		busyWork(cpuWorkMs)
	}

	if downstreamURI != "" {
		r, _ := http.NewRequest("GET", downstreamURI, nil)

		// Propagate the trace header info in the outgoing requests.
		r = r.WithContext(req.Context())
		client := &http.Client{Transport: &ochttp.Transport{}}
		resp, err := client.Do(r)

		if err != nil {
			log.Println(err)
		} else {
			if body, err := ioutil.ReadAll(resp.Body); err == nil {
				fmt.Fprintf(w, ":%s", string(body))
			}
			resp.Body.Close()
		}
	}

}

func main() {
	cpuBusywork := os.Getenv("CPU_BUSYWORK")
	if cpuBusywork != "" {
		if i, err := strconv.Atoi(cpuBusywork); err == nil {
			cpuWorkMs = time.Duration(i) * time.Millisecond
		}
	}

	fmt.Println("cpuWorkMs := ", cpuWorkMs)

	// e.g: http://localhost:8888/
	upstreamURI = os.Getenv("UPSTREAM_URI")
	if upstreamURI == "" {
		fmt.Println("Error: UPSTREAM_URI not present")
		os.Exit(1)
	}

	// URI of the downstream service
	// e.g: http://localhost:8889/
	downstreamURI = os.Getenv("DOWNSTREAM_URI")

	// Setup tracing
	// reporterURI: zipkin reporter URI
	reporterURI := os.Getenv("REPORTER_URI")
	if reporterURI == "" {
		reporterURI = "http://localhost:9411/api/v2/spans"
	}

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "8887"
	}

	serviceName = os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		var err error
		if serviceName, err = os.Hostname(); err != nil {
			serviceName = "service"
		}
	}

	localEndpoint, err := openzipkin.NewEndpoint(serviceName, ":0")
	if err != nil {
		log.Fatalf("Failed to create Zipkin localEndpoint with URI %q error: %v", upstreamURI, err)
	}

	reporter := zipkinHTTP.NewReporter(reporterURI)
	ze := zipkin.NewExporter(reporter, localEndpoint)

	// And now finally register it as a Trace Exporter
	trace.RegisterExporter(ze)

	//TODO: Switch to trace.ProbabilitySampler if needed
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	// Setup metrics
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: serviceName,
	})
	if err != nil {
		log.Fatalf("Failed to create the Prometheus exporter: %v", err)
	}

	// register it as a stats exporter.
	view.RegisterExporter(pe)
	// Report stats at every second.
	view.SetReportingPeriod(1 * time.Second)

	// Register the built in views when using ichttp
	if err := view.Register(ochttp.ClientLatencyView, ochttp.ServerLatencyView); err != nil {
		log.Fatalf("Failed to register metrics view: %v", err)
	}

	//Start the metrics server on the user specified port
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", pe)
		if err := http.ListenAndServe(":"+metricsPort, mux); err != nil {
			log.Fatalf("Failed to run Prometheus /metrics endpoint: %v", err)
		}
	}()

	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler).Methods(http.MethodGet, http.MethodHead)

	handler := &ochttp.Handler{ // add opencensus instrumentation
		Handler:     r,
		Propagation: &b3.HTTPFormat{}}

	log.Fatal("Server", http.ListenAndServe(upstreamURI, handler))

}
