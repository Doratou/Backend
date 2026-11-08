. "$PSScriptRoot\_lib.ps1"
Assert-Docker
Invoke-Compose @('down')
