## Local manual testing

To make sure your install is using the correct clean scripts, set the `PJVM_SCRIPT_HOME` env variable to the directory in your repo clone.

### Installing on PowerShell

```powershell
$Env
$Env:PJVM_SCRIPT_HOME="C:\dev\pjvm\scripts"
go run . install --shell PowerShell | Out-String | Invoke-Expression
```

### Installing on Git Bash

```bash
export PJVM_SCRIPT_HOME="c:\dev\pjvm"
eval "`go run . install -shell GitBash`"
```