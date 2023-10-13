#!/bin/sh

file="./gorrc_client"

if [ -f "$file" ] ; then
    rm "$file"
fi

echo Compiling...
go build .

export $(grep -v '^#' real.env | xargs)
export XDG_RUNTIME_DIR=""

sudo -E ./gorrc_client