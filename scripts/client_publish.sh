#!/bin/bash

cd ~/m4k_client
./publish.sh

cd ~/m4k_server/scripts
./caddy_restart.sh