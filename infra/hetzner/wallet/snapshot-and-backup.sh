#!/bin/bash

trap 'error_handler $? $LINENO' ERR

function error_handler () {
	echo "Snapshot error: ($1) occurred on $2"

	exit
}

SNAPSHOT_NAME=`date +%d_%m_%Y_%H_%M_%S`

echo ZFS taking snapshot $SNAPSHOT_NAME...

zfs snapshot -r backup/live@$SNAPSHOT_NAME
zfs send backup/live@$SNAPSHOT_NAME > /tmp/$SNAPSHOT_NAME.img

echo Restic backing up snapshot $SNAPSHOT_NAME...

restic backup --tag $SNAPSHOT_NAME /tmp/$SNAPSHOT_NAME.img
