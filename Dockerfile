# Copyright (c) 2019 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
FROM golang:latest AS builder

WORKDIR /go/src/testapi

COPY . ./
RUN go get ./... 
RUN CGO_ENABLED=0 go build -o /go/bin/api_server

FROM alpine:latest
COPY --from=builder /go/bin/api_server /api_server
ENTRYPOINT ["/api_server"]
