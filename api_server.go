package main

import (
	"log"
	"os"

	"net/http"
	_ "net/http/pprof"

	loads "github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	flags "github.com/jessevdk/go-flags"
	"github.com/mcastelino/testapi/models"
	"github.com/mcastelino/testapi/restapi"
	"github.com/mcastelino/testapi/restapi/operations"
)

const (
	pprofURL = "0.0.0.0:6060"
)

func getParams(params operations.GetParams) middleware.Responder {
	response, err := os.Hostname()
	if err != nil {
		return operations.NewGetDefault(500).WithPayload(&models.Error{
			Message: "failed to retrieve profile",
		})
	}

	return operations.NewGetOK().WithPayload(&models.Profile{response})
}

func main() {
	// For pprof
	go func() {
		log.Println(http.ListenAndServe(pprofURL, nil))
	}()

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewTestapiAPI(swaggerSpec)
	server := restapi.NewServer(api)
	defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "testapi"
	parser.LongDescription = "A Simple Test API"

	server.ConfigureFlags()
	for _, optsGroup := range api.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}

	/* Setup the handlers */
	api.GetHandler = operations.GetHandlerFunc(getParams)

	server.ConfigureAPI()

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}

}
