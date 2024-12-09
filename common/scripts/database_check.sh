#!/bin/bash

set -e

cd "$(dirname "$0")"
. ./library.sh

cd "$LIBRARY_SH_DIR../../backend" && bash "$LIBRARY_SH_DIR/run_against_psql.sh" ./cmd/database_check
