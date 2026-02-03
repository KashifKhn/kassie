#!/bin/bash

set -e

PROTO_DIR="api/proto"
GO_OUT_DIR="api/gen/go"
TS_OUT_DIR="api/gen/ts"

echo "Cleaning generated code..."
rm -rf ${GO_OUT_DIR}/*
rm -rf ${TS_OUT_DIR}/*

echo "Generating Go code..."
mkdir -p ${GO_OUT_DIR}

protoc \
  --proto_path=${PROTO_DIR} \
  --proto_path=$(go list -m -f '{{.Dir}}' google.golang.org/protobuf)/.. \
  --proto_path=$(go list -m -f '{{.Dir}}' github.com/grpc-ecosystem/grpc-gateway/v2) \
  --go_out=${GO_OUT_DIR} \
  --go_opt=paths=source_relative \
  --go-grpc_out=${GO_OUT_DIR} \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=${GO_OUT_DIR} \
  --grpc-gateway_opt=paths=source_relative \
  ${PROTO_DIR}/*.proto

echo "Generated Go code in ${GO_OUT_DIR}"

echo "TypeScript generation will be added when web client is set up"
echo "Done!"
