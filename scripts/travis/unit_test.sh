#!/bin/sh
set -e

mkdir -p $GOPATH/src/github.com/fsnotify
mkdir -p $GOPATH/src/github.com/spf13
mkdir -p $GOPATH/src/github.com/gorilla
mkdir -p $GOPATH/src/golang.org/x/
mkdir -p $GOPATH/src/github.com/stretchr
go get gopkg.in/yaml.v2

cd $GOPATH/src/github.com/ServiceComb
git clone https://github.com/ServiceComb/go-chassis.git
git clone https://github.com/ServiceComb/paas-lager.git
git clone https://github.com/ServiceComb/go-cc-client.git

cd $GOPATH/src/github.com/fsnotify
git clone https://github.com/fsnotify/fsnotify.git
cd fsnotify
git reset --hard 629574ca2a5df945712d3079857300b5e4da0236

cd $GOPATH/src/github.com/spf13
git clone https://github.com/spf13/cast.git
cd cast
git reset --hard acbeb36b902d72a7a4c18e8f3241075e7ab763e4

cd $GOPATH/src/github.com/gorilla
git clone https://github.com/gorilla/websocket.git
cd websocket
git reset --hard 1f512fc3f05332ba7117626cdfb4e07474e58e60

cd $GOPATH/src/golang.org/x/
git clone https://github.com/golang/sys.git

cd $GOPATH/src/github.com/stretchr
git clone https://github.com/stretchr/testify.git


cd $GOPATH/src/github.com/ServiceComb/go-archaius
#Start unit test
for d in $(go list ./... | grep -v configcenter-source); do
    echo $d
    echo $GOPATH
    cd $GOPATH/src/$d
    if [ $(ls | grep _test.go | wc -l) -gt 0 ]; then
        go test 
    fi
done
