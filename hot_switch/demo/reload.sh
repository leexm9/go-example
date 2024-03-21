#!/usr/bin/env bash

xOS="linux"
if [[ $OSTYPE == darwin* ]]; then
  xOS="darwin"
fi

PROGRAM="demo"
echo "Sending signal..."
kill -USR1 `cat "bin/$xOS/$PROGRAM.pid"`