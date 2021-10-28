#!/usr/bin/env bash

set -e

localport=26257
typename=service/cockroachdb
remoteport=26257

# This would show that the port is closed
# nmap -sT -p $localport localhost || true
echo "Port forwarding to cockroachdb"
kubectl port-forward $typename $localport:$remoteport > /dev/null 2>&1 &

pid=$!
# echo pid: $pid

# kill the port-forward regardless of how this script exits
trap '{
    echo "Killing port forward"
    kill $pid
}' EXIT

# wait for $localport to become available
while ! nc -vz localhost $localport > /dev/null 2>&1 ; do
    # echo sleeping
    sleep 0.1
done

echo "Port forwarding successful"
echo "Running dbinit.sql"

# This would show that the port is open
# nmap -sT -p $localport localhost

# Actually use that port for something useful - here making a backup of the
# keycloak database
cockroach sql --certs-dir=certs --host=localhost:26257 -f dbinit.sql -u root

echo "Complete dbinit.sql"
# the 'trap ... EXIT' above will take care of kill $pid