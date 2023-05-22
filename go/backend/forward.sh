#!/bin/bash

# Define a function to run when SIGINT is received
trap "echo 'Script interrupted. Exiting.'; exit" INT

while true; do
  kubectl -n $1 port-forward $2 $3:$4
  sleep 1
done