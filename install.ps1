<#
.SYNOPSIS
  Install CodeX in one command.
.EXAMPLE
  irm https://raw.githubusercontent.com/Andrey147-ai/CodeX/main/install.ps1 | iex
.EXAMPLE
  .\install.ps1 -Version v0.23.0 -Dir "$env:TEMP\codex-test" -NoPath
#>
param(
  [string]$Version = "latest",
  [string]$Dir = (Join-Path $env:USERPROFILE ".codex\bin"),
  [switch]$NoPath
)

$ErrorActionPreference = "Stop"
$Repo = "Andrey147-ai/CodeX"

function Get-AssetUrl($ver) {
  if ($ver -eq "latest" -or $ver -eq "") {
    $rel = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
  } else {
    $rel = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/tags/$ver"
  }
  $a = $rel.assets | Where-Object { $_.name -eq "codex.exe" } | Select-Object -First 1
  if (-not $a) { throw "codex.exe not found in release $($rel.tag_name). Publish it via tag push (release.yml)." }
  return @{ url = $a.browser_download_url; tag = $rel.tag_name }
}

New-Item -ItemType Directory -Path $Dir -Force | Out-Null
$info = Get-AssetUrl $Version
$out = Join-Path $Dir "codex.exe"
Write-Host "Downloading CodeX $($info.tag) ..."
$ok = $false
for ($i = 1; $i -le 3 -and -not $ok; $i++) {
  try {
    Invoke-WebRequest -Uri $info.url -OutFile $out
    $ok = $true
  } catch {
    Write-Host "Retry $i/3 ..."
    Start-Sleep -Seconds $i
  }
}
if (-not $ok) { throw "Download failed after 3 tries: $($info.url)" }
& $out version
if (-not $NoPath) {
  $cur = [Environment]::GetEnvironmentVariable("Path", "User")
  if ($cur -notlike "*$Dir*") {
    [Environment]::SetEnvironmentVariable("Path", "$cur;$Dir", "User")
    Write-Host "Added to PATH (restart terminal): $Dir"
  }
}
Write-Host "OK: $out"
