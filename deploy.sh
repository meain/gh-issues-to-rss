#!/bin/sh

set -xe

GOOS=linux GOARCH=amd64 go build
scp index.html doom:/home/ubuntu/ # copy index file
scp gh-issues-to-rss doom:/tmp/g-temp
# cannot scp directly to the file when the service is running
ssh doom "mv /tmp/g-temp /home/ubuntu/gh-issues-to-rss && sudo systemctl restart gh-issues-to-rss"
rm gh-issues-to-rss
