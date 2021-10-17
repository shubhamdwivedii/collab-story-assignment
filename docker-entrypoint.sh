#!/bin/sh 

echo "Waiting for DB to start..."

cd collab 

./wait-for database:8080 -- echo "Database Has Started..."
# https://github.com/eficode/wait-for

# Make sure server binary is copies onto prodution container /collab 
echo "Running Server..."
./server 