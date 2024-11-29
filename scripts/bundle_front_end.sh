#!/bin/bash

node_image="xxxyourappyyy/node"

cd "$(dirname "$0")"
. ./library.sh

current_path=$(pwd)

function renew_node_modules() {
    cd ../frontend
    rm -rf node_modules && mkdir -p node_modules
    docker run --user $(id -u):$(id -g) --rm -w/app -v./:/app --entrypoint /bin/sh "$node_image" -c "npm install"
    ret=$?
    cd - >/dev/null
    return $ret
}

function build_bundle() {
    cd ../frontend
    rm -rf dist && mkdir -p dist
    docker run --user $(id -u):$(id -g) --rm -w/app -v./:/app  --entrypoint /bin/sh "$node_image" -c "npx webpack"
    ret=$?
    cd - >/dev/null
    return $ret
}

build_docker_images
run_if_change ../frontend/package.json renew_node_modules
run_if_change ../frontend/src build_bundle
