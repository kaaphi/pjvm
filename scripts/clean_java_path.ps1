param([string]$CurrentPath = $Env:Path)

function Update-Path {
    param (
    [string[]]$RemovePaths,
    [string]$CurrentPath
    )
    $paths = $CurrentPath -split ";"
    foreach ($path  in $RemovePaths) {
        $path = [regex]::Escape($path)
        $paths = $paths | Where-Object { $_ -notMatch "^$path\\?" }
    }
    return ($paths -join ";")
}

$new_path = Update-Path -RemovePaths (Get-Command -Type Application -Name java).path -CurrentPath $CurrentPath

Write-Output $new_path