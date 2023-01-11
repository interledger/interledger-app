#!/bin/sh

set -e

WATCH_FILES=$1
POLL_INTERVAL_SECONDS=$2
SIGHUP_PROCESS=$3
CHECKSUM_FILE=$4/checksums.txt


echo $(date -u) - Watching $WATCH_FILES every $POLL_INTERVAL_SECONDS seconds. Will SIGHUP $SIGHUP_PROCESS. Storing checksums in $CHECKSUM_FILE

md5sum $WATCH_FILES > $CHECKSUM_FILE

while true; do
	if !(md5sum -c $CHECKSUM_FILE -s); then
		echo "$(date -u) - SIGHUP $SIGHUP_PROCESS"
		pkill -SIGHUP $SIGHUP_PROCESS

		echo "$(date -u) - Generating new checksums..."
  		md5sum $WATCH_FILES > $CHECKSUM_FILE
	fi
  	sleep $POLL_INTERVAL_SECONDS
done
