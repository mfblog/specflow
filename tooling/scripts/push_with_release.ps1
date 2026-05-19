param(
    [switch]$Help
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Show-Usage {
    [Console]::Error.WriteLine(@"
Usage: push_with_release.ps1

Push the current SpecFlow branch to origin.
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

function New-FingerprintRoot {
    param(
        [string]$RepoRoot
    )

    $parentRoot = (Resolve-Path (Join-Path $RepoRoot "..")).Path
    if (Test-Path -LiteralPath (Join-Path $parentRoot "specflow/tooling/manifest.tsv") -PathType Leaf) {
        return [pscustomobject]@{
            Path = $parentRoot
            Temporary = $false
        }
    }

    $tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("specflow-fingerprint-" + [System.Guid]::NewGuid().ToString("N"))
    New-Item -ItemType Directory -Path $tempRoot | Out-Null
    try {
        New-Item -ItemType SymbolicLink -Path (Join-Path $tempRoot "specflow") -Target $RepoRoot | Out-Null
    }
    catch {
        Remove-Item -LiteralPath $tempRoot -Recurse -Force -ErrorAction SilentlyContinue
        throw "Cannot create temporary specflow link for fingerprint calculation. Rename the repository directory to 'specflow' or enable symbolic links."
    }

    [pscustomobject]@{
        Path = $tempRoot
        Temporary = $true
    }
}

if ($Help) {
    Show-Usage
    exit 0
}

$scriptDir = Split-Path -Parent $PSCommandPath
$repoRoot = (Resolve-Path (Join-Path $scriptDir "../..")).Path
$fingerprintRoot = $null

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

    Write-Host "Pushing $branch to origin..."
    Invoke-CheckedNative "git" @("push", "-u", "origin", $branch)

    if ($branch -ne "main") {
        Write-Host "Current branch is $branch, not main."
        Write-Host "Release tag push is skipped."
        exit 0
    }

    $fingerprintRoot = New-FingerprintRoot $repoRoot
    $fingerprintScript = Join-Path $fingerprintRoot.Path "specflow/tooling/scripts/tooling_fingerprint.ps1"
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
    if ($null -ne $fingerprintRoot -and $fingerprintRoot.Temporary) {
        Remove-Item -LiteralPath $fingerprintRoot.Path -Recurse -Force -ErrorAction SilentlyContinue
    }
}
