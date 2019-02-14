#! /bin/bash

TESTAPI_NS=${1:-"default"}
TESTAPI_SVC="testapi-svc"

# Create certs for our testapi for kubernetes
mkdir ./deploy
openssl genrsa -out ./deploy/testapiCA.key 2048
openssl req -new -key ./deploy/testapiCA.key -subj "/CN=${TESTAPI_SVC}.${TESTAPI_NS}.svc" -out ./deploy/testapiCA.csr 
openssl x509 -req -days 365 -in ./deploy/testapiCA.csr -signkey ./deploy/testapiCA.key -out ./deploy/testapi.crt

# Create certs secrets for k8s
kubectl create secret generic \
    ${TESTAPI_SVC}-certs \
    --from-file=key.pem=./deploy/testapiCA.key \
    --from-file=cert.pem=./deploy/testapi.crt \
    --dry-run -o yaml > ./deploy/testapi-certs.yaml

# Create the CA bundle
CA_BUNDLE=$(cat ./deploy/testapi.crt | base64 -w0)
