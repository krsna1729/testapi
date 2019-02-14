# testapi

API Responsive testing across different languages using Open API

# How to build

```
docker build -t test-api-server:latest .
```

# How to run local tests

## Launch the server with profiling enabled
```
docker run --rm --name=myserver -p 8888:8888 -p 6060:6060 test-api-server
```

## Test it

```
docker cp myserver:/sslcerts/test.crt /tmp
curl --cacert /tmp/test.crt https://127.0.0.1:8888
```

Note: The profiling information is available at http://localhost:6060/debug/pprof/

# To automatically re-generate the `go` client and server code

```
go get -u github.com/go-swagger/go-swagger
swagger generate server -f ./api.json
swagger generate client -f ./api.json
```
