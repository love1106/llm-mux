# Install myclaude - Claude Code wrapper for llm-mux (Windows)

$ErrorActionPreference = "Stop"

$LlmMuxUrl = if ($env:LLM_MUX_URL) { $env:LLM_MUX_URL } else { "http://localhost:8317" }

# Create Claude Code config to bypass onboarding
$claudeDir = Join-Path $env:USERPROFILE ".claude"
if (-not (Test-Path $claudeDir)) {
    New-Item -ItemType Directory -Path $claudeDir -Force | Out-Null
}

$claudeJson = Join-Path $env:USERPROFILE ".claude.json"
@'
{
  "hasCompletedOnboarding": true,
  "theme": "dark",
  "numStartups": 1,
  "installMethod": "npm"
}
'@ | Set-Content -Path $claudeJson -Encoding UTF8

Write-Host "Created $claudeJson"

# Create myclaude.cmd wrapper in a directory on PATH
$binDir = Join-Path $env:USERPROFILE ".llm-mux" "bin"
if (-not (Test-Path $binDir)) {
    New-Item -ItemType Directory -Path $binDir -Force | Out-Null
}

$cmdPath = Join-Path $binDir "myclaude.cmd"
@"
@echo off
set ANTHROPIC_BASE_URL=$LlmMuxUrl
set ANTHROPIC_API_KEY=sk-ant-api03-mock
claude %*
"@ | Set-Content -Path $cmdPath -Encoding ASCII

Write-Host "Created $cmdPath"

# Create PowerShell function in profile
$profileDir = Split-Path $PROFILE -Parent
if (-not (Test-Path $profileDir)) {
    New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
}

if (-not (Test-Path $PROFILE)) {
    New-Item -ItemType File -Path $PROFILE -Force | Out-Null
}

$profileContent = Get-Content -Path $PROFILE -Raw -ErrorAction SilentlyContinue
$functionMarker = "function myclaude"

if ($profileContent -and $profileContent.Contains($functionMarker)) {
    Write-Host "myclaude already in $PROFILE"
} else {
    $functionBlock = @"

# llm-mux Claude Code wrapper
function myclaude {
    `$env:ANTHROPIC_BASE_URL = "$LlmMuxUrl"
    `$env:ANTHROPIC_API_KEY = "sk-ant-api03-mock"
    claude @args
}
"@
    Add-Content -Path $PROFILE -Value $functionBlock
    Write-Host "Added myclaude to $PROFILE"
}

# Add bin dir to user PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$binDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$binDir", "User")
    $env:Path = "$env:Path;$binDir"
    Write-Host "Added $binDir to user PATH"
}

Write-Host ""
Write-Host "For PowerShell: restart your terminal or run:"
Write-Host "  . `$PROFILE"
Write-Host "Then: myclaude"
Write-Host ""
Write-Host "For CMD: restart your terminal, then: myclaude"
