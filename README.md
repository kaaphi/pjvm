# pjvm

pjvm is a tool to change the version of the JDK used in a given shell session. You use it like this:
```shell
pjvm use <version string>
```

The tool will search the configured paths for a JDK installation that is in a path containing the provided version string.

## Supported Platforms

The tool has only been tested on Windows for the following shells:
- Git Bash
- PowerShell 7.x


## Configuration

pjvm reads from a TOML file:

```toml
basePaths = [
    # an array of strings containing the base paths to look for JDK installations within
]
```

The base paths are searched recursively, so if all your JDKs are installed in `C:\tools\java`, you can enter just that one path in the config.

By default, pvjm looks for a config file named `.pjvm` in your user profile directory.
You can override the config file location by setting the `PJVM_CONFIG` environment variable to the path of your config file.

## Installation

### Git Bash

Add the following to your `.bashrc` profile (change `pjvm` to include the full path where you've installed pjvm):
```bash
eval "`pjvm install -shell GitBash`"
```

### PowerShell

Add the following to your end of your profile (change `pjvm.exe` to include the full path where you've installed pjvm):
```powershell
pjvm.exe install -shell PowerShell | Out-String | Invoke-Expression
```

- For Windows location is either:
  - `%userprofile%\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1` Powershell 5
  - `%userprofile%\Documents\PowerShell\Microsoft.PowerShell_profile.ps1` Powershell 6+
- To create the profile file you can run this in PowerShell:
  ```
  if (-not (Test-Path $profile)) { New-Item $profile -Force }
  ```
