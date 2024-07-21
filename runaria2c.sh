#!/bin/bash

aria2c --enable-rpc --rpc-listen-all \
        --max-concurrent-downloads=10 --max-connection-per-server=5 \
        --max-tries=0 --retry-wait=5 \
        --continue=true --auto-file-renaming=false