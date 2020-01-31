#!/bin/sh
if [ -f ".pid" ]; then
	kill -9 $(cat .pid)
	rm -f .pid
fi
