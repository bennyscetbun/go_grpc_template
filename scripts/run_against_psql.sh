#!/bin/bash

# Determine the operating system (macOS or Linux)
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    path_hash=$(echo -n "$current_path $0" | shasum -a 256 | awk '{print $1}')
else
    # Linux (and other Unix-like systems)
    path_hash=$(echo -n "$current_path $0" | sha256sum | awk '{print $1}')
fi

# Define the name for your PostgreSQL container
container_name="postgres.$path_hash"

postgres_image="postgres:16.0"
db_password="${DBPASSWD:-password}"

# Check if the container is already running
if [[ "$(docker inspect -f '{{.State.Running}}' "$container_name" 2>/dev/null)" == "true" ]]; then
    echo "Container $container_name is already running."
else
    # Check if the container exists (but is not running)
    if [[ "$(docker ps -a -q -f name=$container_name 2>/dev/null)" != "" ]]; then
        # Start the existing container
        docker start "$container_name"
        echo "Container $container_name started."
    else
        # Create and start a new container using the "postgres" image
        docker run --name "$container_name" -p 3000-4000:5432 -e POSTGRES_PASSWORD="$db_password" -d "$postgres_image"
        echo "Container $container_name created and started."
    fi
fi



database=`date +$(basename "$1")"_%Y_%m_%d_%H_%M_%S"`
attempts=0
max_attempts=10

while [ $attempts -lt $max_attempts ]; do
    if docker exec "$container_name" /bin/bash -c "psql -U postgres -c 'CREATE DATABASE $database'"; then
        break
    fi
    echo retrying in 1 second
    sleep 1
    ((attempts++))
done

if  [ $attempts -ge $max_attempts ]; then
    echo "psql server not running"
    exit 1
fi

pwd
DBPASSWD="$db_password" DBNAME="$database" DBPORT=`docker port "$container_name" 5432 | cut -f 2 -d ':'` go run "$@"
result=$?
docker stop "$container_name" > /dev/null
docker rm "$container_name" > /dev/null
exit $result