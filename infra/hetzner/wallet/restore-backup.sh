#!/bin/bash

SNAPSHOT=$(restic list snapshots | head -n 1)

echo Downloading snapshot $SNAPSHOT...
restic restore --target=/ $SNAPSHOT

echo Restoring snapshot $SNAPSHOT...
cat /tmp/$SNAPHOT | zfs receive backup/live

echo Done
