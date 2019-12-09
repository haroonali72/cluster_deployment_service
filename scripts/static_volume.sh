#!/bin/sh
  sudo /sbin/wipefs /dev/sdd
  sudo /sbin/mkfs.ext4 /dev/sdd
  sudo mkdir /data1
  sudo mount /dev/sdd /data1
  sudo /bin/chmod 777 /dev/sdd
  sudo su <<-EOF
  sudo foo=blkid | awk '{print $2}' | sed -n '3p'
  echo '${foo} /data1 ext4 defaults,nofail 1 2'
    EOF