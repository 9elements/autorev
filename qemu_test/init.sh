#!/bin/sh
if [ ! -e /tmp/guest.in ]; then
	mkfifo /tmp/guest.in
fi
if [ ! -e /tmp/guest.out ]; then
	mkfifo /tmp/guest.out
fi
