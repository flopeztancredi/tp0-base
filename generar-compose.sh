#!/bin/bash

# Validate number of arguments
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <output_file> <number_of_clients>"
    exit 1
fi

# Validate that the second argument is a positive integer
if ! [[ "$2" =~ ^[0-9]+$ ]]; then
    echo "Error: number_of_clients must be a positive integer."
    exit 1
fi

# Validate that at least one client is specified
if [ "$2" -lt 1 ]; then
    echo "Error: number_of_clients must be at least 1."
    exit 1
fi

OUTPUT_FILE=$1
N_CLIENTS=$2

# Write server service definition to the output file
cat > $OUTPUT_FILE << 'EOF'
name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net

EOF

# Append client service definitions to the output file
for i in $(seq 1 $N_CLIENTS); do
cat >> $OUTPUT_FILE << EOF
  client$i:
    container_name: client$i
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=$i
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server

EOF
done

# Append network definition to the output file
cat >> $OUTPUT_FILE << 'EOF'
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
EOF