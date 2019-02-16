// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"log"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport/zipkin"

	"github.com/mcastelino/testapi/restapi/operations"
)

const (
	//TODO: The zipkin host needs to come from the env variable
	//The kubernetes pod spec will include this address
	//zipkinURL = "http://zipkin:9411/api/v1/spans"
	zipkinURL = "http://localhost:9411/api/v1/spans"
)

//go:generate swagger generate server --target ../../testapi --name Testapi --spec ../api.json

func configureFlags(api *operations.TestapiAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.TestapiAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	// Jaeger tracer can be initialized with a transport that will
	// report tracing Spans to a Zipkin backend
	transport, err := zipkin.NewHTTPTransport(
		zipkinURL,
		zipkin.HTTPBatchSize(1),
		zipkin.HTTPLogger(jaeger.StdLogger),
	)
	if err != nil {
		log.Fatalf("Cannot initialize HTTP transport: %v", err)
	}

	// create Jaeger tracer
	//tracer, closer := jaeger.NewTracer(
	tracer, _ := jaeger.NewTracer(
		"server",                     //TODO: Should be unique per instance of the app (FQDN)
		jaeger.NewConstSampler(true), // sample all traces
		jaeger.NewRemoteReporter(transport),
	)

	/* TODO
	// Close the tracer to guarantee that all spans that could
	// be still buffered in memory are sent to the tracing backend
	defer closer.Close()
	*/

	return nethttp.Middleware(tracer, handler)
}
