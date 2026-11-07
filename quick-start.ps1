# quick-start.ps1
# Downloads the latest Windows build of CLIProxyAPI (backend), the management
# control panel (frontend), and models.json from the rizxfrog GitHub forks,
# then runs the server. Everything lands next to this script:
#
#   bin/cli-proxy-api.exe        backend (latest release zip, checksum-verified)
#   static/@management.html      frontend panel (served at /management.html)
#   models.json                  model catalog (loaded via MODELS_FILE)
#   config.yaml                  generated once on first run (random API keys)
#   .quick-start.json            versions recorded from the last successful fetch
#
# Update policy: each run probes the remote versions cheaply (no GitHub REST
# API, so anonymous runs never hit the rate limit) and re-downloads a component
# only when the remote version changed, the local file is missing or corrupt,
# or -Force is given. Probes and downloads degrade to "keep what is local" on
# network failure, so the script still runs offline once everything is cached.
#
# Usage:
#   .\quick-start.ps1                 # update if needed, then run on port 8317
#   .\quick-start.ps1 -Port 8319      # custom port
#   .\quick-start.ps1 -Force          # re-download everything regardless
#   .\quick-start.ps1 -SkipRun        # fetch only, do not start the server
#   .\quick-start.ps1 -NoCheck        # skip remote probes (pure local run)

[CmdletBinding()]
param(
    [string]$Dir = "",
    [int]$Port = 8317,
    [string]$BackendRepo = "rizxfrog/CLIProxyAPI-Plus",
    [string]$FrontendRepo = "rizxfrog/Cli-Proxy-API-Management-Center",
    [string]$ModelsRepo = "rizxfrog/CLIProxyAPI-Plus",
    [string]$ModelsRef = "feature/dev",
    [switch]$Force,
    [switch]$NoCheck,
    [switch]$SkipRun
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"
function Download-RemoteFile {
    param([string]$Url, [string]$OutFile)
    $tempPath = "$OutFile.tmp"
    Write-Host "  downloading: $Url"
    # HttpWebRequest streams to a temp file; WebClient.Timeout is protected
    # and cannot be set from PowerShell, so it is not used here.
    $request = [System.Net.HttpWebRequest]::Create($Url)
    $request.UserAgent = "quick-start.ps1"
    $request.Timeout = 60000
    $request.ReadWriteTimeout = 60000
    try {
        $response = $request.GetResponse()
        try {
            $stream = $response.GetResponseStream()
            try {
                $file = [System.IO.File]::Create($tempPath)
                try { $stream.CopyTo($file) } finally { $file.Dispose() }
            } finally { $stream.Dispose() }
        } finally { $response.Close() }
        if (Test-Path $OutFile) { Remove-Item -LiteralPath $OutFile -Force }
        Move-Item -LiteralPath $tempPath -Destination $OutFile
    } finally {
        if (Test-Path $tempPath) { Remove-Item -LiteralPath $tempPath -Force }
    }
}

function Close-Response {
    param($Response)
    if ($Response) { try { $Response.Close() } catch { } }
}

# Resolves the newest release tag from GitHub's /releases/latest redirect,
# which avoids the rate-limited REST API entirely.
function Get-LatestReleaseTag {
    param([string]$Repo)
    $request = [System.Net.HttpWebRequest]::Create("https://github.com/$Repo/releases/latest")
    $request.Method = "HEAD"
    $request.AllowAutoRedirect = $false
    $request.UserAgent = "quick-start.ps1"
    $request.Timeout = 60000
    $response = $null
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
        Close-Response $response
    }
}

# Reads the remote ETag with a HEAD request so an unchanged file never needs a
# body download. Returns $null when the host does not expose one.
function Get-RemoteETag {
    param([string]$Url)
    $request = [System.Net.HttpWebRequest]::Create($Url)
    $request.Method = "HEAD"
    $request.UserAgent = "quick-start.ps1"
    $request.Timeout = 60000
    $response = $null
    try {
        $response = $request.GetResponse()
    } catch [System.Net.WebException] {
        $response = $_.Exception.Response
        if (-not $response) { return $null }
    }
    try {
        if ([int]$response.StatusCode -ge 400) { return $null }
        $etag = $response.Headers["ETag"]
        if ($etag) { return $etag.Trim('"') }
        return $null
    } finally {
        Close-Response $response
    }
}

# Asks the installed binary for its own version, so backend freshness needs no
# download and no recorded state. Returns e.g. "2026.11.7" or $null.
function Get-InstalledBackendVersion {
    param([string]$ExePath)
    if (-not (Test-Path $ExePath)) { return $null }
    try {
        $line = (& $ExePath --version 2>&1 | Select-Object -First 1)
        if ($line -match 'Version:\s*([0-9][^,\s]*)') { return $Matches[1] }
    } catch { }
    return $null
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
$statePath = Join-Path $Dir ".quick-start.json"
New-Item -ItemType Directory -Force -Path $binDir, $staticDir | Out-Null

# ------------------------------------------------------------------ state ----
# Records the remote versions last fetched successfully. Used for the panel and
# the model catalog, which have no locally readable version marker.
$state = @{}
if (Test-Path $statePath) {
    try {
        $raw = [System.IO.File]::ReadAllText($statePath, [System.Text.Encoding]::UTF8)
        foreach ($prop in ($raw | ConvertFrom-Json).PSObject.Properties) { $state[$prop.Name] = $prop.Value }
    } catch {
        Write-Warning "  .quick-start.json unreadable; version tracking restarts on next fetch"
    }
}
function Save-State {
    param([hashtable]$Values)
    ($Values | ConvertTo-Json -Compress) | Set-Content -LiteralPath $statePath -Encoding ascii
}

# ---------------------------------------------------------------- backend ----
Write-Host "`n[1/4] Backend release ($BackendRepo)"
$exePath = Join-Path $binDir "cli-proxy-api.exe"
$backendTag = $null
if (-not $NoCheck) {
    try {
        $backendTag = Get-LatestReleaseTag $BackendRepo
    } catch {
        Write-Warning "  latest backend tag unreachable: $($_.Exception.Message)"
    }
}
$installedBackend = Get-InstalledBackendVersion $exePath
$backendUpToDate = $false
if ($Force) {
    Write-Host "  -Force: refreshing"
} elseif (-not $installedBackend) {
    Write-Host "  bin\cli-proxy-api.exe missing or unreadable"
} elseif ($NoCheck) {
    Write-Host ("  installed v$installedBackend; probe skipped (-NoCheck)")
    $backendUpToDate = $true
} elseif (-not $backendTag) {
    Write-Host "  remote version unknown; keeping installed v$installedBackend"
    $backendUpToDate = $true
} elseif ($installedBackend -eq $backendTag.TrimStart("v")) {
    Write-Host ("  up to date (v{0}, {1}); skipping download" -f $installedBackend, (Format-Size (Get-Item $exePath).Length))
    $backendUpToDate = $true
} else {
    Write-Host "  update available: v$installedBackend -> $backendTag"
}

if (-not $backendUpToDate) {
    New-Item -ItemType Directory -Force -Path $downloadDir | Out-Null
    $arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "aarch64" } else { "amd64" }
    $backendVersion = $backendTag.TrimStart("v")
    $zipName = "CLIProxyAPI_${backendVersion}_windows_${arch}.zip"
    $zipUrl = "https://github.com/$BackendRepo/releases/download/$backendTag/$zipName"
    $zipPath = Join-Path $downloadDir $zipName
    try {
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
        if (-not (Test-Path $exePath)) { throw "cli-proxy-api.exe missing after extraction" }
        $state["backendTag"] = $backendTag
        Save-State $state
        Write-Host ("  backend: $backendTag ({0})" -f (Format-Size $zipSize))
    } catch {
        if (Test-Path $exePath) {
            Write-Warning "  backend update failed ($($_.Exception.Message)); keeping the installed binary"
        } else {
            throw "backend download failed and no local binary exists: $($_.Exception.Message)"
        }
    } finally {
        if (Test-Path $downloadDir) { Remove-Item -Recurse -Force $downloadDir }
    }
}

# --------------------------------------------------------------- frontend ----
Write-Host "`n[2/4] Management panel ($FrontendRepo)"
$panelPath = Join-Path $staticDir "@management.html"
$panelTag = $null
if (-not $NoCheck) {
    try {
        $panelTag = Get-LatestReleaseTag $FrontendRepo
    } catch {
        Write-Warning "  latest panel tag unreachable: $($_.Exception.Message)"
    }
}
$installedPanel = $state["panelTag"]
$panelPresent = (Test-Path $panelPath) -and (Get-Item $panelPath).Length -ge 10000
$panelUpToDate = $false
if ($Force) {
    Write-Host "  -Force: refreshing"
} elseif (-not $panelPresent) {
    Write-Host "  static\@management.html missing or truncated"
} elseif ($NoCheck) {
    Write-Host ("  present ({0}); probe skipped (-NoCheck)" -f (Format-Size (Get-Item $panelPath).Length))
    $panelUpToDate = $true
} elseif (-not $panelTag) {
    Write-Host "  remote version unknown; keeping the local panel"
    $panelUpToDate = $true
} elseif ($installedPanel -eq $panelTag) {
    Write-Host ("  up to date ($panelTag, {0}); skipping download" -f (Format-Size (Get-Item $panelPath).Length))
    $panelUpToDate = $true
} elseif (-not $installedPanel) {
    # File exists but predates version tracking: adopt the current remote tag
    # instead of re-downloading 4 MB on the first upgraded run.
    Write-Host ("  installed panel has no recorded version; adopting $panelTag")
    $state["panelTag"] = $panelTag
    Save-State $state
    $panelUpToDate = $true
} else {
    Write-Host "  update available: $installedPanel -> $panelTag"
}

if (-not $panelUpToDate) {
    $panelUrl = "https://github.com/$FrontendRepo/releases/latest/download/%40management.html"
    try {
        Download-RemoteFile $panelUrl $panelPath
        if ((Get-Item $panelPath).Length -lt 10000) { throw "@management.html looks truncated ($((Get-Item $panelPath).Length) bytes)" }
        $state["panelTag"] = $panelTag
        Save-State $state
        Write-Host ("  panel: $panelTag ({0})" -f (Format-Size (Get-Item $panelPath).Length))
    } catch {
        if ($panelPresent) {
            Write-Warning "  panel update failed ($($_.Exception.Message)); keeping the local copy"
        } else {
            Write-Warning ("  panel download failed: {0}. The backend fetches it on the first /management.html visit." -f $_.Exception.Message)
        }
    }
}

# ------------------------------------------------------------ models.json ----
Write-Host "`n[3/4] Model catalog ($ModelsRepo@$ModelsRef)"
$modelsPath = Join-Path $Dir "models.json"
$modelsUrl = "https://raw.githubusercontent.com/$ModelsRepo/$ModelsRef/internal/registry/models/models.json"
$remoteEtag = $null
if (-not $NoCheck) {
    $remoteEtag = Get-RemoteETag $modelsUrl
}
$installedEtag = $state["modelsEtag"]
$modelsPresent = Test-Path $modelsPath
$modelsUpToDate = $false
if ($Force) {
    Write-Host "  -Force: refreshing"
} elseif (-not $modelsPresent) {
    Write-Host "  models.json missing"
} elseif ($NoCheck) {
    Write-Host ("  present ({0}); probe skipped (-NoCheck)" -f (Format-Size (Get-Item $modelsPath).Length))
    $modelsUpToDate = $true
} elseif (-not $remoteEtag) {
    Write-Host "  remote ETag unavailable; keeping the local catalog"
    $modelsUpToDate = $true
} elseif ($installedEtag -eq $remoteEtag) {
    Write-Host ("  up to date ({0}); skipping download" -f (Format-Size (Get-Item $modelsPath).Length))
    $modelsUpToDate = $true
} elseif (-not $installedEtag) {
    Write-Host "  installed catalog has no recorded version; adopting the remote ETag"
    $state["modelsEtag"] = $remoteEtag
    Save-State $state
    $modelsUpToDate = $true
} else {
    Write-Host "  update available: catalog changed upstream; refreshing"
}

if (-not $modelsUpToDate) {
    try {
        Download-RemoteFile $modelsUrl $modelsPath
        $state["modelsEtag"] = $remoteEtag
        Save-State $state
    } catch {
        if ($modelsPresent) {
            Write-Warning "  catalog update failed ($($_.Exception.Message)); keeping the local file"
        } else {
            throw "models.json download failed and no local copy exists: $($_.Exception.Message)"
        }
    }
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
    Write-Host "`nUp to date (-SkipRun)."
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
