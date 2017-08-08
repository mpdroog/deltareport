#!/bin/bash
# Compile binary for GNU/Linux
set -e
set -x
set -u

env GOOS=linux GOARCH=amd64 go build
# Set ELF ABI to 003 to run under FreeBSD
echo -n $'\003' | dd bs=1 count=1 seek=7 conv=notrunc of=./deltareport
