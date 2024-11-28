#!/bin/bash

set -e

cd "$(dirname "$0")"
cd ../backend && bash ../scripts/run_against_psql.sh ./cmd/database_check
