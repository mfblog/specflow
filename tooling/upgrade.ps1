$ErrorActionPreference = "Stop"

$SpecFlowRoot = Split-Path -Parent $PSScriptRoot
$TargetRoot = (Get-Location).Path
$Manifest = Join-Path $SpecFlowRoot "tooling/manifest.tsv"
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
    "-h" { Write-Host "Usage: .\specflow\tooling\upgrade.ps1"; exit 0 }
    "--help" { Write-Host "Usage: .\specflow\tooling\upgrade.ps1"; exit 0 }
    default { throw "Unknown option: $arg" }
  }
}

$Updated = 0
$Skipped = 0
$Failures = 0

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

  if (Test-ManagedEntryFile -RelativePath $parts[1]) {
    try {
      $srcBlock = Get-ManagedBlock -Path $src
      $destBlock = Get-ManagedBlock -Path $dest
      if (([string]::Join("`n", $srcBlock)) -eq ([string]::Join("`n", $destBlock))) {
        return
      }

      Set-ManagedBlock -SourcePath $src -DestinationPath $dest
      Write-Host "Updated managed block $($parts[1])"
      $Updated++
    } catch {
      Write-Host "Failed to update managed block $($parts[1]): $($_.Exception.Message)"
      $script:Failures++
    }
    return
  }

  if ($mode -ne "framework") {
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

if ($Failures -gt 0) {
  throw "specFlow upgrade failed. updated=$Updated skipped=$Skipped failures=$Failures"
}

Write-Host "specFlow upgrade completed. updated=$Updated skipped=$Skipped"
