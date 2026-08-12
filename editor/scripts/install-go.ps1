param(
    [Parameter(Mandatory = $true)]
    [string]$RuntimeDir
)

$ErrorActionPreference = 'Stop'
$architecture = switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'amd64' }
    'ARM64' { 'arm64' }
    default { throw "Unsupported Windows architecture: $env:PROCESSOR_ARCHITECTURE" }
}

$releases = Invoke-RestMethod -Uri 'https://go.dev/dl/?mode=json'
$file = $releases[0].files | Where-Object {
    $_.os -eq 'windows' -and $_.arch -eq $architecture -and $_.kind -eq 'archive'
} | Select-Object -First 1

if (-not $file) {
    throw "The current Windows Go archive could not be found for $architecture."
}

$archivePath = Join-Path $RuntimeDir $file.filename
$downloadUrl = "https://go.dev/dl/$($file.filename)"
$goRoot = Join-Path $RuntimeDir 'go'
$readyMarker = Join-Path $RuntimeDir '.portable-go-ready'

try {
    $actualHash = if (Test-Path -LiteralPath $archivePath) {
        (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
    } else {
        ''
    }

    if ($actualHash -ne $file.sha256.ToLowerInvariant()) {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath
        $actualHash = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
    }

    if ($actualHash -ne $file.sha256.ToLowerInvariant()) {
        throw 'The downloaded Go archive failed its SHA-256 check.'
    }

    if (Test-Path -LiteralPath $goRoot) {
        Remove-Item -LiteralPath $goRoot -Recurse -Force
    }
    if (Test-Path -LiteralPath $readyMarker) {
        Remove-Item -LiteralPath $readyMarker -Force
    }

    # Windows PowerShell's Expand-Archive is extremely slow for the Go archive.
    # The inbox bsdtar handles ZIP files and completes this step much faster.
    & tar.exe -xf $archivePath -C $RuntimeDir
    if ($LASTEXITCODE -ne 0) {
        throw "tar.exe failed to extract the portable Go toolchain (exit $LASTEXITCODE)."
    }
    if (-not (Test-Path -LiteralPath (Join-Path $goRoot 'src\unsafe\unsafe.go'))) {
        throw 'The portable Go toolchain was not completely extracted.'
    }
    Set-Content -LiteralPath $readyMarker -Value $file.version -Encoding Ascii
}
finally {
    if ((Test-Path -LiteralPath $readyMarker) -and (Test-Path -LiteralPath $archivePath)) {
        Remove-Item -LiteralPath $archivePath -Force
    }
}
