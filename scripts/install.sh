#!/bin/bash

cd "$(dirname "$0")"

# Define the target Go version
target_version="go1.23"  # Change this to your desired version

# Get the installed Go version
go_version=$(go version | awk '{print $3}')

# Compare the installed Go version with the target version
if [[ "$(printf '%s\n' "$target_version" "$go_version" | sort -V | head -n 1)" == "$target_version" ]]; then
  echo "Installed Go version ($go_version) is above or equal to $target_version."
else
  echo "Installed Go version ($go_version) is below $target_version. EXITING"
  exit 1
fi

go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest