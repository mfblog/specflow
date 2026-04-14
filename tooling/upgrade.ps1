$ErrorActionPreference = "Stop"

$SpecFlowRoot = Split-Path -Parent $PSScriptRoot
$TargetRoot = (Get-Location).Path
$Manifest = Join-Path $SpecFlowRoot "tooling/manifest.tsv"
$ForceInstall = $false

foreach ($arg in $args) {
  switch ($arg) {
    "--force" { $ForceInstall = $true }
    "-h" { Write-Host "Usage: .\specflow\tooling\upgrade.ps1 [--force]"; exit 0 }
    "--help" { Write-Host "Usage: .\specflow\tooling\upgrade.ps1 [--force]"; exit 0 }
    default { throw "Unknown option: $arg" }
  }
}

$Updated = 0
$Skipped = 0

Get-Content $Manifest | ForEach-Object {
  if ([string]::IsNullOrWhiteSpace($_)) { return }
  $parts = $_ -split "`t"
  $src = Join-Path $SpecFlowRoot $parts[0]
  $dest = Join-Path $TargetRoot $parts[1]
  $mode = $parts[2]

  if (!(Test-Path $dest)) {
    $destDir = Split-Path -Parent $dest
    if ($destDir -and !(Test-Path $destDir)) {
      New-Item -ItemType Directory -Force -Path $destDir | Out-Null
    }
    Copy-Item -Force $src $dest
    Write-Host "Installed missing $($parts[1])"
    $Updated++
    return
  }

  if (($mode -ne "framework") -and -not $ForceInstall) {
    Write-Host "Skip project-owned file: $($parts[1])"
    $Skipped++
    return
  }

  $srcHash = (Get-FileHash $src).Hash
  $destHash = (Get-FileHash $dest).Hash
  if ($srcHash -eq $destHash) {
    return
  }

  Copy-Item -Force $src $dest
  Write-Host "Updated $($parts[1])"
  $Updated++
}

Write-Host "specFlow upgrade completed. updated=$Updated skipped=$Skipped"
