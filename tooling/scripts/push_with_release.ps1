param(
    [switch]$Help
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Show-Usage {
    [Console]::Error.WriteLine(@"
Usage: push_with_release.ps1

Push the current SpecFlow branch to origin.
Before pushing, update existing project entry files' specFlow Addendum blocks
from the current templates.
When the current branch is main, if the current tooling fingerprint does not
already have a remote release tag, create and push specflow-tooling-<fingerprint>
to trigger the GitHub Release workflow.
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
        throw "Managed block markers are missing or out of order in $Path."
    }

    $updated = [System.Collections.Generic.List[string]]::new()
    if ($begin -gt 0) {
        $updated.AddRange([string[]]$lines[0..($begin - 1)])
    }
    $updated.AddRange($BlockLines)
    if ($end -lt ($lines.Length - 1)) {
        $updated.AddRange([string[]]$lines[($end + 1)..($lines.Length - 1)])
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
        throw "Detached HEAD is not supported. Check out a branch before pushing."
    }

    $status = Invoke-CheckedOutput "git" @("status", "--porcelain")
    if (-not [string]::IsNullOrWhiteSpace($status)) {
        throw "Working tree is not clean. Commit or stash changes before pushing."
    }

    $remoteUrl = Invoke-CheckedOutput "git" @("remote", "get-url", "origin")
    if ([string]::IsNullOrWhiteSpace($remoteUrl)) {
        throw "Git remote 'origin' is missing."
    }

    Sync-ExistingEntryBlocks $repoRoot $projectRoot

    Write-Host "Pushing $branch to origin..."
    Invoke-CheckedNative "git" @("push", "-u", "origin", $branch)

    if ($branch -ne "main") {
        Write-Host "Current branch is $branch, not main."
        Write-Host "Release tag push is skipped."
        exit 0
    }

    $fingerprintScript = Join-Path $repoRoot "tooling/scripts/tooling_fingerprint.ps1"
    $fingerprint = (& $fingerprintScript -Short).Trim()
    $tag = "specflow-tooling-$fingerprint"

    & git ls-remote --exit-code --tags origin "refs/tags/$tag" *> $null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Release tag already exists on origin: $tag"
        Write-Host "No release tag push needed."
        exit 0
    }

    & git rev-parse -q --verify "refs/tags/$tag" *> $null
    if ($LASTEXITCODE -eq 0) {
        $tagCommit = Invoke-CheckedOutput "git" @("rev-list", "-n", "1", $tag)
        $headCommit = Invoke-CheckedOutput "git" @("rev-parse", "HEAD")
        if ($tagCommit -ne $headCommit) {
            throw "Local tag $tag exists but does not point to HEAD. Delete or inspect the local tag manually before pushing a release."
        }
    }
    else {
        Invoke-CheckedNative "git" @("tag", $tag)
    }

    Write-Host "Pushing release tag $tag..."
    Invoke-CheckedNative "git" @("push", "origin", $tag)
    Write-Host "Release workflow triggered by $tag."
}
finally {
}
