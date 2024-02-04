#!/bin/env bash
go build -v -o /tmp/kcalc ./cmd/server
rsync --checksum --compress --update /tmp/kcalc root@mullareros.com:/usr/bin/kcalc
ssh root@mullareros.com systemctl restart kcalc