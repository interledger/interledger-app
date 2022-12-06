#!/bin/sh

set -e

echo $(date -u) - Watching $1 every $2 seconds. Will SIGHUP $3. Storing checksums in $4/checksums.txt

md5sum $1 > $4/checksums.txt

while true; do
	if !(md5sum -c $4/checksums.txt -s); then
		echo "$(date -u) - SIGHUP $3"
		pkill -SIGHUP -x $3

		echo "$(date -u) - Generating new checksums..."
  		md5sum $1 > $4/checksums.txts
	fi
  	sleep $2
done
