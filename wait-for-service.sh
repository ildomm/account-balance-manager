#!/bin/bash

HOST=$1
PORT=$2
TIMEOUT=${3:-30} # Default timeout is 30 seconds

echo "Waiting for $HOST:$PORT to be available..."

start_time=$(date +%s)
while ! nc -z $HOST $PORT; do
  sleep 1
  echo -n "."
  current_time=$(date +%s)
  if (( current_time - start_time >= TIMEOUT )); then
    echo
    echo "Error: Timeout while waiting for $HOST:$PORT to be available"
    exit 1
  fi
done

echo
echo "$HOST:$PORT is available!"
