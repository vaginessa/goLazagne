#!/bin/bash

# Find the PID of the 'dlv' process
pid=$(ps | grep dlv | awk '{print $1}')

# Check if the pid variable is empty or not
if [ -n "$pid" ]; then
    # Kill the process with SIGKILL signal
    kill -9 "$pid"
    echo "Process $pid has been killed"
else
    echo "Process not found"
fi
