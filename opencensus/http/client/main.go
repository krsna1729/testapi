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

	"go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"

	"github.com/codahale/hdrhistogram"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
)

func sinceInMilliseconds(startTime time.Time) int64 {
	return int64(time.Since(startTime).Nanoseconds()) / 1e6
}

func main() {
	serviceName := "client"
	server := os.Getenv("SERVER_URI")
	if server == "" {
		server = "http://localhost:8888"
	}

	count := os.Getenv("COUNT")
	numRequests, err := strconv.ParseInt(count, 0, 64)
	if err != nil {
		numRequests = 1
	}

	// Setup tracing
	// reporterURI: zipkin reporter URI
	reporterURI := os.Getenv("REPORTER_URI")
	if reporterURI == "" {
		reporterURI = "http://localhost:9411/api/v2/spans"
	}

	localEndpoint, err := openzipkin.NewEndpoint(serviceName, ":0")
	if err != nil {
		log.Fatalf("Failed to create Zipkin localEndpoint with URI %q error: %v", serviceName, err)
	}

	reporter := zipkinHTTP.NewReporter(reporterURI)
	defer reporter.Close()
	ze := zipkin.NewExporter(reporter, localEndpoint)

	// And now finally register it as a Trace Exporter
	trace.RegisterExporter(ze)

	//TODO: Switch to trace.ProbabilitySampler if needed
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	client := &http.Client{Transport: &ochttp.Transport{}}

	h := hdrhistogram.New(1, 10000, 3)
	for i := int64(0); i < numRequests; i++ {
		start := time.Now()
		resp, err := client.Get(server)
		if err != nil {
			log.Printf("Failed to get response: %v", err)
		} else {
			if _, err := ioutil.ReadAll(resp.Body); err != nil {
				fmt.Println("error:", err)
			}
			resp.Body.Close()
		}
		elapsed := sinceInMilliseconds(start)
		if err := h.RecordValue(int64(elapsed)); err != nil {
			fmt.Println("error:", err)
		}
	}

	distribution := h.CumulativeDistribution()
	fmt.Println("Distribution:")
	fmt.Println(distribution)

}
