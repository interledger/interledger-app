# Wallet infrastructure

This assumes that this is being run using Hetzner Cloud or Hetzner dedicated server.

## Initial machine config
Use the Hetzner Recovery Image to boot the machine and then ssh into it. We will install Debian
Bookworm on ZFS. 

**Stop and remove software RAID**
```sh
# Find the disk IDs using ls -la /dev/disk/by-id

# See if one or more MD arrays are active:
cat /proc/mdstat
# If so, stop them (replace ``md0`` as required):
mdadm --stop /dev/md0

# For an array using the whole disk:
mdadm --zero-superblock --force $DISK # (remember to do for each disk)
```

Copy the `debian-12-on-zfs.sh` script onto the host and run it.

**FIREWALL**
Configure the Hetzner firewall to allow http, https and openSSH.

## Install the hashistack
You will need AWS security credentials for the following IAM users:
 - `hetzner-backup` to read and write to the S3 bucket `s3://fynbos-wallet`
 - `vault-hetzner` to use the KMS key `7ce2e300-e73c-4069-9da6-867d7bd7767a` to unseal Vault

The script will install the hashistack and create the config files. It *won't* start the services as
you may want to restore a snapshot from s3.

Run the following commands from your local machine:
```sh
# AWS credentials for vault-hetzner
export KMS_AWS_REGION="..."
export KMS_AWS_ACCESS_KEY_ID="..."
export KMS_AWS_SECRET_ACCESS_KEY="..."

# AWS credentials for hetzner-backup
export S3_AWS_REGION="..."
export S3_AWS_ACCESS_KEY_ID="..."
export S3_AWS_SECRET_ACCESS_KEY="..."

# HOST details
export HOST_IP="..."

ssh root@$HOST_IP "
HISTFILE=/dev/null \
KMS_AWS_REGION='$KMS_AWS_REGION' \
KMS_AWS_ACCESS_KEY_ID='$KMS_AWS_ACCESS_KEY_ID' \
KMS_AWS_SECRET_ACCESS_KEY='$KMS_AWS_SECRET_ACCESS_KEY' \
S3_AWS_REGION='$S3_AWS_REGION' \
S3_AWS_ACCESS_KEY_ID='$S3_AWS_ACCESS_KEY_ID' \
S3_AWS_SECRET_ACCESS_KEY='$S3_AWS_SECRET_ACCESS_KEY' \
HOST_IP='$HOST_IP' \
bash -s" < install.sh
```

## Restore a snapshot from ZFS
You will need AWS credentials for the `hetzner-backup` IAM user to read from the S3 bucket `s3://fynbos-wallet`.
```sh
export AWS_REGION="..."
export AWS_ACCESS_KEY_ID="..."
export AWS_SECRET_ACCESS_KEY="..."

aws s3 cp "s3://fynbos-wallet/full/${snapshot_name}" - | zfs receive /backup/live
zfs set mountpoint=/data backup/live
```

Remember to start vault, nomad and consul.

## Bootstrapping hashistack
`bootstrap.sh` does not need to be run if you are restoring from a snapshot. This was used for the
initial configuration.


