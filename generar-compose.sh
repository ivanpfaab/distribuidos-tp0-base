#!/bin/bash

# 2 arguments are required for the script to work
if [ $# -ne 2 ]; then
    echo "Error: 2 arguments are required"
    echo "Usage: $0 <output_file> <number_of_clients>"
    echo "Example: $0 docker-compose-dev.yaml 5"
    exit 1
fi

# Assign arguments to variables
OUTPUT_FILE=$1
NUM_CLIENTS=$2

# Check if the number of clients is a number
if ! [[ "$NUM_CLIENTS" =~ ^[0-9]+$ ]]; then
    echo "Error: Number of clients needs to be a number"
    exit 1
fi

# The number of clients needs to be a number higher than 0
if ! (("$NUM_CLIENTS" > 0 )); then
    echo "Error: Number of clients needs to be a number higher than 0"
    exit 1
fi

echo "Generating Docker Compose file: $OUTPUT_FILE"
echo "Number of clients: $NUM_CLIENTS"

# Create the Docker Compose file
cat > "$OUTPUT_FILE" << EOF
name: tp0
services:
  server:
    container_name: server
    image: server:latest
    volumes:
      - ./server/config.ini:/config.ini
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net
EOF

# Add clients
for i in $(seq 1 $NUM_CLIENTS); do
    cat >> "$OUTPUT_FILE" << EOF

  client$i:
    container_name: client$i
    image: client:latest
    volumes:
      - ./client/config.yaml:/config.yaml
    environment:
      - CLI_ID=$i
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
EOF
done

# Add networks section
cat >> "$OUTPUT_FILE" << EOF

networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
EOF

echo "Docker Compose file generated successfully: $OUTPUT_FILE"
echo "The file contains 1 server and $NUM_CLIENTS client(s)"
echo "Configuration files are mounted as volumes"