#!/bin/bash

cd "$(dirname "$0")"
. ./library.sh

trap cleanup 1 2 3 6
container_name="xxxyourappyyy.live_server"
node_image="xxxyourappyyy/node"
cleanup()
{
  echo "Closing docker"
  docker stop "$container_name"
  exit
}

bash ./bundle_front_end.sh

cd ../frontend && docker run --name "$container_name" --user $(id -u):$(id -g) -p3535:8080 --rm -w/app -v./:/app  --entrypoint /bin/sh "xxxyourappyyy/node" -c "npx webpack serve"