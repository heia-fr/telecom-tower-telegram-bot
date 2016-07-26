#!/usr/bin/env bash

export GOPATH=/vagrant/_build/go
APP_NAME=telecom-tower-telegram-bot
BASE_DIR=github.com/heia-fr/${APP_NAME}

cd /vagrant

mkdir -p ${GOPATH}/pkg ${GOPATH}/bin
mkdir -p ${GOPATH}/src/${BASE_DIR}
rsync --progress /vagrant/*.go ${GOPATH}/src/${BASE_DIR}/

cd ${GOPATH}/src/${BASE_DIR}

go get .
go install

cp $GOPATH/bin/${APP_NAME} /vagrant/${APP_NAME}.linux-release