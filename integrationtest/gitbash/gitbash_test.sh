#!/bin/bash
set -e

go run ../../fakejava install ../fakejavas
go build ../..
export PJVM_CONFIG="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"/.pjvm
eval "`./pjvm.exe env -shell GitBash`"
# .\pjvm.exe env -shell PowerShell | Out-String | Invoke-Expression

pjvm list

pjvm use 1.8
java check ../fakejavas/dir1/java1.8

pjvm use 25
java check ../fakejavas/dir2/java25

pjvm use 17
java check ../fakejavas/dir1/java17

pjvm use 1
java check ../fakejavas/dir1/java1.8