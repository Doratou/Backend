. "$PSScriptRoot\_lib.ps1"

$RunDir = Get-RunDir
New-Item -ItemType Directory -Force $RunDir | Out-Null
$exe = Join-Path $RunDir 'app.exe'
$pidFile = Join-Path $RunDir 'app.pid'

if (Test-Path $pidFile) {
    $oldPid = [int](Get-Content $pidFile)
    $proc = Get-Process -Id $oldPid -ErrorAction SilentlyContinue
    if ($proc -and $proc.Path -eq $exe) {
        Write-Host 'Application is already running'
        exit 1
    }
}

Invoke-Checked 'go' @('build', '-o', $exe, './cmd/app')

$p = Start-Process -FilePath $exe -NoNewWindow -PassThru
$null = $p.Handle   # keeps the handle so ExitCode is available later
Set-Content -Path $pidFile -Value $p.Id
try {
    $p.WaitForExit()
} finally {
    if (-not $p.HasExited) { Stop-Process -Id $p.Id -Force }
    Remove-Item $pidFile -Force -ErrorAction SilentlyContinue
}
exit $p.ExitCode
