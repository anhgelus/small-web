#!/usr/bin/bash

rsync -riz ./{public,data,config.toml,*.sh} vps:~/anhgelus.world/
