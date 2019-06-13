#!/bin/bash

function build_binary_rice {
    name="$1"

    for arch in amd64 386; do
        GOOS=darwin GOARCH=$arch go build -o "$name"osx-$arch
        rice append --exec "$name"osx-$arch
    done

    for arch in amd64 386; do
        GOOS=freebsd GOARCH=$arch go build -o "$name"freebsd-$arch
        rice append --exec "$name"freebsd-$arch
    done

    for arch in arm amd64 386; do
        GOOS=netbsd GOARCH=$arch go build -o "$name"netbsd-$arch
        rice append --exec "$name"netbsd-$arch
    done

    for arch in amd64 386; do
        GOOS=openbsd GOARCH=$arch go build -o "$name"openbsd-$arch
        rice append --exec "$name"openbsd-$arch
    done

    for arch in arm arm64 amd64 386; do
        GOOS=linux GOARCH=$arch go build -o "$name"linux-$arch
        rice append --exec "$name"linux-$arch
    done

    for arch in amd64 386; do
        GOOS=windows GOARCH=$arch go build -o "$name"windows-$arch.exe
        rice append --exec "$name"windows-$arch.exe
    done
}

function build_binary {
    name="$1"

    for arch in amd64 386; do
        GOOS=darwin GOARCH=$arch go build -o "$name"osx-$arch
    done

    for arch in amd64 386; do
        GOOS=freebsd GOARCH=$arch go build -o "$name"freebsd-$arch
    done

    for arch in arm amd64 386; do
        GOOS=netbsd GOARCH=$arch go build -o "$name"netbsd-$arch
    done

    for arch in amd64 386; do
        GOOS=openbsd GOARCH=$arch go build -o "$name"openbsd-$arch
    done

    for arch in arm arm64 amd64 386; do
        GOOS=linux GOARCH=$arch go build -o "$name"linux-$arch
    done

    for arch in amd64 386; do
        GOOS=windows GOARCH=$arch go build -o "$name"windows-$arch.exe
    done
}

version="$1"
mkdir -p "binaries/""$version"

build_binary_rice "binaries/""$version""/linx-server-v""$version""_"

cd linx-genkey
build_binary "../binaries/""$version""/linx-genkey-v""$version""_"
cd ..

cd linx-cleanup
build_binary "../binaries/""$version""/linx-cleanup-v""$version""_"
cd ..
