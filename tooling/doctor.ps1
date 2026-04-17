$ErrorActionPreference = "Stop"

$SpecFlowRoot = Split-Path -Parent $PSScriptRoot
$TargetRoot = (Get-Location).Path
$Manifest = Join-Path $SpecFlowRoot "tooling/manifest.tsv"
$Failures = 0
$ManagedBegin = "<!-- SPECFLOW:BEGIN -->"
$ManagedEnd = "<!-- SPECFLOW:END -->"
$CurrentBinaryRel = "specflow/tooling/bin/specflowctl-windows-amd64.exe"

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

Get-Content $Manifest | ForEach-Object {
  if ([string]::IsNullOrWhiteSpace($_)) { return }
  $parts = $_ -split "`t"
  $dest = Join-Path $TargetRoot $parts[1]
  if (!(Test-Path $dest)) {
    Write-Host "MISSING $($parts[1])"
    $script:Failures++
  }
}

$Agents = Join-Path $TargetRoot "AGENTS.md"
if (Test-Path $Agents) {
  $a = $null
  try {
    $a = Get-ManagedBlock -Path $Agents
  } catch {
    Write-Host "INVALID managed block in AGENTS.md: $($_.Exception.Message)"
    $Failures++
  }
  foreach ($peerName in @("GEMINI.md", "CLAUDE.md")) {
    $peer = Join-Path $TargetRoot $peerName
    if (Test-Path $peer) {
      try {
        $p = Get-ManagedBlock -Path $peer
      } catch {
        Write-Host "INVALID managed block in $peerName: $($_.Exception.Message)"
        $Failures++
        continue
      }
      if ($a -and ([string]::Join("`n", $a)) -ne ([string]::Join("`n", $p))) {
        Write-Host "DIFF managed blocks in AGENTS.md and $peerName"
        $Failures++
      }
    }
  }
}

try {
  $HookPath = git -C $TargetRoot config --get core.hooksPath 2>$null
  if ($HookPath -ne ".githooks") {
    Write-Host "WARN git core.hooksPath is not .githooks"
  }
} catch {
}

$Arch = $env:PROCESSOR_ARCHITECTURE
if ($Arch -eq "ARM64") {
  $CurrentBinaryRel = "specflow/tooling/bin/specflowctl-windows-arm64.exe"
}

$CurrentBinary = Join-Path $TargetRoot $CurrentBinaryRel
if (!(Test-Path $CurrentBinary)) {
  Write-Host "MISSING $CurrentBinaryRel"
  $Failures++
}

$HookFile = Join-Path $TargetRoot ".githooks/pre-commit"
if ((Test-Path $HookFile) -and (-not (Select-String -Path $HookFile -SimpleMatch "specflow/tooling/bin/specflowctl-" -Quiet) -or -not (Select-String -Path $HookFile -SimpleMatch "entry sync --stage" -Quiet))) {
  Write-Host "INVALID .githooks/pre-commit does not call specflow binary entry sync"
  $Failures++
}

if ($Failures -gt 0) {
  throw "specFlow doctor failed: $Failures issue(s)"
}

Write-Host "specFlow doctor passed"
