#!/bin/bash

# Auto Shutdown Microservice test script
# Replace these variables with your actual values
MICROSERVICE_URL="your-microservice-url"
ROOM_NAME="your-room-name"

echo "Testing Auto Shutdown Microservice..."
echo "Microservice URL: $MICROSERVICE_URL"
echo "Room Name: $ROOM_NAME"
echo "----------------------------------------"

# GET Time Avoidance
echo "Testing GET Time Avoidance..."
curl -X GET "http://${MICROSERVICE_URL}/${ROOM_NAME}/time_avoidance?from=0730&to=2000"
sleep 1

# GET Occupancy Detected
echo "Testing GET Occupancy Detected..."
curl -X GET "http://${MICROSERVICE_URL}/${ROOM_NAME}/occupancy_detected?last_x_minutes=180"
sleep 1

# SET Occupancy Detected
echo "Testing SET Occupancy Detected..."
curl -X PUT "http://${MICROSERVICE_URL}/${ROOM_NAME}/occupancy_detected" \
     -H "Content-Type: application/json" \
     -d '"true"'
sleep 1

echo "----------------------------------------"
echo "All tests completed."