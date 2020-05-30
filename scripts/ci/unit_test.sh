#!/bin/sh
set -xe

go test $(go list ./... | grep -v /pkg/kieclient | grep -v /source/remote/kie) -cover