#!/bin/sh
qemu-system-x86_64 -M q35 -m 1024M -bios coreboot.rom -serial pipe:/tmp/guest &
echo $! > .pid

