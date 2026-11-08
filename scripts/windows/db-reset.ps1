# Wipes the whole database (tables, types, ...) and re-runs all migrations.
# Skip the confirmation with:  make db-reset YES=1   (or $env:YES = '1')
. "$PSScriptRoot\_lib.ps1"
Assert-Docker

$user = Get-EnvValue 'PDB_USER' 'admin'
$name = Get-EnvValue 'PDB_DB_NAME' 'doratou'

if ((Get-Setting 'YES' '') -ne '1') {
    Write-Host "This will DELETE ALL DATA in database '$name' and re-run all migrations."
    $answer = Read-Host "Type 'yes' to continue"
    if ($answer -ne 'yes') { Write-Host 'Aborted.'; exit 1 }
}

Invoke-Compose @('exec', '-T', $DbService, 'psql', '-U', $user, '-d', $name,
    '-v', 'ON_ERROR_STOP=1', '-c', 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;')
Invoke-Migrate @('up')
Write-Host "Database '$name' was reset."
