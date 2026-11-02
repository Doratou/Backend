. "$PSScriptRoot\_lib.ps1"
Assert-Docker
Invoke-Compose @('up', '-d', '--wait')
