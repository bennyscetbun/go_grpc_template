#!/bin/bash

set -e

cd "$(dirname "$0")"
. ./library.sh

function generate_database() {
    rm -rf ../backend/generated/database
    (cd ../backend && bash ../scripts/run_against_psql.sh ./cmd/database_gen)
}

function generate_proto() {
    rm -rf ../frontend/src/generated/rpc
    rm -rf ../backend/generated/rpc
    list_proto_files=$(cd ../common/proto ; find . -name '*.proto')
    (cd .. && mkdir -p frontend/src/generated/rpc/ backend/generated/rpc/ && docker run --user $(id -u):$(id -g) --rm -v./common/proto:/proto -w/proto -v./backend/generated/rpc:/gogen -v./frontend/src/generated/rpc:/jsproto xxxyourappyyy/protoc -I=. $list_proto_files   --js_out=import_style=commonjs:/jsproto   --grpc-web_out=import_style=typescript,mode=grpcweb:/jsproto --go_out=/gogen --go_opt=paths=source_relative     --go-grpc_out=/gogen --go-grpc_opt=paths=source_relative)
}

build_docker_images
run_if_change ../backend/resources/database generate_database
run_if_change ../common/proto generate_proto
