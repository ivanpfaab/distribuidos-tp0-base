#!/bin/bash

# Script to validate echo server functionality using netcat
# This script runs netcat from within the Docker network to test the server

# Test message to send
TEST_MESSAGE="Hello Echo Server Test 123"

# Function to run the test
test_echo_server() {
    # echo "=== Starting Echo Server Test ==="
    
    # Step 1: Create container
    # We use tail -f /dev/null to keep the container running.
    # echo "1. Creating Alpine container..."
    CONTAINER_ID=$(docker run -d \
        --network tp0_testing_net \
        alpine \
        sh -c "tail -f /dev/null")
    
    # Step 2: Install netcat
    # echo "2. Installing netcat..."
    docker exec "$CONTAINER_ID" sh -c "apk add --no-cache netcat-openbsd > /dev/null 2>&1"
    
    # Step 3: Test connection
    # echo "3. Testing server connection..."
    RESPONSE=$(docker exec "$CONTAINER_ID" sh -c "echo '$TEST_MESSAGE' | nc server 12345")
    
    # Step 4: Clean up
    # echo "4. Cleaning up..."
    docker stop "$CONTAINER_ID" > /dev/null 2>&1 # Suppress output
    docker rm "$CONTAINER_ID" > /dev/null 2>&1 # Suppress output
    
    # Step 5: Validate response
    # echo "5. Validating response..."
    if [ "$RESPONSE" = "$TEST_MESSAGE" ]; then
        echo "action: test_echo_server | result: success"
        return 0
    else
        #echo "Test FAILED"
        #echo "Expected: '$TEST_MESSAGE'"
        #echo "Received: '$RESPONSE'"
        echo "action: test_echo_server | result: fail"
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