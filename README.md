# testapi

API Responsive testing across different languages using Open API

# How to build

```
docker build -t test-api-server:latest .
```

# How to run local tests

## Launch the server with profiling enabled
```
docker run --rm --net=host --name=myserver test-api-server
```

## Test it

Run zipkin to collect and visulize the traces
```
docker run -d -p 9410-9411:9410-9411 --name=zipkin openzipkin/zipkin:1.12.0
```

Invoke the service
```
docker cp myserver:/sslcerts/test.crt /tmp
curl --cacert /tmp/test.crt https://127.0.0.1:8888/
ab -n 100000 -c 1000 https://127.0.0.1:8888/
```

Visualize the results
```
firefox http://localhost:9411/
```

Note: The profiling information is available at http://localhost:6060/debug/pprof/

# To automatically re-generate the `go` client and server code

```
go get -u github.com/go-swagger/go-swagger
swagger generate server -f ./api.json
swagger generate client -f ./api.json
```
