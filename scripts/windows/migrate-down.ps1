# Rolls back STEPS migrations (default 1).  Usage: make migrate-down STEPS=2
. "$PSScriptRoot\_lib.ps1"
$steps = Get-Setting 'STEPS' '1'
if ($steps -notmatch '^\d+$') { Fail "STEPS must be a positive number, got '$steps'" }
Invoke-Migrate @('down', $steps)
