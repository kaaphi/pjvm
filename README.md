# pjvm

pjvm is a tool to change the version of the JDK used in a given shell session. You use it like this:
```shell
pjvm <version string>
```

The tool will search the configured paths for a JDK installation that is in a path containing the provided version string.

## Supported Platforms

The tool has only been tested on Windows for the following shells:
- Git Bash
- PowerShell 7.x


## Configuration

pjvm reads from a [PowerShell data file](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_data_files). The data file format is as follows:
```
@{
    BasePaths = @(
        "[path1]",
        "[path2]",
        ...
    )
}
```

For example, to look for JDK installations in `C:\tools\java`, you'd create a config file like this:
```
@{
    BasePaths = @(
        "C:\tools\java"
    )
}
```

By default, pvjm looks for a config file named `.pjvm` in your user profile directory.
You can override the config file location by setting the `PJVM_CONFIG` environment variable to the path of you config file.

## Installation

### Git Bash

Add the following to your `.bashrc` profile (change `pjvm_install.ps1` to the full path where you've installed pjvm):
```bash
eval "$(pwsh pjvm_install.ps1 -Shell gitbash)"
```

### PowerShell

Add the following to your end of your profile (change `pjvm_install.ps1` to the full path where you've installed pjvm):
```powershell
pjvm_install.ps1 -Shell powershell | Invoke-Expression
```

- For Windows location is either:
  - `%userprofile%\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1` Powershell 5
  - `%userprofile%\Documents\PowerShell\Microsoft.PowerShell_profile.ps1` Powershell 6+
- To create the profile file you can run this in PowerShell:
  ```
  if (-not (Test-Path $profile)) { New-Item $profile -Force }
  ```
