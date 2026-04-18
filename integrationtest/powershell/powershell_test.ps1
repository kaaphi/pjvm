$ErrorActionPreference = 'Stop'

# Helper function to check external command exit codes
function Test-ExitCode {
    param([string]$Command)
    if ($LASTEXITCODE -ne 0) {
        throw "External command failed: $Command (Exit code: $LASTEXITCODE)"
    }
}

go run ..\..\fakejava install ..\fakejavas
go build ..\..
$Env:PJVM_CONFIG=(Get-Location).Path + "\.pjvm"
.\pjvm.exe env -shell PowerShell | Out-String | Invoke-Expression

pjvm use 1.8
java check ..\fakejavas\dir1\java1.8
Test-ExitCode -Command java1.8

pjvm use 25
java check ..\fakejavas\dir2\java25
Test-ExitCode -Command java25