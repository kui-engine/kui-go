#requires -Version 5.1
[CmdletBinding()]
param(
    [ValidateSet('Debug', 'Release')]
    [string]$Configuration = 'Release',
    # 默认 Release 输出到 bin（与包内预编译产物同一目录），Debug 输出到 bin\debug，
    # 不覆盖标准 Release 产物。
    [string]$Output = '',
    # 核心运行时目录：放 kui_c.dll + kui.dll。默认依次取包内 native\bin（公开包只带 Release 核心）、
    # 对外仓库的 runtime\windows-x64\bin、仓库内开发的 ..\c\build\<配置>。
    [string]$RuntimeDir = '',
    # Full Profile 扩展运行时目录：kui_ext_*、kui_media_*、FFmpeg、pdfium、*.filamat。
    [string]$FullProfileRuntimeDir = ''
)

$ErrorActionPreference = 'Stop'
$SdkRoot = Split-Path -Parent $PSCommandPath
if (-not $Output) { $Output = if ($Configuration -eq 'Debug') { 'bin\debug' } else { 'bin' } }
$Output = if ([IO.Path]::IsPathRooted($Output)) { $Output } else { Join-Path $SdkRoot $Output }
if (-not $RuntimeDir) {
    # 依次找：SDK 包布局 native\bin；对外仓库（git clone kui-go）布局 runtime\windows-x64\bin；KUI 仓库内开发
    # ..\c\build\<配置>。包与对外仓库只有 Release 核心：C ABI 不跨边界传 CRT 对象，Go 的 Debug 构建（-N -l、
    # 保留 DWARF）同样链接、部署它。
    $runtimeCandidates = @(
        (Join-Path $SdkRoot 'native\bin'),
        (Join-Path $SdkRoot 'runtime\windows-x64\bin'),
        (Join-Path $SdkRoot ('..\c\build\' + $Configuration))
    )
    $RuntimeDir = $runtimeCandidates | Where-Object { Test-Path (Join-Path $_ 'kui_c.dll') } | Select-Object -First 1
    if (-not $RuntimeDir) {
        throw "找不到 C ABI 运行时 kui_c.dll（可用 -RuntimeDir 指定）；试过：$($runtimeCandidates -join '；')"
    }
}
$RuntimeDir = [IO.Path]::GetFullPath($RuntimeDir)
$coreRuntime = @('kui_c.dll', 'kui.dll')
foreach ($name in $coreRuntime) {
    if (-not (Test-Path (Join-Path $RuntimeDir $name))) { throw "缺少运行时：$RuntimeDir\$name" }
}
# 扩展与媒体插件只发 Release 版。它们和核心之间走纯 C 协议（kuiext_* / kui_media_*），
# 不跨 CRT 传 C++ 对象，所以 Debug 核心（/MDd）加载 Release 扩展是受支持的组合——
# 与 C++ SDK（bin/debug 只有 kui.dll）和 C SDK 的 kui_c_copy_runtime_dlls 相同。
# 解包后取 native\bin；对外仓库取 runtime\windows-x64\bin；仓库内开发取 C++ SDK staging 的 bin。
if (-not $FullProfileRuntimeDir) {
    $fullProfileCandidates = @(
        (Join-Path $SdkRoot 'native\bin'),
        (Join-Path $SdkRoot 'runtime\windows-x64\bin'),
        (Join-Path $SdkRoot '..\cpp\bin')
    )
    $FullProfileRuntimeDir = $fullProfileCandidates | Where-Object { Test-Path (Join-Path $_ 'kui_ext_3d.dll') } |
        Select-Object -First 1
    if (-not $FullProfileRuntimeDir) {
        throw "找不到 Full Profile 扩展运行时 kui_ext_3d.dll（可用 -FullProfileRuntimeDir 指定）；试过：$($fullProfileCandidates -join '；')"
    }
}
$FullProfileRuntimeDir = [IO.Path]::GetFullPath($FullProfileRuntimeDir)

$go = (Get-Command go -ErrorAction Stop).Source
$python = (Get-Command python -ErrorAction Stop).Source
$clang = (Get-Command clang -ErrorAction Stop).Source
$adapter = Join-Path $SdkRoot 'tools\clang_go_cgo.py'
$env:CGO_ENABLED = '1'
$env:CC = "`"$python`" `"$adapter`""
$env:CGO_CFLAGS = '-g0'
$env:KUI_C_TEST_RUNTIME = Join-Path $RuntimeDir 'kui_c.dll'
# Release 去掉符号、裁掉源码路径，产物最小；Debug 保留 DWARF、关闭优化与内联，
# 让 Delve 能单步、看局部变量（Go 侧的「Debug 构建」本来就是这个意思）。
# 外层 @() 不能省：if 表达式只产出一个元素时 PowerShell 会把它拆成标量字符串，
# 再用 @goBuildFlags 展开就成了逐个字符的参数。
$goBuildFlags = @(if ($Configuration -eq 'Debug') {
    '-gcflags=all=-N -l'
} else {
    '-trimpath', '-ldflags', '-s -w'
})

Push-Location $SdkRoot
try {
    & $go test ./...
    if ($LASTEXITCODE -ne 0) { throw "go test 失败：$LASTEXITCODE" }
    New-Item -ItemType Directory -Path $Output -Force | Out-Null
    $quickstarts = @(Get-ChildItem (Join-Path $SdkRoot 'examples') -Directory | Where-Object {
        Test-Path (Join-Path $_.FullName 'main.go')
    })
    if ($quickstarts.Count -ne 8) { throw "Go SDK 应有 8 个专项 examples，实际 $($quickstarts.Count)" }
    $quickstarts | ForEach-Object {
        if (Test-Path (Join-Path $_.FullName 'main.go')) {
            $target = Join-Path $Output ("kui_go_{0}.exe" -f $_.Name)
            & $go build @goBuildFlags -o $target ("./examples/{0}" -f $_.Name)
            if ($LASTEXITCODE -ne 0) { throw "Go example 编译失败：$($_.Name)" }
        }
    }
    # canonical 示例数只读包内生成的清单（产品清单 + API 巡览示例），不在脚本里写死。
    $canonicalManifest = [IO.File]::ReadAllText((Join-Path $SdkRoot 'examples\canonical\manifest.json'), [Text.Encoding]::UTF8) | ConvertFrom-Json
    $canonicalCount = [int]$canonicalManifest.count
    if ($canonicalCount -le 0 -or $canonicalCount -ne @($canonicalManifest.examples).Count) {
        throw "examples\canonical\manifest.json 的 count=$canonicalCount 与 examples 列表不一致"
    }
    $canonicalRoot = Join-Path $SdkRoot 'examples\canonical\generated'
    $canonicalSources = @(Get-ChildItem $canonicalRoot -Directory | Where-Object {
        Test-Path (Join-Path $_.FullName 'main.go')
    })
    $canonicalPages = @(Get-ChildItem (Join-Path $SdkRoot 'examples\canonical\assets\pages') -Filter '*.html' -File)
    if ($canonicalSources.Count -ne $canonicalCount -or $canonicalPages.Count -ne $canonicalCount) {
        throw "Go canonical examples 必须是 $canonicalCount sources/$canonicalCount pages，实际 $($canonicalSources.Count)/$($canonicalPages.Count)"
    }
    $canonicalBuild = Join-Path $Output '.canonical-build'
    if (Test-Path $canonicalBuild) { Remove-Item -Recurse -Force -LiteralPath $canonicalBuild }
    New-Item -ItemType Directory -Path $canonicalBuild -Force | Out-Null
    & $go build @goBuildFlags -o "$canonicalBuild\" ./examples/canonical/generated/...
    if ($LASTEXITCODE -ne 0) { throw "Go canonical examples 批量编译失败：$LASTEXITCODE" }
    $canonicalExecutables = @(Get-ChildItem $canonicalBuild -Filter '*.exe' -File)
    if ($canonicalExecutables.Count -ne $canonicalCount) {
        throw "Go canonical examples 应生成 $canonicalCount 个 EXE，实际 $($canonicalExecutables.Count)"
    }
    $canonicalExecutables | ForEach-Object {
        Copy-Item -Force $_.FullName (Join-Path $Output ("kui_go_canonical_{0}" -f $_.Name))
    }
    Remove-Item -Recurse -Force -LiteralPath $canonicalBuild
    Copy-Item -Recurse -Force (Join-Path $SdkRoot 'examples\canonical\assets') (Join-Path $Output 'assets')
    $repositoryRoot = [IO.Path]::GetFullPath((Join-Path $SdkRoot '..\..'))
    if (Test-Path -LiteralPath (Join-Path $repositoryRoot 'examples\html_loader\samples')) {
        $assetCopies = @(
            @((Join-Path $repositoryRoot 'examples\html_loader\samples'), (Join-Path $Output 'assets\native_samples')),
            @((Join-Path $repositoryRoot 'tests\script\html'), (Join-Path $Output 'assets\script_html')),
            @((Join-Path $repositoryRoot 'examples\native\animation\assets\pulse.json'), (Join-Path $Output 'assets\native_animation\pulse.json')),
            @((Join-Path $repositoryRoot 'resources\audio\sample.mp3'), (Join-Path $Output 'assets\native_media\sample.mp3')),
            @((Join-Path $repositoryRoot 'resources\video\sample.mp4'), (Join-Path $Output 'assets\native_media\sample.mp4')),
            @((Join-Path $repositoryRoot 'resources\3d models\stylized_tank.glb'), (Join-Path $Output 'assets\native_3d\stylized_tank.glb')),
            @((Join-Path $repositoryRoot 'examples\font_awesome_showcase\css'), (Join-Path $Output 'assets\pages\css')),
            @((Join-Path $repositoryRoot 'examples\font_awesome_showcase\webfonts'), (Join-Path $Output 'assets\pages\webfonts'))
        )
        foreach ($copy in $assetCopies) {
            if (Test-Path -LiteralPath $copy[0]) {
                New-Item -ItemType Directory -Path (Split-Path -Parent $copy[1]) -Force | Out-Null
                Copy-Item -Recurse -Force -LiteralPath $copy[0] -Destination $copy[1]
            }
        }
    }
    # 先铺 Release 扩展运行时（跳过核心两个 DLL），再放与本配置匹配的核心。
    # 以前是「运行时目录 + ..\cpp\bin 全量依次覆盖」：仓库内 Debug 构建会被 cpp\bin 的
    # Release kui.dll 盖掉 Debug 版；解包后 Debug 目录里又没有扩展，构建直接失败。
    $extensionFiles = @(Get-ChildItem -LiteralPath $FullProfileRuntimeDir -File | Where-Object {
        # FFmpeg 的许可文本与 FFmpeg DLL 同行（T-DEBT-SDK-WINDOWS-FFMPEG-GPL-BUILD）。
        ($_.Extension -eq '.dll' -or $_.Extension -eq '.filamat' -or $_.Name -like 'LICENSE.FFmpeg*.txt') -and
            $coreRuntime -notcontains $_.Name
    })
    foreach ($runtimeFile in $extensionFiles) {
        Copy-Item -Force -LiteralPath $runtimeFile.FullName -Destination $Output
    }
    foreach ($name in $coreRuntime) {
        Copy-Item -Force -LiteralPath (Join-Path $RuntimeDir $name) -Destination $Output
    }
    $quickBins = @(Get-ChildItem $Output -Filter 'kui_go_*.exe' -File | Where-Object { $_.Name -notlike 'kui_go_canonical_*' })
    $canonicalBins = @(Get-ChildItem $Output -Filter 'kui_go_canonical_*.exe' -File)
    if ($quickBins.Count -ne 8 -or $canonicalBins.Count -ne $canonicalCount) {
        throw "Go bin 门失败：quick=$($quickBins.Count), canonical=$($canonicalBins.Count)"
    }
    foreach ($runtimeName in @('kui_ext_3d.dll', 'kui_media_audio.dll')) {
        if (-not (Test-Path -LiteralPath (Join-Path $Output $runtimeName))) {
            throw "Go bin 运行时门失败：缺少 $runtimeName（扩展运行时目录：$FullProfileRuntimeDir）"
        }
    }
    foreach ($name in $coreRuntime) {
        $expected = (Get-FileHash -LiteralPath (Join-Path $RuntimeDir $name) -Algorithm SHA256).Hash
        $actual = (Get-FileHash -LiteralPath (Join-Path $Output $name) -Algorithm SHA256).Hash
        if ($expected -ne $actual) {
            throw "Go bin 运行时门失败：$name 不是 $Configuration 核心运行时（$RuntimeDir）"
        }
    }
    Write-Host "GO SDK BUILD PASS: configuration=$Configuration, go=$(& $go version), clang=$(& $clang --version | Select-Object -First 1), quick=8, canonical=$canonicalCount, core=$RuntimeDir, extensions=$FullProfileRuntimeDir, output=$Output" -ForegroundColor Green
} finally {
    Pop-Location
}
