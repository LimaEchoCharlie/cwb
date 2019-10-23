#!/usr/bin/env bash

rm -fr build/*

# build golang code into shared library
env GOPATH="${GOPATH}:$(pwd)/go" CGO_CFLAGS="-I$(pwd)/shared" go build -o build/libscramble.so -buildmode=c-shared scramble

# build C executable
gcc -o build/scrambler -Ibuild -Ishared -Lbuild -lscramble c/src/scambler.c