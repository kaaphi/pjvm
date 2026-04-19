$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $true

go run ..\..\fakejava install ..\fakejavas
go build ..\..
$Env:PJVM_CONFIG=(Get-Location).Path + "\.pjvm"
.\pjvm.exe env -shell PowerShell | Out-String | Invoke-Expression

pjvm list

pjvm use 1.8
java check ..\fakejavas\dir1\java1.8

pjvm use 25
java check ..\fakejavas\dir2\java25

pjvm use 17
java check ..\fakejavas\dir1\java17

pjvm use 1
java check ..\fakejavas\dir1\java1.8