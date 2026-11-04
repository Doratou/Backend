# Shared helpers for the Windows scripts. Dot-sourced by the other scripts.
# Keep this file ASCII-only so Windows PowerShell 5.1 reads it correctly.
$ErrorActionPreference = 'Stop'

$RootDir = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
Set-Location $RootDir

function Get-Setting([string]$Name, [string]$Default) {
    $v = [Environment]::GetEnvironmentVariable($Name)
    if ($v) { return $v }
    return $Default
}

$ComposePath   = Get-Setting 'COMPOSE_PATH'   'docker/docker-compose.yaml'
$EnvFile       = Get-Setting 'ENV_FILE'       '.env'
$MigrationsDir = Get-Setting 'MIGRATIONS_DIR' (Join-Path $RootDir 'migrations')
$DbService     = Get-Setting 'DB_SERVICE'     'db'

function Fail([string]$Message) {
    Write-Host "ERROR: $Message" -ForegroundColor Red
    exit 1
}

# Runs a native command and stops the script if it fails.
function Invoke-Checked([string]$Exe, [string[]]$CmdArgs) {
    & $Exe @CmdArgs
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Assert-Docker {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) { Fail 'docker is not installed or not in PATH.' }
    & docker info *> $null
    if ($LASTEXITCODE -ne 0) { Fail 'Docker daemon is not running.' }
}

# Get-EnvValue KEY DEFAULT : shell env first, then .env file, then default.
function Get-EnvValue([string]$Key, [string]$Default) {
    $v = [Environment]::GetEnvironmentVariable($Key)
    if ($v) { return $v }
    if (Test-Path $EnvFile) {
        $line = Get-Content $EnvFile |
            Where-Object { $_ -match ('^\s*' + [regex]::Escape($Key) + '=') } |
            Select-Object -Last 1
        if ($line) {
            $val = ($line -split '=', 2)[1].Trim().Trim('"').Trim("'")
            if ($val) { return $val }
        }
    }
    return $Default
}

function Invoke-Compose([string[]]$ComposeArgs) {
    Invoke-Checked 'docker' (@('compose', '--env-file', $EnvFile, '-f', $ComposePath) + $ComposeArgs)
}

function Get-DbContainer {
    $id = & docker compose --env-file $EnvFile -f $ComposePath ps -q $DbService | Select-Object -First 1
    if (-not $id) { Fail "Database container '$DbService' is not running. Run 'make up' first." }
    return $id
}

# Uses the locally installed Goose CLI. The database must be reachable at
# localhost on DB_HOST_PORT (published by Compose).
function Invoke-Migrate([string[]]$MigrateArgs) {
    if (-not (Get-Command goose -ErrorAction SilentlyContinue)) { Fail 'goose is not installed or not in PATH.' }
    if (-not (Test-Path $MigrationsDir)) { Fail "Migrations folder not found: $MigrationsDir" }
    $user = Get-EnvValue 'PDB_USER' 'admin'
    $pass = Get-EnvValue 'PDB_PASSWORD' 'admin'
    $name = Get-EnvValue 'PDB_DB_NAME' 'doratou'
    $port = Get-EnvValue 'DB_HOST_PORT' (Get-EnvValue 'PDB_PORT' '5432')
    $url  = "postgres://${user}:${pass}@127.0.0.1:${port}/${name}?sslmode=disable"
    Invoke-Checked 'goose' (@('-dir', $MigrationsDir, 'postgres', $url) + $MigrateArgs)
}

# Per-checkout runtime dir (pid file + built binary).
function Get-RunDir {
    $md5 = [System.Security.Cryptography.MD5]::Create()
    $bytes = $md5.ComputeHash([Text.Encoding]::UTF8.GetBytes($RootDir))
    $hash = ([BitConverter]::ToString($bytes) -replace '-', '').Substring(0, 10).ToLower()
    return (Join-Path ([IO.Path]::GetTempPath()) "doratou-$hash")
}
