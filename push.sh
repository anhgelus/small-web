#!/usr/bin/bash

rsync -rvz ./{public,data,config.toml,*.sh} vps:~/small-web-data/
