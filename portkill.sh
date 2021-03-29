#!/bin/bash

PORT=$1
if [ "$PORT" ]; then
	lsof -n -i4TCP:$PORT -sTCP:LISTEN -t | xargs kill
	echo "$PORT is dead, long live $PORT!"
else
	echo "Must specify a port! Else I don't know who to kill..."
	exit 1
fi
