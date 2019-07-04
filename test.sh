#!/usr/bin/env bash

echo "Running with environment:"
env

sleep 1
echo "Test running"
sleep 2
echo "This is an error" >&2
sleep 7
echo "Bye!"
