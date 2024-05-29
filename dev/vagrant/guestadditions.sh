#!/bin/bash

sudo apt-get -y install linux-headers-$(uname -r) build-essential dkms

wget https://download.virtualbox.org/virtualbox/7.0.14/VBoxGuestAdditions_7.0.14.iso
sudo mkdir /media/VBoxGuestAdditions
sudo mount -o loop,ro VBoxGuestAdditions_7.0.14.iso /media/VBoxGuestAdditions
sudo sh /media/VBoxGuestAdditions/VBoxLinuxAdditions.run
rm VBoxGuestAdditions_7.0.14.iso
sudo umount /media/VBoxGuestAdditions
sudo rmdir /media/VBoxGuestAdditions
