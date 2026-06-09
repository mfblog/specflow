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

function Get-OSArchitecture {
    $runtimeInfo = [System.Runtime.InteropServices.RuntimeInformation]
    $property = $runtimeInfo.GetProperty("OSArchitecture")
    if ($null -ne $property) {
        return [string]$property.GetValue($null, $null)
    }

    $arch = [System.Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITEW6432")
    if ([string]::IsNullOrWhiteSpace($arch)) {
        $arch = [System.Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITECTURE")
    }
    if (-not [string]::IsNullOrWhiteSpace($arch)) {
        return $arch
    }

    throw "Unable to determine CPU architecture."
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

    $osArchitecture = Get-OSArchitecture
    $arch = switch ($osArchitecture.ToString().ToUpperInvariant()) {
        "X64" { "amd64" }
        "AMD64" { "amd64" }
        "ARM64" { "arm64" }
        default { throw "Unsupported CPU architecture: $osArchitecture" }
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
$projectRoot = (Resolve-Path (Join-Path $repoRoot "..")).Path
$binDir = Join-Path $repoRoot "tooling/bin"
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

    Sync-ExistingEntryBlocks $repoRoot $projectRoot

    $fingerprintScript = Join-Path $repoRoot "tooling/scripts/tooling_fingerprint.ps1"
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
    if ($null -ne $downloadDir) {
        Remove-Item -LiteralPath $downloadDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}
