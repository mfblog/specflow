$ErrorActionPreference = "Stop"

$SpecFlowRoot = Split-Path -Parent $PSScriptRoot
$TargetRoot = (Get-Location).Path
$Manifest = Join-Path $SpecFlowRoot "tooling/manifest.tsv"
$Failures = 0

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
$Gemini = Join-Path $TargetRoot "GEMINI.md"
if ((Test-Path $Agents) -and (Test-Path $Gemini)) {
  $a = Get-FileHash $Agents
  $g = Get-FileHash $Gemini
  if ($a.Hash -ne $g.Hash) {
    Write-Host "DIFF AGENTS.md and GEMINI.md are inconsistent"
    $Failures++
  }
}

try {
  $HookPath = git -C $TargetRoot config --get core.hooksPath 2>$null
  if ($HookPath -ne ".githooks") {
    Write-Host "WARN git core.hooksPath is not .githooks"
  }
} catch {
}

if ($Failures -gt 0) {
  throw "specFlow doctor failed: $Failures issue(s)"
}

Write-Host "specFlow doctor passed"
