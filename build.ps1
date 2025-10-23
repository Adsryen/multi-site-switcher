param()

# ===========================
# Multi Site Switcher - Interactive Build Assistant (PowerShell)
# ===========================

# UTF-8 console
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8
[Console]::InputEncoding = [System.Text.Encoding]::UTF8
try { chcp 65001 | Out-Null } catch {}

# Flags
$releaseOnly = $false
$autoNotes = $false
$versionType = $null # major/minor/patch/null

$RepoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $RepoRoot

function Show-Menu {
    Write-Host "=================================================" -ForegroundColor Cyan
    Write-Host " Multi Site Switcher - Interactive Build Assistant" -ForegroundColor Cyan
    Write-Host "=================================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Please choose the build type:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "  [1] Major Release (breaking changes)" -ForegroundColor White
    Write-Host "  [2] Minor  Release (new features)" -ForegroundColor White
    Write-Host "  [3] Patch  Release (bug fixes)" -ForegroundColor White
    Write-Host ""
    Write-Host "  [4] Just Build (package only, no version bump)" -ForegroundColor White
    Write-Host "  [5] Generate GitHub Release (skip build, auto notes)" -ForegroundColor White
    Write-Host "  [6] Exit" -ForegroundColor White
    Write-Host ""
}

function Get-UserChoice {
    param([string]$Prompt, [string]$Default = "")
    if ($Default) {
        $userInput = Read-Host "$Prompt [$Default]"
        if ([string]::IsNullOrWhiteSpace($userInput)) { return $Default }
    } else {
        $userInput = Read-Host $Prompt
    }
    return $userInput
}

function Show-Error {
    Write-Host ""; Write-Host "################################################" -ForegroundColor Red
    Write-Host "# An error occurred. Process halted.           #" -ForegroundColor Red
    Write-Host "################################################" -ForegroundColor Red
    Write-Host ""
    Read-Host "Press Enter to continue..." | Out-Null
}

function Show-Success {
    Write-Host ""
    Write-Host "Process finished." -ForegroundColor Green
}

# ---------------------------
# Utility
# ---------------------------
function Get-ManifestPath {
    $path = Join-Path $RepoRoot 'manifest.json'
    if (-not (Test-Path $path)) { throw "manifest.json not found" }
    return $path
}

function Read-Manifest {
    $manifestPath = Get-ManifestPath
    return Get-Content $manifestPath -Raw | ConvertFrom-Json
}

function Write-Manifest($obj) {
    $manifestPath = Get-ManifestPath
    $json = $obj | ConvertTo-Json -Depth 64
    Set-Content -Encoding UTF8 -NoNewline -Path $manifestPath -Value $json
}

function Bump-Version([string]$v, [string]$type) {
    if (-not $v) { return '0.1.0' }
    $parts = ($v.Split('-')[0]).Split('.')
    while ($parts.Length -lt 3) { $parts += '0' }
    $major = [int]$parts[0]
    $minor = [int]$parts[1]
    $patch = [int]$parts[2]
    switch ($type) {
        'major' { $major++; $minor = 0; $patch = 0 }
        'minor' { $minor++; $patch = 0 }
        'patch' { $patch++ }
        default { }
    }
    return ('{0}.{1}.{2}' -f $major, $minor, $patch)
}

function Sanitize-Name([string]$name) {
    if (-not $name) { return 'multi-site-switcher' }
    $n = $name.ToLower()
    $n = ($n -replace '[^a-z0-9\.-]+', '-').Trim('-')
    if ([string]::IsNullOrWhiteSpace($n)) { $n = 'multi-site-switcher' }
    return $n
}

function Test-GitInstalled { try { & git --version | Out-Null; return $LASTEXITCODE -eq 0 } catch { return $false } }
function Test-GhInstalled { try { & gh --version | Out-Null; return $LASTEXITCODE -eq 0 } catch { return $false } }
function Test-GitRepo { try { & git rev-parse --is-inside-work-tree 2>$null | Out-Null; return $LASTEXITCODE -eq 0 } catch { return $false } }

function Get-PackageManager {
    if (-not (Test-Path (Join-Path $RepoRoot 'package.json'))) { return $null }
    if (Get-Command pnpm -ErrorAction SilentlyContinue) { return 'pnpm' }
    if (Get-Command yarn -ErrorAction SilentlyContinue) { return 'yarn' }
    if (Get-Command npm  -ErrorAction SilentlyContinue) { return 'npm' }
    return $null
}

function Run-NodeBuild {
    $pm = Get-PackageManager
    if (-not $pm) { return $true } # No Node project detected; skip

    Write-Host "Detected package.json, using $pm to build..." -ForegroundColor Gray
    $prevNoColor = $env:NO_COLOR; $prevForceColor = $env:FORCE_COLOR
    $env:NO_COLOR = '1'; $env:FORCE_COLOR = '0'
    try {
        if ($pm -eq 'pnpm') {
            Write-Host "Running pnpm install..." -ForegroundColor Gray
            & pnpm install; if ($LASTEXITCODE -ne 0) { throw "pnpm install failed" }
            Write-Host "Running pnpm run build..." -ForegroundColor Gray
            & pnpm run build; if ($LASTEXITCODE -ne 0) { throw "pnpm run build failed" }
        } elseif ($pm -eq 'yarn') {
            Write-Host "Running yarn..." -ForegroundColor Gray
            & yarn; if ($LASTEXITCODE -ne 0) { throw "yarn install failed" }
            Write-Host "Running yarn build..." -ForegroundColor Gray
            & yarn build; if ($LASTEXITCODE -ne 0) { throw "yarn build failed" }
        } else {
            Write-Host "Running npm ci / npm install..." -ForegroundColor Gray
            & npm ci; if ($LASTEXITCODE -ne 0) { & npm install; if ($LASTEXITCODE -ne 0) { throw "npm install failed" } }
            Write-Host "Running npm run build..." -ForegroundColor Gray
            & npm run build; if ($LASTEXITCODE -ne 0) { throw "npm run build failed" }
        }
        return $true
    } catch {
        Write-Host "Node build failed: $_" -ForegroundColor Red
        return $false
    } finally {
        if ($null -ne $prevNoColor) { $env:NO_COLOR = $prevNoColor } else { Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue }
        if ($null -ne $prevForceColor) { $env:FORCE_COLOR = $prevForceColor } else { Remove-Item Env:FORCE_COLOR -ErrorAction SilentlyContinue }
    }
}

function Prepare-Zip([string]$zipPath) {
    $tmpDir = Join-Path $RepoRoot 'build-tmp'
    if (Test-Path $tmpDir) { Remove-Item -Recurse -Force $tmpDir }
    New-Item -ItemType Directory -Path $tmpDir | Out-Null

    # Copy to temp folder and exclude unwanted directories/files
    $excludeDirs = @('.git','node_modules','dist-zip','build-tmp','.vscode','.idea','coverage','.cache','.parcel-cache')
    $excludeFiles = @('.gitignore','build.ps1','*.log','npm-debug.log*','yarn-debug.log*','pnpm-debug.log*')

    # Use robocopy for copying (Windows), ignore its return code semantics
    $xd = @(); foreach ($d in $excludeDirs) { $xd += @('/XD', (Join-Path $RepoRoot $d)) }
    $xf = @(); foreach ($f in $excludeFiles) { $xf += @('/XF', $f) }
    $args = @($RepoRoot, $tmpDir, '/E') + $xd + $xf
    & robocopy @args | Out-Null

    $destDir = Split-Path -Parent $zipPath
    if (-not (Test-Path $destDir)) { New-Item -ItemType Directory -Path $destDir | Out-Null }

    if (Test-Path $zipPath) { Remove-Item $zipPath -Force }
    Compress-Archive -Path (Join-Path $tmpDir '*') -DestinationPath $zipPath -Force

    Remove-Item -Recurse -Force $tmpDir
}

function Prepare-Zip-From([string]$sourceDir, [string]$zipPath) {
    $destDir = Split-Path -Parent $zipPath
    if (-not (Test-Path $destDir)) { New-Item -ItemType Directory -Path $destDir | Out-Null }
    if (Test-Path $zipPath) { Remove-Item $zipPath -Force }
    Compress-Archive -Path (Join-Path $sourceDir '*') -DestinationPath $zipPath -Force
}

function Update-Version-And-Commit([string]$type) {
    $manifest = Read-Manifest
    $old = [string]$manifest.version
    $new = Bump-Version $old $type
    $manifest.version = $new
    Write-Manifest $manifest

    if (Test-GitInstalled -and (Test-GitRepo)) {
        try {
            & git add (Get-ManifestPath)
            & git commit -m "chore: release v$new"
        } catch {
            Write-Host "git commit failed (ignored): $_" -ForegroundColor Yellow
        }
    }
    return $new
}

function Read-Version() {
    try { return [string](Read-Manifest).version } catch { return '0.1.0' }
}

# ---------------------------
# Main menu
# ---------------------------
while ($true) {
    Show-Menu
    $choice = Get-UserChoice "Enter 1-6" "4"
    switch ($choice) {
        '1' { $versionType = 'major'; break }
        '2' { $versionType = 'minor'; break }
        '3' { $versionType = 'patch'; break }
        '4' { $versionType = $null; break }
        '5' { $versionType = $null; $releaseOnly = $true; $autoNotes = $true; break }
        '6' { exit 0 }
        default { Write-Host "Invalid choice" -ForegroundColor Red; continue }
    }
    break
}

# Version bump
$targetVersion = $null
if ($versionType) {
    Write-Host ""; Write-Host "You selected a $versionType release. This will create a new commit and tag." -ForegroundColor Yellow
    $confirm = Get-UserChoice "Are you sure? (y/n)" "Y"
    if ($confirm.ToLower() -ne 'y') { Write-Host "Action cancelled." -ForegroundColor Yellow; exit 0 }

    Write-Host "Updating version..." -ForegroundColor Green
    try { $targetVersion = Update-Version-And-Commit $versionType } catch { Show-Error; exit 1 }
}

# Install dependencies and build (non release-only)
if (-not $releaseOnly) {
    Write-Host ""; Write-Host "Installing dependencies and building..." -ForegroundColor Green
    $ok = Run-NodeBuild
    if (-not $ok) { Show-Error; exit 1 }

    try {
        if (-not $targetVersion) { $targetVersion = Read-Version }
        $manifest = Read-Manifest
        $name = Sanitize-Name([string]$manifest.name)
        $zipName = "$name-v$targetVersion.zip"
        $zipPath = Join-Path (Join-Path $RepoRoot 'dist-zip') $zipName
        Write-Host "Packaging to: $zipPath" -ForegroundColor Gray
        $distDir = Join-Path $RepoRoot 'dist'
        if (Test-Path $distDir) {
            Write-Host "Packaging from dist: $distDir" -ForegroundColor Gray
            Prepare-Zip-From -sourceDir $distDir -zipPath $zipPath
        } else {
            Write-Host "Packaging from source (no dist found) ..." -ForegroundColor Gray
            Prepare-Zip -zipPath $zipPath
        }
        Write-Host "Build and packaging completed!" -ForegroundColor Green
    } catch { Show-Error; exit 1 }
} else {
    Write-Host "Release-only mode: skipping build step." -ForegroundColor Yellow
}

# Just Build case: end early (no GitHub Release)
if (-not $versionType -and -not $releaseOnly) {
    Write-Host ""; Write-Host "Just Build selected. Skipping GitHub Release." -ForegroundColor Yellow
    Show-Success; exit 0
}

# Ask whether to create Release (if bumped and not release-only)
$shouldRelease = $false
if ($versionType -and -not $releaseOnly) {
    $ans = Get-UserChoice "Create GitHub Release now? (y/n)" "N"
    if ($ans.ToLower() -eq 'y') { $shouldRelease = $true }
}
if (-not $releaseOnly -and -not $shouldRelease) {
    Write-Host ""; Write-Host "Skip GitHub Release." -ForegroundColor Yellow
    Show-Success; exit 0
}

# Always use auto-generated notes
if (-not $releaseOnly) { $autoNotes = $true }

# Check gh CLI
Write-Host "Checking GitHub CLI..." -ForegroundColor Gray
if (-not (Test-GhInstalled)) {
    Write-Host ""; Write-Host "################################################" -ForegroundColor Red
    Write-Host "# GitHub CLI not found                         #" -ForegroundColor Red
    Write-Host "################################################" -ForegroundColor Red
    Write-Host ""
    Write-Host "Install gh: https://cli.github.com/ and run: gh auth login" -ForegroundColor Yellow
    Write-Host "Build completed. Skipping GitHub Release creation." -ForegroundColor Green
    Show-Success; exit 0
}

# Read version and artifact
try {
    if (-not $targetVersion) { $targetVersion = Read-Version }
    $manifest = Read-Manifest
    $name = Sanitize-Name([string]$manifest.name)
    $zipName = "$name-v$targetVersion.zip"
    $zipPath = Join-Path (Join-Path $RepoRoot 'dist-zip') $zipName
    if (-not (Test-Path $zipPath)) { throw "Build artifact not found: $zipPath" }
} catch { Write-Host "Failed to read version or build artifact: $_" -ForegroundColor Red; Show-Error; exit 1 }

$tagName = "v$targetVersion"

# Ensure tag exists
if (Test-GitInstalled -and (Test-GitRepo)) {
    try {
        & git rev-parse -q --verify "refs/tags/$tagName" | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Creating tag $tagName ..." -ForegroundColor Gray
            & git tag -a $tagName -m "Release $tagName"
        }

        Write-Host "Pushing commits and tags..." -ForegroundColor Gray
        & git push; if ($LASTEXITCODE -ne 0) { throw "git push failed" }
        & git push --tags; if ($LASTEXITCODE -ne 0) { throw "git push --tags failed" }
    } catch { Show-Error; exit 1 }
}

# Create Release
try {
    Write-Host "Creating GitHub Release and uploading $zipName ..." -ForegroundColor Gray
    & gh release create $tagName $zipPath --title "Release $tagName" --generate-notes
    if ($LASTEXITCODE -ne 0) { throw "GitHub release creation failed" }
    Write-Host "GitHub Release created successfully!" -ForegroundColor Green
    Show-Success; exit 0
} catch {
    Show-Error; exit 1
}
