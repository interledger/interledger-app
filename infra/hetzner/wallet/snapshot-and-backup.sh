#!/bin/bash

trap 'error_handler $? $LINENO' ERR

function error_handler () {
        echo "Snapshot error: ($1) occurred on $2"

        exit
}

SNAPSHOT_NAME=`date +%d_%m_%Y_%H_%M_%S`

echo ZFS taking full snapshot $SNAPSHOT_NAME...

zfs snapshot -r backup/live@$SNAPSHOT_NAME
expected_size=`zfs send -c -v "backup/live@$SNAPSHOT_NAME" -n | grep "total estimated size is" | awk '{print $5}' | sed 's/\.[0-9]*//g' | sed 's/G/000000000/' | sed 's/M/000000/'`
expected_size=`echo "($expected_size * 1.3) / 1" | bc`
zfs send backup/live@$SNAPSHOT_NAME > /tmp/$SNAPSHOT_NAME.img

echo Streaming full snapshot $SNAPSHOT_NAME...

zfs send -c -v "backup/live@$SNAPSHOT_NAME" | aws s3 cp --region eu-central-1 --expected-size="${expected_size}" - "s3://fynbos-wallet/full/$SNAPSHOT_NAME"

echo Cleaning up snapshot...
zfs destroy "backup/live@$SNAPSHOT_NAME" || true

echo Done.
