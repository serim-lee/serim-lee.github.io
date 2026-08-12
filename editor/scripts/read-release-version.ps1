param([Parameter(Mandatory = $true)][string]$RepositoryRoot)

$ErrorActionPreference = 'Stop'
$manifest = Get-Content -LiteralPath (Join-Path $RepositoryRoot 'app\release.json') -Raw -Encoding UTF8 | ConvertFrom-Json
$releaseVersion = [string]$manifest.version
$versionMatch = [regex]::Match($releaseVersion, '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$')
if (-not $versionMatch.Success -or $versionMatch.Value.Length -ne $releaseVersion.Length) {
    throw 'app/release.json must contain a major.minor.patch version.'
}
for ($partIndex = 1; $partIndex -le 3; $partIndex++) {
    [void][UInt64]::Parse($versionMatch.Groups[$partIndex].Value)
}
[Console]::Out.WriteLine($releaseVersion)
