param([string]$Version, [string]$Shell, [string]$ConfigFile)

function Write-Shell {
    param (
    [string]$Message,
    [switch]$IsError
    )
    switch($Shell) {
        "gitbash" {
            if ($IsError) {
                Write-Output ">&2 echo `"$Message`""
            } else {
                Write-Output "echo `"$Message`""
            }
        }
        "powershell" {
            if ($IsError) {
                Write-Output "Write-Error `"$Message`""
            } else {
                Write-Output "Write-Host `"$Message`""
            }
        }
    }
}

if (!$Version) {
    Write-Shell -IsError "You must specify a version."
    Exit
}

if (!$ConfigFile) {
    if ($Env:PJVM_CONFIG) {
        $ConfigFile = $Env:PJVM_CONFIG
    } else {
        $ConfigFile = "~/.pjvm"
    }
}

if (!(Test-Path $ConfigFile -PathType Leaf)) {
    Write-Shell -IsError "No configuration file at $ConfigFile"
    Exit
}

$config = Import-PowerShellDataFile $ConfigFile

$paths = @()
foreach ($path in $config.BasePaths) {
    $paths += (Get-ChildItem $path -Recurse -Filter "javac.exe").Directory.parent
}

$version_match = [regex]::Escape($Version)

foreach ($path in $paths) {
    if ($path.Name -match "\d+([.-_]\d+)*") {
        $path_version = $Matches.0
        if ($path_version -match "^$version_match([.-_]\d+)*`$") {
            $result = $path
        }
    }
}

if ($result) {
    switch ($Shell) {
        "gitbash" {
            Write-Output "PJVM_HOME=`"``cygpath -u `"$PSScriptRoot`"```""
            Write-Output "export JAVA_HOME=`"``cygpath -u `"$result`"```""
            Write-Output "export PATH=`"`$JAVA_HOME/bin:```$PJVM_HOME/clean_java_path.sh `"`$PATH`"```""
        }
        
        "powershell" {
            Write-Output "`$Env:JAVA_HOME=`"$result`""
            Write-Output "`$Env:PATH=`"$result\bin;`$($PSScriptRoot\clean_java_path.ps1)`""
        }
    }
    Write-Shell "Using Java in $result"
} else {
    Write-Shell -IsError "No version matching $Version found!"
}