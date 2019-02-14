# testapi

API Responsive testing across different languages using Open API

# How to build

```
docker build -t test-api-server:latest .
```

# To automatically re-generate the `go` client and server code

```
go get -u github.com/go-swagger/go-swagger
swagger generate server -f ./api.json
swagger generate client -f ./api.json
```
