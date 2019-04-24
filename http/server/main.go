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
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"

	"github.com/kavehmz/prime"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
)

var (
	upstreamURI   string
	downstreamURI string
	serviceName   string
	jobFile       string
	primeMax      uint64
)

var (
	// The latency in milliseconds
	mPrimeLatencyMs    = stats.Float64("prime/latency", "The latency in milliseconds per prime generation call", stats.UnitMilliseconds)
	mBusyworkLatencyMs = stats.Float64("prime/latency", "The latency in milliseconds per busywork call", stats.UnitMilliseconds)
)

var (
	workloadLatencyView = &view.View{
		Name:        "prime/latency",
		Measure:     mPrimeLatencyMs,
		Description: "The distribution of the latencies for prime calculation",

		// Latency in buckets: in ms
		// TODO: Make this a runtime configuration vs build time
		Aggregation: view.Distribution(1, 8, 9, 10, 11, 12, 13, 14, 15, 16, 18, 20, 22, 24, 26, 28, 30, 35, 40, 45, 50, 100, 200, 400),
		TagKeys:     []tag.Key{keyMethod}}

	busyworkLatencyView = &view.View{
		Name:        "busywork/latency",
		Measure:     mBusyworkLatencyMs,
		Description: "The distribution of the latencies for CPU busywork",

		// Latency in buckets: in ms
		// TODO: Make this a runtime configuration vs build time
		Aggregation: view.Distribution(1, 8, 9, 10, 11, 12, 13, 14, 15, 16, 18, 20, 22, 24, 26, 28, 30, 35, 40, 45, 50, 100, 200, 400),
		TagKeys:     []tag.Key{keyMethod}}
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

func sinceInMilliseconds(startTime time.Time) int64 {
	return int64(time.Since(startTime).Nanoseconds()) / 1e6
}

func downstreamHandler(work string, w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello:%s:%s", serviceName, work)

	if downstreamURI != "" {
		r, _ := http.NewRequest("GET", downstreamURI+work, nil)

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

func homeHandler(w http.ResponseWriter, req *http.Request) {

	downstreamHandler("/", w, req)
}

func jobHandler(w http.ResponseWriter, req *http.Request) {
	if jobFile == "" {
		return
	}
	var cmd *exec.Cmd
	cmd = exec.Command("stress-ng", "--job", jobFile)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
	}

	downstreamHandler("/stress-ng", w, req)
}

var (
	keyMethod, _ = tag.NewKey("method")
)

func busyworkHandler(w http.ResponseWriter, req *http.Request) {
	_, span := trace.StartSpan(req.Context(), "busyHandler")
	_, spanp := trace.StartSpan(req.Context(), "busy")
	defer span.End()

	startTime := time.Now()

	ctx, err := tag.New(context.Background(), tag.Insert(keyMethod, "busyHandler"))
	if err != nil {
		return
	}

	busyWork(10 * time.Millisecond)

	ms := float64(time.Since(startTime).Nanoseconds()) / 1e6
	stats.Record(ctx, mPrimeLatencyMs.M(ms))
	spanp.End()

	downstreamHandler("/busywork", w, req)
}

func primeHandler(w http.ResponseWriter, req *http.Request) {
	_, span := trace.StartSpan(req.Context(), "primeHandler")
	_, spanp := trace.StartSpan(req.Context(), "primeCalc")
	defer span.End()

	startTime := time.Now()

	ctx, err := tag.New(context.Background(), tag.Insert(keyMethod, "primeHandler"))
	if err != nil {
		return
	}

	if primeMax != 0 {
		p := prime.Primes(primeMax)
		if len(p) == 0 {
			log.Printf("primes finished with error")
		}
	}

	ms := float64(time.Since(startTime).Nanoseconds()) / 1e6
	stats.Record(ctx, mPrimeLatencyMs.M(ms))
	spanp.End()

	downstreamHandler("/prime", w, req)
}

func forkHandler(w http.ResponseWriter, req *http.Request) {

	var cmd *exec.Cmd
	cmd = exec.Command("date")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
	}

	downstreamHandler("/fork", w, req)
}

func main() {
	genPrime := os.Getenv("PRIME_MAX")
	if genPrime != "" {
		var err error
		primeMax, err = strconv.ParseUint(genPrime, 10, 64)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	}

	jobFile = os.Getenv("JOBFILE")

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

	//Reduce sampling as it introduces trace induced latency
	probSampler := trace.ProbabilitySampler(1 / 100.0)
	trace.ApplyConfig(trace.Config{DefaultSampler: probSampler})

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
	if err := view.Register(ochttp.ClientLatencyView, ochttp.ServerLatencyView, workloadLatencyView); err != nil {
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
	r.HandleFunc("/busywork", busyworkHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/prime", primeHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/stress-ng", jobHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/fork", forkHandler).Methods(http.MethodGet, http.MethodHead)

	handler := &ochttp.Handler{ // add opencensus instrumentation
		Handler:     r,
		Propagation: &b3.HTTPFormat{}}

	log.Fatal("Server", http.ListenAndServe(upstreamURI, handler))

}
