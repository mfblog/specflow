param(
    [switch]$Help
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Show-Usage {
    [Console]::Error.WriteLine(@"
Usage: pull_with_release.ps1

Pull the current SpecFlow branch from origin.
After pulling, update existing project entry files' specFlow Addendum blocks
from the current templates.
Then run update_tooling_binaries.ps1 to make sure the current platform's
specflowctl and specflow-reader binaries match the pulled tooling source
fingerprint.
"@)
}

function Invoke-CheckedNative {
    param(
        [string]$FilePath,
        [string[]]$Arguments
    )

    & $FilePath @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed: $FilePath $($Arguments -join ' ')"
    }
}

function Invoke-CheckedOutput {
    param(
        [string]$FilePath,
        [string[]]$Arguments
    )

    $output = & $FilePath @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed: $FilePath $($Arguments -join ' ')"
    }
    ($output -join "`n").Trim()
}

$ManagedBegin = "==SPECFLOW:BEGIN=="
$ManagedEnd = "==SPECFLOW:END=="

function Get-ManagedBlockLines {
    param(
        [string]$Path
    )

    $lines = [string[]][System.IO.File]::ReadAllLines($Path)
    $begin = -1
    $end = -1
    for ($i = 0; $i -lt $lines.Length; $i++) {
        if ($lines[$i] -eq $ManagedBegin) {
            if ($begin -ne -1) {
                throw "Managed block begin marker must appear exactly once in $Path."
            }
            $begin = $i
        }
        if ($lines[$i] -eq $ManagedEnd) {
            if ($end -ne -1) {
                throw "Managed block end marker must appear exactly once in $Path."
            }
            $end = $i
        }
    }

    if ($begin -eq -1 -or $end -eq -1 -or $begin -ge $end) {
        throw "Managed block markers are missing or out of order in $Path."
    }

    [string[]]$lines[$begin..$end]
}

function Set-ManagedBlock {
    param(
        [string]$Path,
        [string[]]$BlockLines
    )

    $lines = [string[]][System.IO.File]::ReadAllLines($Path)
    $begin = -1
    $end = -1
    for ($i = 0; $i -lt $lines.Length; $i++) {
        if ($lines[$i] -eq $ManagedBegin) {
            if ($begin -ne -1) {
                throw "Managed block begin marker must appear exactly once in $Path."
            }
            $begin = $i
        }
        if ($lines[$i] -eq $ManagedEnd) {
            if ($end -ne -1) {
                throw "Managed block end marker must appear exactly once in $Path."
            }
            $end = $i
        }
    }

    if ($begin -eq -1 -or $end -eq -1 -or $begin -ge $end) {
        # Markers not found in target — insert the block at the beginning of the file.
        $updated = [System.Collections.Generic.List[string]]::new()
        $updated.AddRange($BlockLines)
        $updated.Add("")
        $updated.AddRange([string[]]$lines)
    }
    else {
        $updated = [System.Collections.Generic.List[string]]::new()
        if ($begin -gt 0) {
            $updated.AddRange([string[]]$lines[0..($begin - 1)])
        }
        $updated.AddRange($BlockLines)
        if ($end -lt ($lines.Length - 1)) {
            $updated.AddRange([string[]]$lines[($end + 1)..($lines.Length - 1)])
        }
    }

    $originalText = [System.IO.File]::ReadAllText($Path)
    if ($originalText.Contains("`r`n")) {
        $newline = "`r`n"
    }
    else {
        $newline = "`n"
    }
    $updatedText = ([string[]]$updated.ToArray()) -join $newline
    if ($originalText.EndsWith("`r`n") -or $originalText.EndsWith("`n")) {
        $updatedText += $newline
    }

    if ($updatedText -eq $originalText) {
        return $false
    }

    $encoding = [System.Text.UTF8Encoding]::new($false)
    [System.IO.File]::WriteAllText($Path, $updatedText, $encoding)
    return $true
}

function Sync-ExistingEntryBlocks {
    param(
        [string]$SpecFlowRoot,
        [string]$ProjectRoot
    )

    $changed = $false
    $found = $false
    foreach ($entry in @("AGENTS.md", "CLAUDE.md", "GEMINI.md")) {
        $source = Join-Path (Join-Path $SpecFlowRoot "templates") $entry
        $target = Join-Path $ProjectRoot $entry
        if (-not (Test-Path -LiteralPath $target -PathType Leaf)) {
            continue
        }
        $found = $true
        if (-not (Test-Path -LiteralPath $source -PathType Leaf)) {
            throw "Entry template missing: $source"
        }

        $block = Get-ManagedBlockLines $source
        if (Set-ManagedBlock $target $block) {
            Write-Host "Updated $entry specFlow Addendum."
            $changed = $true
        }
    }

    if (-not $found) {
        Write-Host "No existing project entry files found to update."
    }
    elseif (-not $changed) {
        Write-Host "Existing project entry Addendum blocks are already current."
    }
}

if ($Help) {
    Show-Usage
    exit 0
}

$scriptDir = Split-Path -Parent $PSCommandPath
$repoRoot = (Resolve-Path (Join-Path $scriptDir "../..")).Path
$projectRoot = (Resolve-Path (Join-Path $repoRoot "..")).Path

try {
    Set-Location $repoRoot

    $branch = Invoke-CheckedOutput "git" @("branch", "--show-current")
    if ([string]::IsNullOrWhiteSpace($branch)) {
        throw "Detached HEAD is not supported. Check out a branch before pulling."
    }

    $status = Invoke-CheckedOutput "git" @("status", "--porcelain")
    if (-not [string]::IsNullOrWhiteSpace($status)) {
        throw "Working tree is not clean. Commit or stash changes before pulling."
    }

    $remoteUrl = Invoke-CheckedOutput "git" @("remote", "get-url", "origin")
    if ([string]::IsNullOrWhiteSpace($remoteUrl)) {
        throw "Git remote 'origin' is missing."
    }

    # Sync entry blocks BEFORE pulling, so in-memory script functions
    # are used before git pull can modify the script file on disk.
    Sync-ExistingEntryBlocks $repoRoot $projectRoot

    Write-Host "Pulling $branch from origin..."
    Invoke-CheckedNative "git" @("pull", "--ff-only", "origin", $branch)

    # Delegate binary update to the standalone per-platform script.
    & (Join-Path $scriptDir "update_tooling_binaries.ps1")
}
catch {
    Write-Error $_.Exception.Message
    exit 1
}
