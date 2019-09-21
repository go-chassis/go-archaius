#!/bin/sh
set -e

cd $GOPATH/src/github.com/go-chassis/go-archaius
go test ./... -v -covermode=count -coverprofile=coverage.out
$HOME/gopath/bin/goveralls -coverprofile=coverage.out -service=travis-ci