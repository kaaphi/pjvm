param([string]$JavaPath, [string]$CurrentPath = $Env:Path)

function Update-Path {
    param (
    [string[]]$RemovePaths,
    [string]$AddPath,
    [string]$CurrentPath
    )
    $paths = $CurrentPath -split ";"
    foreach ($path  in $RemovePaths) {
        $path = [regex]::Escape($path)
        $paths = $paths | Where-Object { $_ -notMatch "^$path\\?" }
    }
    return "$AddPath;" + ($paths -join ";")
}

$new_path = Update-Path -RemovePaths (Get-Command -Type Application -Name java).path -AddPath "$JavaPath" -CurrentPath $CurrentPath

Write-Output "`$Env:JAVA_HOME = `"$JavaPath`""
Write-Output "`$Env:Path = `"$new_path`""

    