#!/bin/bash
  DRIVE=""
  SAME_SIZE_DRIVES=$(lsblk | awk ' $4 == "VOLUMEG" {print $1}')
  arr=($SAME_SIZE_DRIVES)
  if [ ${#arr[@]} -eq 1 ]
  then
     DRIVE=${arr[0]}
  else
     for var in "${arr[@]}"
     do
        MOUNT_VAL=$(lsblk | awk ' $1 == "'$var'" {print $7}')
        if [ -z "$MOUNT_VAL"]
        then
            DRIVE=$var
            break
        fi
     done
  fi
  sudo /sbin/wipefs /dev/${DRIVE}
  sudo /sbin/mkfs.ext4 /dev/${DRIVE}
  sudo mkdir /data0
  sudo mount /dev/${DRIVE} /data0
  sudo /bin/chmod 777 /dev/${DRIVE}
  sudo su <<-EOF
  sudo foo=blkid | awk '{print $2}' | sed -n '3p'
  echo '${foo} /data0 ext4 defaults,nofail 1 2'
    EOF