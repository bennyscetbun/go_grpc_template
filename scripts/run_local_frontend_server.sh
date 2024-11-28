#!/bin/bash

cd "$(dirname "$0")"
. ./library.sh

trap cleanup 1 2 3 6
container_name="xxx_your_app_xxx.live_server"
node_image="xxx_your_app_xxx/node"
cleanup()
{
  echo "Closing docker"
  docker stop "$container_name"
  exit
}

bash ./bundle_front_end.sh

cd ../frontend && docker run --name "$container_name" --user $(id -u):$(id -g) -p3535:8080 --rm -w/app -v./:/app  --entrypoint /bin/sh "xxx_your_app_xxx/node" -c "npx webpack serve"