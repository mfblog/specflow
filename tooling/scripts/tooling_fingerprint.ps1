param(
    [switch]$Short
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "../../..")).Path
$records = New-Object System.Collections.Generic.List[string]

function Add-GoTree {
    param(
        [string]$RelativeRoot
    )

    $absoluteRoot = Join-Path $repoRoot ($RelativeRoot -replace '/', [System.IO.Path]::DirectorySeparatorChar)
    if (-not (Test-Path -LiteralPath $absoluteRoot -PathType Container)) {
        throw "Required tooling source directory missing: $RelativeRoot"
    }

    Get-ChildItem -LiteralPath $absoluteRoot -Recurse -File -Filter "*.go" | ForEach-Object {
        $relativePath = $_.FullName.Substring($repoRoot.Length + 1).Replace('\', '/')
        $records.Add($relativePath)
    }
}

function Add-FileTree {
    param(
        [string]$RelativeRoot
    )

    $absoluteRoot = Join-Path $repoRoot ($RelativeRoot -replace '/', [System.IO.Path]::DirectorySeparatorChar)
    if (-not (Test-Path -LiteralPath $absoluteRoot -PathType Container)) {
        throw "Required tooling runtime directory missing: $RelativeRoot"
    }

    Get-ChildItem -LiteralPath $absoluteRoot -Recurse -File | ForEach-Object {
        $relativePath = $_.FullName.Substring($repoRoot.Length + 1).Replace('\', '/')
        $records.Add($relativePath)
    }
}

function Add-RequiredFile {
    param(
        [string]$RelativePath
    )

    $absolutePath = Join-Path $repoRoot ($RelativePath -replace '/', [System.IO.Path]::DirectorySeparatorChar)
    if (-not (Test-Path -LiteralPath $absolutePath -PathType Leaf)) {
        throw "Required tooling source file missing: $RelativePath"
    }
    $records.Add($RelativePath)
}

function Add-OptionalFile {
    param(
        [string]$RelativePath
    )

    $absolutePath = Join-Path $repoRoot ($RelativePath -replace '/', [System.IO.Path]::DirectorySeparatorChar)
    if (Test-Path -LiteralPath $absolutePath -PathType Leaf) {
        $records.Add($RelativePath)
    }
}

Add-GoTree "specflow/tooling/cmd"
Add-GoTree "specflow/tooling/internal"
Add-FileTree "specflow/tooling/reader/web"
Add-RequiredFile "specflow/tooling/go.mod"
Add-RequiredFile "specflow/tooling/manifest.tsv"
Add-OptionalFile "specflow/tooling/go.sum"

$sortedRecords = [string[]]$records.ToArray()
[System.Array]::Sort($sortedRecords, [System.StringComparer]::Ordinal)
$files = New-Object System.Collections.Generic.List[string]
$previous = $null
foreach ($record in $sortedRecords) {
    if ($null -eq $previous -or $record -cne $previous) {
        $files.Add($record)
    }
    $previous = $record
}

$stream = [System.IO.MemoryStream]::new()

try {
    foreach ($relativePath in $files) {
        $absolutePath = Join-Path $repoRoot ($relativePath -replace '/', [System.IO.Path]::DirectorySeparatorChar)
        $pathBytes = [System.Text.Encoding]::UTF8.GetBytes($relativePath)
        $contentBytes = [System.IO.File]::ReadAllBytes($absolutePath)

        $stream.Write($pathBytes, 0, $pathBytes.Length)
        $stream.WriteByte(0)
        $stream.Write($contentBytes, 0, $contentBytes.Length)
        $stream.WriteByte(0)
    }

    $sha256 = [System.Security.Cryptography.SHA256]::Create()
    try {
        $hashBytes = $sha256.ComputeHash($stream.ToArray())
    }
    finally {
        $sha256.Dispose()
    }
}
finally {
    $stream.Dispose()
}

$fingerprint = [System.BitConverter]::ToString($hashBytes).Replace("-", "").ToLowerInvariant()
if ($Short) {
    $fingerprint.Substring(0, 12)
}
else {
    $fingerprint
}
