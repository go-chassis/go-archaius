#!/bin/sh
set -xe

mkdir -p $GOPATH/src/github.com/apache
cd $GOPATH/src/github.com/apache
git clone https://github.com/apache/servicecomb-kie.git
cd $GOPATH/src/github.com/apache/servicecomb-kie/build
bash build_docker.sh
sudo docker-compose -f $GOPATH/src/github.com/apache/servicecomb-kie/deployments/docker/docker-compose.yaml down
sudo docker-compose -f $GOPATH/src/github.com/apache/servicecomb-kie/deployments/docker/docker-compose.yaml up -d

cd $GOPATH/src/github.com/go-chassis/go-archaius
go test ./... -v -covermode=count -coverprofile=coverage.out
$HOME/gopath/bin/goveralls -coverprofile=coverage.out -service=travis-ci