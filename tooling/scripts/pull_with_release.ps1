param(
    [switch]$Help
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Show-Usage {
    [Console]::Error.WriteLine(@"
Usage: pull_with_release.ps1

Pull the current SpecFlow branch from origin.
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

if ($Help) {
    Show-Usage
    exit 0
}

$scriptDir = Split-Path -Parent $PSCommandPath
$repoRoot = (Resolve-Path (Join-Path $scriptDir "../..")).Path

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

    Write-Host "Pulling $branch from origin..."
    Invoke-CheckedNative "git" @("pull", "--ff-only", "origin", $branch)

    # Clear tooling/bin before updating binaries, so stale files are
    # removed before fresh ones are downloaded.
    $binDir = Join-Path $repoRoot "tooling/bin"
    if (Test-Path -LiteralPath $binDir) {
        Remove-Item -LiteralPath $binDir -Recurse -Force
        Write-Host "Cleared tooling/bin."
    }

    # Delegate binary update to the standalone per-platform script.
    & (Join-Path $scriptDir "update_tooling_binaries.ps1")
}
catch {
    Write-Error $_.Exception.Message
    exit 1
}
