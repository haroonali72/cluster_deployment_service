#!/bin/sh
  sudo /sbin/wipefs /dev/xvdf
  sudo /sbin/mkfs.ext4 /dev/xvdf
  sudo mkdir /data0
  sudo mount /dev/xvdf /data0
  sudo /bin/chmod 777 /dev/xvdf
  sudo su <<-EOF
  echo '/dev/xvdf /data0 ext4 defaults 0 0' >> /etc/fstab
    EOF