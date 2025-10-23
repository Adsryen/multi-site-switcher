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
    Write-Host "请选择构建方式:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "  [1] Major Release (不兼容变更)" -ForegroundColor White
    Write-Host "  [2] Minor  Release (新增功能)" -ForegroundColor White
    Write-Host "  [3] Patch  Release (修复补丁)" -ForegroundColor White
    Write-Host "" 
    Write-Host "  [4] Just Build (仅打包, 不改版本)" -ForegroundColor White
    Write-Host "  [5] Generate GitHub Release (自动发布, 跳过构建)" -ForegroundColor White
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
    Write-Host "# 发生错误, 过程已中止.                         #" -ForegroundColor Red
    Write-Host "################################################" -ForegroundColor Red
    Write-Host ""
    Read-Host "按回车继续..." | Out-Null
}

function Show-Success {
    Write-Host ""
    Write-Host "流程完成." -ForegroundColor Green
}

# ---------------------------
# Utility
# ---------------------------
function Get-ManifestPath {
    $path = Join-Path $RepoRoot 'manifest.json'
    if (-not (Test-Path $path)) { throw "manifest.json 未找到" }
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
    $parts = $v.Split('-')[0].Split('.')
    if ($parts.Length -lt 3) { $parts = @($parts + (0..(2 - ($parts.Length - 1)))) }
    $major = [int]$parts[0]; $minor = [int]$parts[1]; $patch = [int]$parts[2]
    switch ($type) {
        'major' { $major++; $minor = 0; $patch = 0 }
        'minor' { $minor++; $patch = 0 }
        'patch' { $patch++ }
        default { }
    }
    return "$major.$minor.$patch"
}

function Sanitize-Name([string]$name) {
    if (-not $name) { return 'multi-site-switcher' }
    $n = $name.ToLower()
    $n = ($n -replace "[^a-z0-9\.-]+", "-").Trim('-')
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
    if (-not $pm) { return $true } # 没有 Node 工程则跳过

    Write-Host "检测到 package.json, 使用 $pm 进行构建..." -ForegroundColor Gray
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
        Write-Host "Node 构建失败: $_" -ForegroundColor Red
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

    # 复制到临时目录并排除不需要的文件夹/文件
    $excludeDirs = @('.git','node_modules','dist-zip','build-tmp','.vscode','.idea','coverage','.cache','.parcel-cache')
    $excludeFiles = @('.gitignore','build.ps1','*.log','npm-debug.log*','yarn-debug.log*','pnpm-debug.log*')

    # 使用 robocopy 进行拷贝（Windows 可用），忽略返回码差异
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
            Write-Host "git 提交失败（忽略继续）: $_" -ForegroundColor Yellow
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
    $choice = Get-UserChoice "输入 1-6" "4"
    switch ($choice) {
        '1' { $versionType = 'major'; break }
        '2' { $versionType = 'minor'; break }
        '3' { $versionType = 'patch'; break }
        '4' { $versionType = $null; break }
        '5' { $versionType = $null; $releaseOnly = $true; $autoNotes = $true; break }
        '6' { exit 0 }
        default { Write-Host "无效选项" -ForegroundColor Red; continue }
    }
    break
}

# 版本更新
$targetVersion = $null
if ($versionType) {
    Write-Host ""; Write-Host "你选择了 $versionType 发布，这将创建一个新的提交/标签。" -ForegroundColor Yellow
    $confirm = Get-UserChoice "确认执行? (y/n)" "Y"
    if ($confirm.ToLower() -ne 'y') { Write-Host "操作已取消" -ForegroundColor Yellow; exit 0 }

    Write-Host "更新版本号中..." -ForegroundColor Green
    try { $targetVersion = Update-Version-And-Commit $versionType } catch { Show-Error; exit 1 }
}

# 安装与构建（非 release-only）
if (-not $releaseOnly) {
    Write-Host ""; Write-Host "安装依赖与构建中..." -ForegroundColor Green
    $ok = Run-NodeBuild
    if (-not $ok) { Show-Error; exit 1 }

    try {
        if (-not $targetVersion) { $targetVersion = Read-Version }
        $manifest = Read-Manifest
        $name = Sanitize-Name([string]$manifest.name)
        $zipName = "$name-v$targetVersion.zip"
        $zipPath = Join-Path (Join-Path $RepoRoot 'dist-zip') $zipName
        Write-Host "打包为: $zipPath" -ForegroundColor Gray
        Prepare-Zip -zipPath $zipPath
        Write-Host "构建与打包完成!" -ForegroundColor Green
    } catch { Show-Error; exit 1 }
} else {
    Write-Host "Release-only 模式：跳过构建步骤。" -ForegroundColor Yellow
}

# Just Build 情况：直接结束（不创建 Release）
if (-not $versionType -and -not $releaseOnly) {
    Write-Host ""; Write-Host "Just Build 选中，跳过 GitHub Release。" -ForegroundColor Yellow
    Show-Success; exit 0
}

# 询问是否创建 Release（当已 bump 且非 release-only）
$shouldRelease = $false
if ($versionType -and -not $releaseOnly) {
    $ans = Get-UserChoice "现在创建 GitHub Release? (y/n)" "N"
    if ($ans.ToLower() -eq 'y') { $shouldRelease = $true }
}
if (-not $releaseOnly -and -not $shouldRelease) {
    Write-Host ""; Write-Host "跳过 GitHub Release。" -ForegroundColor Yellow
    Show-Success; exit 0
}

# 始终使用自动生成说明
if (-not $releaseOnly) { $autoNotes = $true }

# 校验 gh CLI
Write-Host "检查 GitHub CLI 安装..." -ForegroundColor Gray
if (-not (Test-GhInstalled)) {
    Write-Host ""; Write-Host "################################################" -ForegroundColor Red
    Write-Host "# 未检测到 GitHub CLI                          #" -ForegroundColor Red
    Write-Host "################################################" -ForegroundColor Red
    Write-Host ""
    Write-Host "请安装并登录 gh: https://cli.github.com/，然后运行 gh auth login" -ForegroundColor Yellow
    Write-Host "已跳过 Release 创建，但构建结果可用。" -ForegroundColor Green
    Show-Success; exit 0
}

# 读取版本与制品
try {
    if (-not $targetVersion) { $targetVersion = Read-Version }
    $manifest = Read-Manifest
    $name = Sanitize-Name([string]$manifest.name)
    $zipName = "$name-v$targetVersion.zip"
    $zipPath = Join-Path (Join-Path $RepoRoot 'dist-zip') $zipName
    if (-not (Test-Path $zipPath)) { throw "构建产物不存在: $zipPath" }
} catch { Write-Host "读取版本或构建产物失败: $_" -ForegroundColor Red; Show-Error; exit 1 }

$tagName = "v$targetVersion"

# 确保 tag 存在
if (Test-GitInstalled -and (Test-GitRepo)) {
    try {
        & git rev-parse -q --verify "refs/tags/$tagName" | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Host "创建标签 $tagName ..." -ForegroundColor Gray
            & git tag -a $tagName -m "Release $tagName"
        }

        Write-Host "推送提交与标签..." -ForegroundColor Gray
        & git push; if ($LASTEXITCODE -ne 0) { throw "git push failed" }
        & git push --tags; if ($LASTEXITCODE -ne 0) { throw "git push --tags failed" }
    } catch { Show-Error; exit 1 }
}

# 创建 Release
try {
    Write-Host "创建 GitHub Release 并上传 $zipName ..." -ForegroundColor Gray
    & gh release create $tagName $zipPath --title "Release $tagName" --generate-notes
    if ($LASTEXITCODE -ne 0) { throw "GitHub release creation failed" }
    Write-Host "GitHub Release 创建成功!" -ForegroundColor Green
    Show-Success; exit 0
} catch {
    Show-Error; exit 1
}
