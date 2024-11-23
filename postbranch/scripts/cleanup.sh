#!/bin/bash

# Step 1: Destroy all zpools
for pool in $(zpool list -Ho name); do
  echo "Destroying zpool: $pool"
  sudo zpool destroy "$pool"
done

# Step 2: Remove all loopback devices associated with /var/lib/post-branch/virtualdisk01.img
for loop_dev in $(losetup -a | grep "post-branch" | awk -F: '{print $1}'); do
  echo "Detaching and removing loopback device: $loop_dev"
  sudo losetup -d "$loop_dev" && sudo rm "$loop_dev"
done

# Step 3: Remove all files in /var/lib/post-branch/
echo "Removing all files from /var/lib/post-branch/"
sudo rm -rf /var/lib/post-branch/*

# Step 4: Remove all files in /mnt that match "pb" in their names
echo "Removing all files in /mnt matching 'pb'"
for file in /mnt/*pb*; do
  if [ -e "$file" ]; then
    echo "Removing file: $file"
    sudo rm -rf "$file"
  fi
done

echo "Removing all files in /var/run/postbranch"
for file in /var/run/postbranch/*; do
  if [ -e "$file" ]; then
    echo "Removing file: $file"
    sudo rm -rf "$file"
  fi
done

echo "Cleanup complete."
