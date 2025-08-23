#!/bin/bash

# Script to validate echo server functionality using netcat
# This script runs netcat from within the Docker network to test the server

# Test message to send
TEST_MESSAGE="Hello Echo Server Test 123"

# Function to run the test
test_echo_server() {
    # Create a temporary container with netcat and test the server
    # We'll use Alpine Linux (minimal image) and install netcat on-the-fly
    
    RESPONSE=$(
        docker run --rm \
            --network tp0_testing_net \
            alpine \
            sh -c "
                # Install netcat (suppress output)
                apk add --no-cache netcat-openbsd > /dev/null 2>&1 && \
                
                # Send test message and capture response
                echo '$TEST_MESSAGE' | nc server 12345
            "
    )
    
    # Check if the response matches the sent message
    if [ "$RESPONSE" = "$TEST_MESSAGE" ]; then
        echo "action: test_echo_server | result: success"
        return 0
    else
        echo "action: test_echo_server | result: fail"
        # echo "Expected: '$TEST_MESSAGE'"
        # echo "Received: '$RESPONSE'"
        return 1
    fi
}

# Check if the server container is running
if ! docker ps | grep -q "server"; then
    echo "Error: Server container is not running. Please run 'make docker-compose-up' first."
    exit 1
fi

# Run the test
test_echo_server
# Exit code is the return value of the last command executed.
exit_code=$?

exit $exit_code