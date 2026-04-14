$ErrorActionPreference = "Stop"

$SpecFlowRoot = Split-Path -Parent $PSScriptRoot
$TargetRoot = (Get-Location).Path
$Manifest = Join-Path $SpecFlowRoot "tooling/manifest.tsv"
$ForceInstall = $false
$ManagedBegin = "<!-- SPECFLOW:BEGIN -->"
$ManagedEnd = "<!-- SPECFLOW:END -->"

function Test-ManagedEntryFile {
  param([string]$RelativePath)

  return @("AGENTS.md", "GEMINI.md", "CLAUDE.md") -contains $RelativePath
}

function Get-ManagedBlock {
  param([string]$Path)

  $lines = Get-Content $Path
  $beginMatches = @()
  $endMatches = @()

  for ($i = 0; $i -lt $lines.Count; $i++) {
    if ($lines[$i] -eq $ManagedBegin) { $beginMatches += $i }
    if ($lines[$i] -eq $ManagedEnd) { $endMatches += $i }
  }

  if ($beginMatches.Count -ne 1 -or $endMatches.Count -ne 1) {
    throw "Managed block markers must appear exactly once in $Path"
  }

  $start = $beginMatches[0]
  $end = $endMatches[0]
  if ($start -ge $end) {
    throw "Managed block markers are out of order in $Path"
  }

  return $lines[$start..$end]
}

function Set-ManagedBlock {
  param(
    [string]$SourcePath,
    [string]$DestinationPath
  )

  $sourceBlock = Get-ManagedBlock -Path $SourcePath
  $destinationLines = Get-Content $DestinationPath
  $beginMatches = @()
  $endMatches = @()

  for ($i = 0; $i -lt $destinationLines.Count; $i++) {
    if ($destinationLines[$i] -eq $ManagedBegin) { $beginMatches += $i }
    if ($destinationLines[$i] -eq $ManagedEnd) { $endMatches += $i }
  }

  if ($beginMatches.Count -ne 1 -or $endMatches.Count -ne 1) {
    throw "Managed block markers must appear exactly once in $DestinationPath"
  }

  $start = $beginMatches[0]
  $end = $endMatches[0]
  if ($start -ge $end) {
    throw "Managed block markers are out of order in $DestinationPath"
  }

  $before = if ($start -gt 0) { $destinationLines[0..($start - 1)] } else { @() }
  $after = if ($end + 1 -lt $destinationLines.Count) { $destinationLines[($end + 1)..($destinationLines.Count - 1)] } else { @() }
  $merged = @($before + $sourceBlock + $after)
  Set-Content -Path $DestinationPath -Value $merged
}

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
$Failures = 0

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

  if ((Test-Path $dest) -and (Test-ManagedEntryFile -RelativePath $parts[1])) {
    try {
      Set-ManagedBlock -SourcePath $src -DestinationPath $dest
      Write-Host "Installed managed block $($parts[1])"
      $Copied++
    } catch {
      Write-Host "Failed to install managed block into existing $($parts[1]): $($_.Exception.Message)"
      $script:Failures++
    }
    return
  }

  Copy-Item -Force $src $dest
  Write-Host "Installed $($parts[1])"
  $Copied++
}

if ($Failures -gt 0) {
  throw "specFlow init failed. copied=$Copied skipped=$Skipped failures=$Failures"
}

Write-Host "specFlow init completed. copied=$Copied skipped=$Skipped"
Write-Host "If you want Git hooks to use .githooks, run: git config core.hooksPath .githooks"
