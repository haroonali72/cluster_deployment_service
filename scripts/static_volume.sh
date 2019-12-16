#!/bin/sh
  sudo /sbin/wipefs /dev/sdc
  sudo /sbin/mkfs.ext4 /dev/sdc
  sudo mkdir /data1
  sudo mount /dev/sdc /data1
  sudo /bin/chmod 777 /dev/sdc
  sudo su <<-EOF
  sudo foo=blkid | awk '{print $2}' | sed -n '3p'
  echo '${foo} /data1 ext4 defaults,nofail 1 2'
    EOF