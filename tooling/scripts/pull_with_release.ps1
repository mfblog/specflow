param(
    [switch]$Help
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Show-Usage {
    [Console]::Error.WriteLine(@"
Usage: pull_with_release.ps1

Pull the current SpecFlow branch from origin.
Then make sure the current platform's specflowctl and specflow-reader binaries
match the pulled tooling source fingerprint. Missing or stale binaries are
downloaded from the matching GitHub Release.
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

function Get-PlatformSuffix {
    $os = ""
    if ([System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Windows)) {
        $os = "windows"
    }
    elseif ([System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Linux)) {
        $os = "linux"
    }
    elseif ([System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::OSX)) {
        $os = "darwin"
    }
    else {
        throw "Unsupported operating system."
    }

    $arch = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
        "X64" { "amd64" }
        "Arm64" { "arm64" }
        default { throw "Unsupported CPU architecture: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)" }
    }

    if ($os -eq "windows") {
        "$os-$arch.exe"
    }
    else {
        "$os-$arch"
    }
}

function Read-BinaryFingerprint {
    param(
        [string]$BinaryPath
    )

    if (-not (Test-Path -LiteralPath $BinaryPath -PathType Leaf)) {
        return ""
    }

    try {
        $output = & $BinaryPath "__print-build-fingerprint" 2>$null
        if ($LASTEXITCODE -ne 0) {
            return ""
        }
        return (($output -join "`n").Trim())
    }
    catch {
        return ""
    }
}

function Test-Checksums {
    param(
        [string]$Directory,
        [string]$CtlName,
        [string]$ReaderName
    )

    $sumsPath = Join-Path $Directory "SHA256SUMS"
    if (-not (Test-Path -LiteralPath $sumsPath -PathType Leaf)) {
        return $false
    }

    $expected = @{}
    foreach ($line in Get-Content -LiteralPath $sumsPath) {
        $parts = $line -split "\s+", 2
        if ($parts.Count -ne 2) {
            continue
        }
        $name = $parts[1].Trim()
        if ($name -eq $CtlName -or $name -eq $ReaderName) {
            $expected[$name] = $parts[0].Trim().ToLowerInvariant()
        }
    }

    if (-not $expected.ContainsKey($CtlName) -or -not $expected.ContainsKey($ReaderName)) {
        return $false
    }

    foreach ($name in @($CtlName, $ReaderName)) {
        $path = Join-Path $Directory $name
        if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
            return $false
        }
        $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $path).Hash.ToLowerInvariant()
        if ($actual -ne $expected[$name]) {
            return $false
        }
    }

    return $true
}

function Test-NeedsDownload {
    param(
        [string]$ExpectedFingerprint,
        [string]$CtlPath,
        [string]$ReaderPath,
        [string]$BinDir,
        [string]$CtlName,
        [string]$ReaderName
    )

    $ctlFingerprint = Read-BinaryFingerprint $CtlPath
    $readerFingerprint = Read-BinaryFingerprint $ReaderPath
    if ($ctlFingerprint -ne $ExpectedFingerprint) {
        return $true
    }
    if ($readerFingerprint -ne $ExpectedFingerprint) {
        return $true
    }
    if (-not (Test-Checksums $BinDir $CtlName $ReaderName)) {
        return $true
    }

    return $false
}

if ($Help) {
    Show-Usage
    exit 0
}

$scriptDir = Split-Path -Parent $PSCommandPath
$repoRoot = (Resolve-Path (Join-Path $scriptDir "../..")).Path
$binDir = Join-Path $repoRoot "tooling/bin"
$fingerprintRoot = $null
$downloadDir = $null

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

    $fingerprintRoot = New-FingerprintRoot $repoRoot
    $fingerprintScript = Join-Path $fingerprintRoot.Path "specflow/tooling/scripts/tooling_fingerprint.ps1"
    $fingerprint = (& $fingerprintScript).Trim()
    $shortFingerprint = $fingerprint.Substring(0, 12)
    $tag = "specflow-tooling-$shortFingerprint"
    $suffix = Get-PlatformSuffix
    $ctlName = "specflowctl-$suffix"
    $readerName = "specflow-reader-$suffix"
    $ctlPath = Join-Path $binDir $ctlName
    $readerPath = Join-Path $binDir $readerName

    if (-not (Test-NeedsDownload $fingerprint $ctlPath $readerPath $binDir $ctlName $readerName)) {
        Write-Host "Local binaries already match $tag."
        exit 0
    }

    & git ls-remote --exit-code --tags origin "refs/tags/$tag" *> $null
    if ($LASTEXITCODE -ne 0) {
        throw "Release tag does not exist on origin: $tag. Run push_with_release.ps1 on main first, then run this pull script again."
    }

    $downloadDir = Join-Path ([System.IO.Path]::GetTempPath()) ("specflow-download-" + [System.Guid]::NewGuid().ToString("N"))
    New-Item -ItemType Directory -Path $downloadDir | Out-Null
    $base = "https://github.com/Bingordinary/SpecFlow/releases/download/$tag"

    Write-Host "Downloading $tag binaries for $suffix..."
    Invoke-WebRequest -Uri "$base/$ctlName" -OutFile (Join-Path $downloadDir $ctlName)
    Invoke-WebRequest -Uri "$base/$readerName" -OutFile (Join-Path $downloadDir $readerName)
    Invoke-WebRequest -Uri "$base/SHA256SUMS" -OutFile (Join-Path $downloadDir "SHA256SUMS")

    if (-not (Test-Checksums $downloadDir $ctlName $readerName)) {
        throw "Downloaded files failed checksum verification."
    }

    New-Item -ItemType Directory -Path $binDir -Force | Out-Null
    Move-Item -LiteralPath (Join-Path $downloadDir $ctlName) -Destination $ctlPath -Force
    Move-Item -LiteralPath (Join-Path $downloadDir $readerName) -Destination $readerPath -Force
    Move-Item -LiteralPath (Join-Path $downloadDir "SHA256SUMS") -Destination (Join-Path $binDir "SHA256SUMS") -Force

    if (-not $suffix.EndsWith(".exe")) {
        Invoke-CheckedNative "chmod" @("+x", $ctlPath, $readerPath)
    }

    Write-Host "Installed $ctlName, $readerName, and SHA256SUMS from $tag."
}
finally {
    if ($null -ne $fingerprintRoot -and $fingerprintRoot.Temporary) {
        Remove-Item -LiteralPath $fingerprintRoot.Path -Recurse -Force -ErrorAction SilentlyContinue
    }
    if ($null -ne $downloadDir) {
        Remove-Item -LiteralPath $downloadDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}
