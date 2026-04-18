## Local manual testing

### Installing on PowerShell

```powershell
go run . env --shell PowerShell | Out-String | Invoke-Expression
```

### Installing on Git Bash

```bash
eval "`go run . env -shell GitBash`"
```