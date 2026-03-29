param([string]$Shell)

switch ($Shell) {
    "gitbash" {
        Write-Output "pjvm() { eval `"`$(pwsh -File `"``cygpath -u `"$PSScriptRoot\pjvm.ps1`"```" -Version `"`$1`" -Shell gitbash)`"; }"
    }

    "powershell" {
        Write-Output "function pjvm(`$v) { pwsh -File `"$PSScriptRoot\pjvm.ps1`" -Version `"`$v`" -Shell powershell | Invoke-Expression }"
    }
    
    default {
        Write-Error "Invalid shell type: $Shell"
    }
}