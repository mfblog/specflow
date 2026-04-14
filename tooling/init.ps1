$ErrorActionPreference = "Stop"

$SpecFlowRoot = Split-Path -Parent $PSScriptRoot
$TargetRoot = (Get-Location).Path
$Manifest = Join-Path $SpecFlowRoot "tooling/manifest.tsv"
$ForceInstall = $false

foreach ($arg in $args) {
  switch ($arg) {
    "--force" { $ForceInstall = $true }
    "-h" { Write-Host "Usage: .\specflow\tooling\init.ps1 [--force]"; exit 0 }
    "--help" { Write-Host "Usage: .\specflow\tooling\init.ps1 [--force]"; exit 0 }
    default { throw "Unknown option: $arg" }
  }
}

if (!(Test-Path $Manifest)) {
  throw "Manifest not found: $Manifest"
}

$Copied = 0
$Skipped = 0

Get-Content $Manifest | ForEach-Object {
  if ([string]::IsNullOrWhiteSpace($_)) { return }
  $parts = $_ -split "`t"
  $src = Join-Path $SpecFlowRoot $parts[0]
  $dest = Join-Path $TargetRoot $parts[1]
  $mode = $parts[2]

  if (!(Test-Path $src)) {
    throw "Template not found: $src"
  }

  $destDir = Split-Path -Parent $dest
  if ($destDir -and !(Test-Path $destDir)) {
    New-Item -ItemType Directory -Force -Path $destDir | Out-Null
  }

  if ((Test-Path $dest) -and -not $ForceInstall) {
    Write-Host "Skip existing $mode file: $($parts[1])"
    $Skipped++
    return
  }

  Copy-Item -Force $src $dest
  Write-Host "Installed $($parts[1])"
  $Copied++
}

Write-Host "specFlow init completed. copied=$Copied skipped=$Skipped"
Write-Host "If you want Git hooks to use .githooks, run: git config core.hooksPath .githooks"
