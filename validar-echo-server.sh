#!/bin/bash

MESSAGE="hello world"
RESPONSE=$(docker run --rm --network tp0_testing_net alpine:latest sh -c "echo '$MESSAGE' | nc -w 5 server 12345")

if [ "$RESPONSE" = "$MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
    exit 1
fi