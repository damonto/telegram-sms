#!/bin/bash

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o build/linux-amd64-sms *.go
chmod +x build/linux-amd64-sms

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o build/linux-arm64-sms *.go
chmod +x build/linux-arm64-sms
