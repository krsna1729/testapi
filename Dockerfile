# Copyright (c) 2019 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
FROM clearlinux:latest AS sslbuilder
ARG HOST="127.0.0.1"
WORKDIR /sslcerts
RUN openssl genrsa -out /sslcerts/testCA.key 2048 && \
    openssl req -new -key /sslcerts/testCA.key -subj "/CN=${HOST}" -out /sslcerts/testCA.csr && \
    openssl x509 -req -days 365 -in /sslcerts/testCA.csr -signkey /sslcerts/testCA.key -out /sslcerts/test.crt

FROM golang:latest AS builder
WORKDIR /testapi
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -o api_server

FROM alpine:latest
VOLUME /sslcerts
COPY --from=builder /testapi/api_server /api_server
COPY --from=sslbuilder /sslcerts/* /sslcerts/
EXPOSE 8888
EXPOSE 6060
ENTRYPOINT ["/api_server", "--tls-certificate=/sslcerts/test.crt", "--tls-key=/sslcerts/testCA.key", "--host=0.0.0.0", "--tls-port=8888"]
