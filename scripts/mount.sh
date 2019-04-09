#!/bin/sh
  sudo /sbin/wipefs xvdg
  sudo /sbin/mkfs.ext4 xvdg
  sudo mkdir data0
  sudo mount xvdg /data0
  sudo /bin/chmod 777 xvdg

  echo 'xvdg /data0 ext4 defaults 0 0' >> /etc/fstab
  EOF
