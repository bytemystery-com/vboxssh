#!/bin/bash
# Copyright (c) 2026 Reiner Pröls
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.
#
# SPDX-License-Identifier: MIT

# CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H=windowsgui" .
# CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" .
# CGO_ENABLED=1 GOOS=android GOARCH=arm64 CC=$CC CXX=$CXX fyne package --os android --release --archs arm64,armv7 --metadata buildts="${ts}"

export CC_WIN=x86_64-w64-mingw32-gcc
export CXX_WIN=x86_64-w64-mingw32-g++

ANDROID_KEY=~/Dropbox/private/android_key/android_key.txt

go clean -cache -modcache

ts=$(date -u +'%Y-%m-%d - %H:%M:%S')
fyne translate assets/lang/xx.json
echo "Build windows ..."
CGO_ENABLED=1 CXX=${CXX_WIN} CC=${CC_WIN} GOOS=windows GOARCH=amd64 fyne package --release --metadata buildts="${ts}"
mkdir -p dist/windows
mv VBoxSsh.exe dist/windows

echo "Build linux ..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 fyne package --release --metadata buildts="${ts}"
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 fyne build --release --metadata buildts="${ts}"
mkdir -p dist/linux
mv vboxssh dist/linux
mv VBoxSsh.tar.xz dist/linux

# Android stuff
echo "Build android ..."
export ANDROID_HOME=${HOME}/Android
export ANDROID_NDK_HOME=${ANDROID_HOME}/ndk/25.2.9519653
OLD_PATH=${PATH}
OLD_TOOLCHAIN=${TOOLCHAIN}
export PATH=$PATH:${ANDROID_HOME}/platform-tools:${ANDROID_HOME}/cmdline-tools/latest/bin
export TOOLCHAIN=${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/linux-x86_64
export CC_AND=${TOOLCHAIN}/bin/aarch64-linux-android21-clang
export CXX_AND=${TOOLCHAIN}/bin/aarch64-linux-android21-clang++
CGO_ENABLED=1 GOOS=android GOARCH=arm64 CC=${CC_AND} CXX=${CXX_AND} fyne package --os android --release --metadata buildts="${ts}"
if [[ -f "${ANDROID_KEY}" ]] ; then
    dir=$(dirname "${ANDROID_KEY}")
    ANDROID_KEY_FILE="${dir}"/$(tail -n+4 "${ANDROID_KEY}" | head -n1)
    ANDROID_KEYSTORE_PASS=$(tail -n+5 "${ANDROID_KEY}" | head -n1)
    ANDROID_KEY_PASS=$(tail -n+6 "${ANDROID_KEY}" | head -n1)
    ${ANDROID_HOME}/build-tools/33.0.2/apksigner sign --ks "${ANDROID_KEY_FILE}" -ks-pass "pass:${ANDROID_KEYSTORE_PASS}" --key-pass "pass:${ANDROID_KEY_PASS}" VBoxSsh.apk
    rm VBoxSsh.apk.idsig
fi
PATH=${OLD_PATH}
TOOLCHAIN=${OLD_TOOLCHAIN}
mkdir -p dist/android
mv VBoxSsh.apk dist/android
