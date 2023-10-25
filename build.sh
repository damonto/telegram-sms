#!/bin/bash

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-w -s" -o build/linux-amd64-telegram-sms *.go
chmod +x build/linux-amd64-telegram-sms

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-w -s" -o build/linux-arm64-telegram-sms *.go
chmod +x build/linux-arm64-telegram-sms
