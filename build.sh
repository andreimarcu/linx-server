#!/bin/bash

version="$1"
mkdir -p "binairies/""$version"
name="binairies/""$version""/linx-server-v""$version""_"

GOOS=darwin GOARCH=amd64 go build -o "$name"osx-amd64
rice append --exec "$name"osx-amd64

GOOS=darwin GOARCH=386 go build -o "$name"osx-386
rice append --exec "$name"osx-386

GOOS=freebsd GOARCH=amd64 go build -o "$name"freebsd-amd64
rice append --exec "$name"freebsd-amd64

GOOS=freebsd GOARCH=386 go build -o "$name"freebsd-386
rice append --exec "$name"freebsd-386

GOOS=openbsd GOARCH=amd64 go build -o "$name"openbsd-amd64
rice append --exec "$name"openbsd-amd64

GOOS=openbsd GOARCH=386 go build -o "$name"openbsd-386
rice append --exec "$name"openbsd-386

GOOS=linux GOARCH=arm go build -o "$name"linux-arm
rice append --exec "$name"linux-arm

GOOS=linux GOARCH=amd64 go build -o "$name"linux-amd64
rice append --exec "$name"linux-amd64

GOOS=linux GOARCH=386 go build -o "$name"linux-386
rice append --exec "$name"linux-386

GOOS=windows GOARCH=amd64 go build -o "$name"windows-amd64.exe
rice append --exec "$name"windows-amd64.exe

GOOS=windows GOARCH=386 go build -o "$name"windows-386.exe
rice append --exec "$name"windows-386.exe


cd linx-genkey
name="../binairies/""$version""/linx-genkey-v""$version""_"

GOOS=darwin GOARCH=amd64 go build -o "$name"osx-amd64

GOOS=darwin GOARCH=386 go build -o "$name"osx-386

GOOS=freebsd GOARCH=amd64 go build -o "$name"freebsd-amd64

GOOS=freebsd GOARCH=386 go build -o "$name"freebsd-386

GOOS=openbsd GOARCH=amd64 go build -o "$name"openbsd-amd64

GOOS=openbsd GOARCH=386 go build -o "$name"openbsd-386

GOOS=linux GOARCH=arm go build -o "$name"linux-arm

GOOS=linux GOARCH=amd64 go build -o "$name"linux-amd64

GOOS=linux GOARCH=386 go build -o "$name"linux-386

GOOS=windows GOARCH=amd64 go build -o "$name"windows-amd64.exe

GOOS=windows GOARCH=386 go build -o "$name"windows-386.exe

cd ..


cd linx-cleanup
name="../binairies/""$version""/linx-cleanup-v""$version""_"

GOOS=darwin GOARCH=amd64 go build -o "$name"osx-amd64

GOOS=darwin GOARCH=386 go build -o "$name"osx-386

GOOS=freebsd GOARCH=amd64 go build -o "$name"freebsd-amd64

GOOS=freebsd GOARCH=386 go build -o "$name"freebsd-386

GOOS=openbsd GOARCH=amd64 go build -o "$name"openbsd-amd64

GOOS=openbsd GOARCH=386 go build -o "$name"openbsd-386

GOOS=linux GOARCH=arm go build -o "$name"linux-arm

GOOS=linux GOARCH=amd64 go build -o "$name"linux-amd64

GOOS=linux GOARCH=386 go build -o "$name"linux-386

GOOS=windows GOARCH=amd64 go build -o "$name"windows-amd64.exe

GOOS=windows GOARCH=386 go build -o "$name"windows-386.exe

cd ..
