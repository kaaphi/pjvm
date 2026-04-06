function pjvm {
	param(
		[Parameter(ValueFromRemainingArguments)]
		[string[]]$AllArgs
	)
	$eval_lines = @()
	$read_eval_lines = $false
	& "@@@PJVM_EXEC@@@" -shell PowerShell $AllArgs | ForEach-Object {
		switch ($_) {
            "@@@START_SHELL@@@" {
                $read_eval_lines = $true
            }
            "@@@END_SHELL@@@" {
                if ($read_eval_lines) {
                    $nl = [Environment]::NewLine
                    $commands = $eval_lines -join $nl
                    Invoke-Expression -Command $commands
                }
                $read_eval_lines = $false
                $eval_lines = @()
            }
            default {
                if ($read_eval_lines) {
                    $eval_lines += $_
                } else {
                    Write-Host $_
                }
            }
		}
	}
}