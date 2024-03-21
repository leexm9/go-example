#!/usr/bin/env bash

xOS="linux"
if [[ $OSTYPE == darwin* ]]; then
  xOS="darwin"
fi

PROGRAM="demo"
PROGRAM_BUILD_OUTPUT_DIR="bin/$xOS/$PROGRAM"
PROGRAM_EXE="bin/$xOS/$PROGRAM"

printf "Building $PROGRAM...\n"
echo

CGO_ENABLED=1 GOARCH=amd64 go build -trimpath -o "$PROGRAM_BUILD_OUTPUT_DIR"
[[ $? -ne 0 ]] && exit 1

printf "Starting $PROGRAM...\n\n"
"$PROGRAM_EXE" --pluginDir="bin/$xOS/modules" --pidFile="bin/$xOS/$PROGRAM.pid"