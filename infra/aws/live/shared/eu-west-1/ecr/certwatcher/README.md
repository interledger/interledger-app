Creates checksums of all the files in the specified directories. Validates these checksums on an interval.
Sends a SIGHUP to the specified process if the checksums aren't valid.

NOTE: uses an interval to avoid issues with inotify events.
https://ahmet.im/blog/kubernetes-inotify/
