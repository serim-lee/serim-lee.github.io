param([Parameter(Mandatory = $true)][string] $Handoff, [Parameter(Mandatory = $true)][string] $Generation)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$lock = $null
$pending = $null
$resultPath = $null
$temporary = $null
$backup = $null
$target = $null
$replaced = $false

function Assert-Within([string] $Root, [string] $Name) {
    $full = [IO.Path]::GetFullPath($Name)
    $prefix = $Root.TrimEnd('\') + '\'
    if (-not $full.Equals($Root, [StringComparison]::OrdinalIgnoreCase) -and -not $full.StartsWith($prefix, [StringComparison]::OrdinalIgnoreCase)) { throw 'Update path leaves repository.' }
    $current = $Root
    $relative = $full.Substring($Root.Length).TrimStart('\')
    foreach ($part in $relative.Split('\')) {
        if ($part.Length -eq 0) { continue }
        $current = [IO.Path]::Combine($current, $part)
        if ([IO.File]::Exists($current) -or [IO.Directory]::Exists($current)) {
            if (([IO.File]::GetAttributes($current) -band [IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'Update path contains a link or reparse point.' }
        }
    }
    return $full
}
function Assert-FileOrMissing([string] $Name) {
    if ([IO.Directory]::Exists($Name)) { throw 'Expected a file path.' }
}
function Get-Digest([string] $Name) {
    return (Get-FileHash -LiteralPath $Name -Algorithm SHA256).Hash.ToLowerInvariant()
}
function Test-Active {
    $current = Get-Content -LiteralPath $Handoff -Raw -Encoding UTF8 | ConvertFrom-Json
    if ($current.id -ne $pending.id) { return $false }
    if ([IO.File]::Exists($resultPath)) {
        $result = Get-Content -LiteralPath $resultPath -Raw -Encoding UTF8 | ConvertFrom-Json
        if ($result.id -eq $pending.id -and $result.finalized) { return $false }
    }
    return $true
}
function Write-Result([string] $Status, [string] $Message) {
    if (-not (Test-Active)) { return }
    $result = [ordered]@{ id = $pending.id; status = $Status; message = $Message; updated_at = [DateTime]::UtcNow.ToString('o'); finalized = $false }
    $resultTemporary = $resultPath + '.' + [Guid]::NewGuid().ToString('N') + '.tmp'
    [IO.File]::WriteAllText($resultTemporary, ($result | ConvertTo-Json), [Text.UTF8Encoding]::new($false))
    Move-Item -LiteralPath $resultTemporary -Destination $resultPath -Force
}
try {
    $Handoff = [IO.Path]::GetFullPath($Handoff)
    $pending = Get-Content -LiteralPath $Handoff -Raw -Encoding UTF8 | ConvertFrom-Json
    if ($pending.id -ne $Generation) { exit 0 }
    if ($pending.format_version -ne 1 -or $pending.id -cnotmatch '^[a-f0-9]{24}$' -or $pending.expected_sha256 -cnotmatch '^[a-f0-9]{64}$') { throw 'Invalid editor update handoff.' }
    $root = [IO.Path]::GetFullPath($pending.root).TrimEnd('\')
    if (-not [IO.Directory]::Exists($root) -or ([IO.File]::GetAttributes($root) -band [IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'Invalid repository root.' }
    $transaction = Assert-Within $root $pending.transaction_dir
    $transactionsRoot = [IO.Path]::Combine($root, 'editor', 'runtime', 'updates')
    if (-not ([IO.Directory]::GetParent($transaction).FullName).Equals($transactionsRoot, [StringComparison]::OrdinalIgnoreCase)) { throw 'Invalid transaction directory.' }
    $stateDirectory = Assert-Within $root ([IO.Path]::Combine($transaction, 'self-update'))
    if (-not $PSScriptRoot.Equals($stateDirectory, [StringComparison]::OrdinalIgnoreCase) -or -not $Handoff.Equals([IO.Path]::Combine($stateDirectory, 'handoff.json'), [StringComparison]::OrdinalIgnoreCase)) { throw 'Helper and handoff directories differ.' }
    $staged = Assert-Within $root $pending.staged_executable
    $payloadPath = [IO.Path]::Combine($transaction, 'payload', 'Profile-Editor.exe')
    $rollbackPath = [IO.Path]::Combine($transaction, 'backups', 'Profile-Editor.exe')
    if (-not $staged.Equals($payloadPath, [StringComparison]::OrdinalIgnoreCase) -and -not $staged.Equals($rollbackPath, [StringComparison]::OrdinalIgnoreCase)) { throw 'Invalid staged editor path.' }
    $target = Assert-Within $root ([IO.Path]::Combine($root, 'Profile-Editor.exe'))
    $backup = Assert-Within $root ([IO.Path]::Combine($stateDirectory, 'previous.exe'))
    $temporary = Assert-Within $root ([IO.Path]::Combine($root, '.Profile-Editor-' + $pending.id + '.tmp'))
    $resultPath = Assert-Within $root ([IO.Path]::Combine($stateDirectory, 'result.json'))
    $lockPath = Assert-Within $root ([IO.Path]::Combine($stateDirectory, 'helper.lock'))
    foreach ($leaf in @($staged, $target, $backup, $temporary, $resultPath, $lockPath, $Handoff)) { Assert-FileOrMissing $leaf }
    for ($attempt = 0; $attempt -lt 90; $attempt++) {
        if (-not (Test-Active)) { exit 0 }
        try { $lock = [IO.File]::Open($lockPath, [IO.FileMode]::OpenOrCreate, [IO.FileAccess]::ReadWrite, [IO.FileShare]::None); break }
        catch [IO.IOException] { Start-Sleep -Milliseconds 1000 }
    }
    if ($null -eq $lock) { throw 'Another editor update helper is still active.' }
    if (-not (Test-Active)) { exit 0 }
    if ((Get-Digest $staged) -ne $pending.expected_sha256) { throw 'Staged editor digest mismatch.' }
    Write-Result 'waiting' ''
    if ($pending.wait_pid -gt 0) {
        try { $process = [Diagnostics.Process]::GetProcessById([int]$pending.wait_pid) }
        catch [ArgumentException] { $process = $null }
        if ($null -ne $process) {
            try {
                $started = ''
                try { $started = $process.StartTime.ToUniversalTime().ToFileTimeUtc().ToString() }
                catch [InvalidOperationException] { }
                if ($started -eq $pending.wait_started -and -not $process.WaitForExit(600000)) { throw 'The editor is still running.' }
            } finally { $process.Dispose() }
        }
    }
    if (-not (Test-Active)) { exit 0 }
    $targetDigest = ''
    if ([IO.File]::Exists($target)) {
        # A reader may hold an exclusive lock. Hashing here is an optimization;
        # the installed bytes are verified after the atomic replacement.
        try { $targetDigest = Get-Digest $target } catch { }
    }
    if ($targetDigest -eq $pending.expected_sha256) {
        $replaced = $true
    } else {
        if ([IO.File]::Exists($temporary)) { Remove-Item -LiteralPath $temporary -Force }
        [IO.File]::Copy($staged, $temporary, $false)
        [IO.File]::SetAttributes($temporary, [IO.FileAttributes]::Normal)
        if ((Get-Digest $temporary) -ne $pending.expected_sha256) { throw 'Temporary editor digest mismatch.' }
        for ($attempt = 0; $attempt -lt 90; $attempt++) {
            if (-not (Test-Active)) { exit 0 }
            foreach ($name in @($target, $temporary, $backup)) { $null = Assert-Within $root $name; Assert-FileOrMissing $name }
            try {
                if ([IO.File]::Exists($target)) { [IO.File]::Replace($temporary, $target, $backup, $true) }
                else { [IO.File]::Move($temporary, $target) }
                $replaced = $true
                break
            } catch [IO.IOException] { Start-Sleep -Milliseconds 1000 }
        }
        if (-not $replaced) { throw 'The editor executable remains locked.' }
    }
    if ((Get-Digest $target) -ne $pending.expected_sha256) { throw 'Installed editor digest mismatch.' }
    if (-not (Test-Active)) { exit 0 }
    Write-Result 'applied' ''
    if ($pending.relaunch) { Start-Process -FilePath $target -WorkingDirectory $root }
} catch {
    $message = $_.Exception.Message
    if ($null -ne $resultPath) {
        try {
            if (Test-Active) {
                if ($replaced -and [IO.File]::Exists($backup)) {
                    $restore = Assert-Within $root ([IO.Path]::Combine($stateDirectory, 'restore-' + $pending.id + '.tmp'))
                    Assert-FileOrMissing $restore
                    [IO.File]::Copy($backup, $restore, $true)
                    [IO.File]::Replace($restore, $target, $null, $true)
                }
                Write-Result 'failed' $message
            }
        } catch { Write-Error -ErrorAction Continue $_ }
    }
    Write-Error -ErrorAction Continue $message
    exit 1
} finally {
    if ($null -ne $lock) { $lock.Dispose() }
    if ($null -ne $temporary -and [IO.File]::Exists($temporary)) {
        try { $null = Assert-Within $root $temporary; Remove-Item -LiteralPath $temporary -Force } catch { }
    }
}
