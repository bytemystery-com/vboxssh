#!/bin/bash

# CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H=windowsgui" .
CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 fyne package --release
# CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" .
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 fyne package --release
mv VBoxSsh.tar.xz dist/linux
mv VBoxSsh.exe dist/windows

