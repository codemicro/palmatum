#!/usr/bin/env bash

set -ex

xcaddy build --with github.com/codemicro/palmatum/caddyZipFs --replace github.com/codemicro/palmatum=$(pwd)
