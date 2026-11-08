. "$PSScriptRoot\_lib.ps1"

$RunDir = Get-RunDir
$exe = Join-Path $RunDir 'app.exe'
$pidFile = Join-Path $RunDir 'app.pid'

if (Test-Path $pidFile) {
    $procId = [int](Get-Content $pidFile)
    $proc = Get-Process -Id $procId -ErrorAction SilentlyContinue
    # Only kill the PID if it really is our binary.
    if ($proc -and $proc.Path -eq $exe) {
        Stop-Process -Id $procId -Force
        Write-Host "Stopped application (pid $procId)"
    } else {
        Write-Host 'Application is not running'
    }
    Remove-Item $pidFile -Force -ErrorAction SilentlyContinue
} else {
    Write-Host 'Application is not running'
}
