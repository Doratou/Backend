. "$PSScriptRoot\_lib.ps1"

$RunDir = Get-RunDir
New-Item -ItemType Directory -Force $RunDir | Out-Null

$workers = @('email', 'sms', 'ws')
$processes = @()

Write-Host "Building Doratou notification workers (email, sms, ws)..." -ForegroundColor Cyan
foreach ($w in $workers) {
    $exe = Join-Path $RunDir "worker-$w.exe"
    Invoke-Checked 'go' @('build', '-o', $exe, "./cmd/worker/$w")
}

Write-Host "Starting all workers. Press Ctrl+C to terminate." -ForegroundColor Green
try {
    foreach ($w in $workers) {
        $exe = Join-Path $RunDir "worker-$w.exe"
        $pidFile = Join-Path $RunDir "worker-$w.pid"
        $p = Start-Process -FilePath $exe -NoNewWindow -PassThru
        $null = $p.Handle
        Set-Content -Path $pidFile -Value $p.Id
        $processes += $p
        Write-Host "  -> Started $w worker (pid $($p.Id))" -ForegroundColor DarkCyan
    }

    # Monitor processes until user interrupts with Ctrl+C or a process exits
    while ($true) {
        Start-Sleep -Seconds 1
        foreach ($p in $processes) {
            if ($p.HasExited) {
                Write-Host "Worker process $($p.Id) exited with code $($p.ExitCode)." -ForegroundColor Yellow
                break
            }
        }
    }
} finally {
    Write-Host "`nShutting down all workers..." -ForegroundColor Yellow
    foreach ($p in $processes) {
        if (-not $p.HasExited) {
            Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue
        }
    }
    foreach ($w in $workers) {
        $pidFile = Join-Path $RunDir "worker-$w.pid"
        Remove-Item $pidFile -Force -ErrorAction SilentlyContinue
    }
    Write-Host "All workers stopped cleanly." -ForegroundColor Green
}
