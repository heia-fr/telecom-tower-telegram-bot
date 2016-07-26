#!/usr/bin/env bash

cd html
find . | cpio -o --format ustar - | gzip > ../web.tar.gz