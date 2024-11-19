#!/bin/bash

# Start the backend server (Go)
echo "Starting the backend server..."
cd back || exit
go run main.go &
BACK_PID=$!

# Start the frontend server (Python3)
echo "Starting the frontend server..."
cd ../front || exit
python3 -m http.server 8000 &
FRONT_PID=$!

# Display the address to open in the browser
echo "The project is running at: http://localhost:8000"
echo "Press [Enter] to stop the servers and exit."

# Wait for user input to stop the servers
read

# Kill the background processes
kill $BACK_PID
kill $FRONT_PID

echo "Servers stopped. Exiting."
