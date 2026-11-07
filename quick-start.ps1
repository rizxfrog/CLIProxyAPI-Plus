# quick-start.ps1
# Downloads the latest Windows build of CLIProxyAPI (backend), the management
# control panel (frontend), and models.json from the rizxfrog GitHub forks,
# then runs the server. Everything lands next to this script:
#
#   bin/cli-proxy-api.exe        backend (latest release zip, checksum-verified)
#   static/@management.html      frontend panel (served at /management.html)
#   models.json                  model catalog (loaded via MODELS_FILE)
#   config.yaml                  generated once on first run (random API key)
#
# Existing local files are kept as-is and downloads are skipped; pass -Force
# to re-fetch everything.
#
# GitHub's REST API is deliberately not used: the latest release tag comes from
# the /releases/latest redirect, so anonymous runs never hit the API rate limit.
#
# Usage:
#   .\quick-start.ps1                 # download + run on port 8317
#   .\quick-start.ps1 -Port 8319      # custom port
#   .\quick-start.ps1 -Force          # re-download even if files exist
#   .\quick-start.ps1 -SkipRun        # download only

[CmdletBinding()]
param(
    [string]$Dir = "",
    [int]$Port = 8317,
    [string]$BackendRepo = "rizxfrog/CLIProxyAPI-Plus",
    [string]$FrontendRepo = "rizxfrog/Cli-Proxy-API-Management-Center",
    [string]$ModelsRepo = "rizxfrog/CLIProxyAPI-Plus",
    [string]$ModelsRef = "feature/dev",
    [switch]$Force,
    [switch]$SkipRun
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"
[Net.ServicePointManager]::SecurityProtocol =
    [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

function Download-RemoteFile {
    param([string]$Url, [string]$OutFile)
    Write-Host "  downloading: $Url"
    $web = New-Object System.Net.WebClient
    try {
        $web.DownloadFile($Url, $OutFile)
    } finally {
        $web.Dispose()
    }
}

# Resolves the newest release tag from GitHub's /releases/latest 302 target,
# which avoids the rate-limited REST API entirely.
function Get-LatestReleaseTag {
    param([string]$Repo)
    $request = [System.Net.HttpWebRequest]::Create("https://github.com/$Repo/releases/latest")
    $request.AllowAutoRedirect = $false
    $request.UserAgent = "quick-start.ps1"
    try {
        $response = $request.GetResponse()
    } catch [System.Net.WebException] {
        $response = $_.Exception.Response
        if (-not $response) { throw }
    }
    try {
        $status = [int]$response.StatusCode
        if ($status -lt 300 -or $status -gt 399) {
            throw "$Repo/releases/latest did not redirect (HTTP $status)."
        }
        $location = $response.Headers["Location"]
        if (-not $location) { throw "$Repo/releases/latest redirect has no Location header." }
        $absolute = (New-Object Uri([Uri]"https://github.com/$Repo/releases/latest", $location)).AbsoluteUri
        if ($absolute -notmatch '/releases/tag/([^/]+)$') {
            throw "unexpected redirect target for $Repo : $absolute"
        }
        return $Matches[1]
    } finally {
        $response.Close()
    }
}

function Format-Size {
    param([long]$Bytes)
    "{0:N1} MB" -f ($Bytes / 1MB)
}

if (-not $Dir) {
    if ($PSCommandPath) { $Dir = Split-Path -Parent $PSCommandPath } else { $Dir = (Get-Location).Path }
}
$Dir = (Resolve-Path -LiteralPath $Dir).Path
$binDir = Join-Path $Dir "bin"
$staticDir = Join-Path $Dir "static"
$downloadDir = Join-Path $Dir ".download"
New-Item -ItemType Directory -Force -Path $binDir, $staticDir | Out-Null

# ---------------------------------------------------------------- backend ----
Write-Host "`n[1/4] Backend release ($BackendRepo)"
$exePath = Join-Path $binDir "cli-proxy-api.exe"
if (-not $Force -and (Test-Path $exePath)) {
    Write-Host ("  bin\cli-proxy-api.exe present ({0}); skipping download (-Force to refresh)" -f (Format-Size (Get-Item $exePath).Length))
} else {
    New-Item -ItemType Directory -Force -Path $downloadDir | Out-Null
    $arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "aarch64" } else { "amd64" }
    $backendTag = Get-LatestReleaseTag $BackendRepo
    $backendVersion = $backendTag.TrimStart("v")
    $zipName = "CLIProxyAPI_${backendVersion}_windows_${arch}.zip"
    $zipUrl = "https://github.com/$BackendRepo/releases/download/$backendTag/$zipName"
    $zipPath = Join-Path $downloadDir $zipName
    Download-RemoteFile $zipUrl $zipPath
    $zipSize = (Get-Item $zipPath).Length

    $sumUrl = "https://github.com/$BackendRepo/releases/download/$backendTag/checksums.txt"
    $sumPath = Join-Path $downloadDir "checksums.txt"
    try {
        Download-RemoteFile $sumUrl $sumPath
        $line = Get-Content $sumPath | Where-Object { $_ -like "* $zipName" } | Select-Object -First 1
        $expected = if ($line) { ($line -split '\s+')[0] } else { $null }
        if ($expected) {
            $actual = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash.ToLowerInvariant()
            if ($actual -ne $expected.ToLowerInvariant()) {
                throw "SHA256 mismatch for ${zipName}: expected $expected, got $actual"
            }
            Write-Host "  checksum OK"
        } else {
            Write-Warning "  $zipName absent from checksums.txt; skipping verification"
        }
    } catch [System.Net.WebException] {
        Write-Warning "  checksums.txt unavailable; skipping verification"
    }

    Write-Host "  extracting to bin\"
    Expand-Archive -Path $zipPath -DestinationPath $binDir -Force
    Remove-Item -Recurse -Force $downloadDir
    if (-not (Test-Path $exePath)) { throw "cli-proxy-api.exe missing after extraction" }
    Write-Host ("  backend: $backendTag ({0})" -f (Format-Size $zipSize))
}

# --------------------------------------------------------------- frontend ----
Write-Host "`n[2/4] Management panel ($FrontendRepo)"
$panelPath = Join-Path $staticDir "@management.html"
if (-not $Force -and (Test-Path $panelPath) -and (Get-Item $panelPath).Length -ge 10000) {
    Write-Host ("  static\@management.html present ({0}); skipping download (-Force to refresh)" -f (Format-Size (Get-Item $panelPath).Length))
} else {
    $panelUrl = "https://github.com/$FrontendRepo/releases/latest/download/%40management.html"
    Download-RemoteFile $panelUrl $panelPath
    if ((Get-Item $panelPath).Length -lt 10000) { throw "@management.html looks truncated ($((Get-Item $panelPath).Length) bytes)" }
    Write-Host ("  panel: {0}" -f (Format-Size (Get-Item $panelPath).Length))
}

# ------------------------------------------------------------ models.json ----
Write-Host "`n[3/4] Model catalog ($ModelsRepo@$ModelsRef)"
$modelsPath = Join-Path $Dir "models.json"
if (-not $Force -and (Test-Path $modelsPath)) {
    Write-Host ("  models.json present ({0}); skipping download (-Force to refresh)" -f (Format-Size (Get-Item $modelsPath).Length))
} else {
    $modelsUrl = "https://raw.githubusercontent.com/$ModelsRepo/$ModelsRef/internal/registry/models/models.json"
    Download-RemoteFile $modelsUrl $modelsPath
}
try {
    $sections = ([System.IO.File]::ReadAllText($modelsPath, [System.Text.Encoding]::UTF8) | ConvertFrom-Json).PSObject.Properties.Name
} catch {
    throw "models.json is not valid JSON: $_"
}
Write-Host ("  catalog sections: {0}" -f ($sections -join ", "))

# ----------------------------------------------------------------- config ----
Write-Host "`n[4/4] Config"
$configPath = Join-Path $Dir "config.yaml"
if (-not (Test-Path $configPath)) {
    $apiKey = "sk-" + (New-Guid).ToString("N")
    $mgmtKey = "mk-" + (New-Guid).ToString("N")
    @(
        "server:"
        '  host: "127.0.0.1"'
        "  port: $Port"
        "management:"
        "  # WebUI login key. Without it the whole /v8/management API is not"
        "  # registered and the panel fails with a 404."
        "  secret-key: `"$mgmtKey`""
        "access:"
        "  api-keys:"
        "    - `"$apiKey`""
    ) | Set-Content -Path $configPath -Encoding ascii
    Write-Host "  generated config.yaml"
    Write-Host "    proxy API key:    $apiKey"
    Write-Host "    management key:   $mgmtKey  (use it to log in to the WebUI)"
} else {
    Write-Host "  keeping existing config.yaml"
}

# ------------------------------------------------------------------- run ----
if ($SkipRun) {
    Write-Host "`nDownload complete (-SkipRun)."
    return
}

$env:MODELS_FILE = $modelsPath
Write-Host "`nStarting cli-proxy-api.exe (MODELS_FILE=$modelsPath)"
Write-Host ("Management panel: http://127.0.0.1:{0}/management.html" -f $Port)
Write-Host "Press Ctrl+C to stop.`n"
Push-Location $Dir
try {
    & $exePath --config $configPath --no-browser
} finally {
    Pop-Location
}
