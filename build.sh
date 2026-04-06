#!/bin/bash

mkdir -p dist
go build -o dist/ .
cp scripts/* dist