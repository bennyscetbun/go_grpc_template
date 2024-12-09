#!/bin/bash

set -e

cd "$(dirname "$0")"
. ./library.sh
ROOT_DIR="$LIBRARY_SH_DIR/../.."
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"

function generate_database() {
    rm -rf "$BACKEND_DIR/generated/database"
    (cd "$BACKEND_DIR" && bash "$LIBRARY_SH_DIR/run_against_psql.sh" ./cmd/database_gen)
}

function generate_proto() {
    rm -rf "$FRONTEND_DIR/src/generated/rpc"
    rm -rf "$BACKEND_DIR/generated/rpc"
    list_proto_files=$(cd "$LIBRARY_SH_DIR/../proto" ; find . -name '*.proto')
    (cd "$ROOT_DIR" && mkdir -p frontend/src/generated/rpc/ backend/generated/rpc/ && docker run --user $(id -u):$(id -g) --rm -v./common/proto:/proto -w/proto -v./backend/generated/rpc:/gogen -v./frontend/src/generated/rpc:/jsproto xxxyourappyyy/protoc -I=. $list_proto_files   --js_out=import_style=commonjs:/jsproto   --grpc-web_out=import_style=typescript,mode=grpcweb:/jsproto --go_out=/gogen --go_opt=paths=source_relative     --go-grpc_out=/gogen --go-grpc_opt=paths=source_relative) && echo code generated from protofiles
}

build_docker_images
run_if_change "$BACKEND_DIR/resources/database" generate_database
run_if_change "$LIBRARY_SH_DIR/../proto" generate_proto
