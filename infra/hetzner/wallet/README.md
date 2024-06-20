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

## Unattended Security Updates
This uses the `unattended-upgrades` package to check for and install package updates on a schedule.
Any failures to do so will be emailed to `engineering@fynbos.dev`.

*Configure unattended-upgrades*
Update the following lines:
```sh
# /etc/apt/apt.conf.d/50unattended-upgrades
Unattended-Upgrade::Mail "root";
Unattended-Upgrade::MailReport "on-change";
Unattended-Upgrade::Remove-Unused-Kernel-Packages "true";
Unattended-Upgrade::Remove-New-Unused-Dependencies "true";
Unattended-Upgrade::Remove-Unused-Dependencies "true";
Unattended-Upgrade::Automatic-Reboot "false";

Unattended-Upgrade::Package-Blacklist {
    // We will manage these manually
    "consul";
    "vault";
    "nomad";
};
```
NB! run `sudo systemctl status unattended-upgrades` and check it is enabled and active.

*Update systemd timers*
```sh
# run sudo systemctl edit apt-daily.timer
[Unit]
Description=Daily apt download activities

[Timer]
OnCalendar=*-*-* 01:00
RandomizedDelaySec=12h
Persistent=true

[Install]
WantedBy=timers.target

# run sudo systemctl edit apt-daily-upgrade.timer
[Unit]
Description=Daily apt upgrade and clean activities
After=apt-daily.timer

[Timer]
OnCalendar=*-*-* 02:00
RandomizedDelaySec=60m
Persistent=true

[Install]
WantedBy=timers.target
```

*Configure exim4 with SendGrid*
```sh
# /etc/exim4/update-exim4.conf.conf

dc_eximconfig_configtype='smarthost'
dc_other_hostnames='fynbos.app'
dc_local_interfaces='127.0.0.1'
dc_readhost='fynbos.app'
dc_relay_domains=''
dc_minimaldns='false'
dc_relay_nets=''
dc_smarthost='smtp.sendgrid.net::587'
CFILEMODE='644'
dc_use_split_config='false'
dc_hide_mailname='true'
dc_mailname_in_oh='true'
dc_localdelivery='mail_spool'

# /etc/exim4/passwd.client
*:apikey:<SendGrid apikey> # this apikey must only be allowed to use the MailSend api
```
NB! run `sudo systemctl restart exim4`

*Configure alias for root*
```sh
# /etc/aliases
mailer-daemon: postmaster
postmaster: root
nobody: root
hostmaster: root
usenet: root
news: root
webmaster: root
www: root
ftp: root
abuse: root
noc: root
security: root
root: engineering@fynbos.dev
```
NB! run `sudo newaliases`

Test everything is working as expected with `sudo unattended-upgrades -d`. Note that you may not
get an email if no packages were installed. you can set `Unattended-Upgrade::MailReport "always";` 
temporarily to test the email notifications.

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


