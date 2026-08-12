param([ValidateRange(1, 65535)][int]$Port = 8000)
$ErrorActionPreference = 'Stop'
$repoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..\..')).Path
$runtimeDir = Join-Path $repoRoot 'editor\runtime'
$binaryDir = Join-Path $runtimeDir 'bin'
$serverBinary = Join-Path $binaryDir 'site-server.exe'
$url = "http://127.0.0.1:$Port/"

# Check the requested port before building or starting a second background server.
$probe = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, $Port)
try { $probe.Start() } catch { throw "Port $Port is already in use. Close its server or choose another -Port." } finally { $probe.Stop() }
[System.IO.Directory]::CreateDirectory($binaryDir) | Out-Null
& (Join-Path $PSScriptRoot 'run-go.cmd') build -o $serverBinary ./cmd/site
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$stdoutLog = Join-Path $runtimeDir 'site-server.log'
$stderrLog = Join-Path $runtimeDir 'site-server-error.log'
$serverProcess = Start-Process -FilePath $serverBinary -ArgumentList @('serve', '--root', ('"{0}"' -f $repoRoot), '--port', $Port) -WindowStyle Hidden -PassThru -RedirectStandardOutput $stdoutLog -RedirectStandardError $stderrLog
$ready = $false
$deadline = [DateTime]::UtcNow.AddSeconds(15)
while ([DateTime]::UtcNow -lt $deadline) {
  if ($serverProcess.HasExited) { throw "Site server stopped. Check $stderrLog." }
  try {
    $response = Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 1
    if ($response.StatusCode -eq 200) { $ready = $true; break }
  } catch { Start-Sleep -Milliseconds 200 }
}
if (-not $ready) { throw "Site server did not become ready. Check $stderrLog." }
Write-Output "Site server running at $url (PID $($serverProcess.Id))."
Start-Process -FilePath $url
