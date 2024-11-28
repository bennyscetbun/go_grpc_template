#!/bin/bash

set -e

cd "$(dirname "$0")"
. library.sh

bash generate_files.sh
bash bundle_front_end.sh

function build_backend() {
    (cd ../backend && go build ./...)
}

run_if_change ../backend build_backend