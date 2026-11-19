param(
    [Parameter(Mandatory = $true)]
    [string]$Version,
    [switch]$Tag
)

$ErrorActionPreference = "Stop"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$modules = Get-ChildItem (Join-Path $repoRoot "contrib") -Recurse -Filter "go.mod" | Sort-Object FullName

foreach ($mod in $modules) {
    $dir = $mod.DirectoryName
    $rel = $dir.Substring($repoRoot.Length + 1).Replace("\", "/")
    Write-Host ">>> Updating $rel -> github.com/ngq/gorp $Version"

    $content = Get-Content $mod.FullName -Raw
    $updated = [regex]::Replace($content, 'github\.com/ngq/gorp\s+v[0-9]+\.[0-9]+\.[0-9]+', "github.com/ngq/gorp $Version")
    if ($updated -ne $content) {
        Set-Content -Path $mod.FullName -Value $updated -NoNewline
    }

    Push-Location $dir
    try {
        & go mod tidy
        if ($LASTEXITCODE -ne 0) { throw "go mod tidy failed in $rel" }
        & go build ./...
        if ($LASTEXITCODE -ne 0) { throw "go build failed in $rel" }
        & go test ./... -race -count=1
        if ($LASTEXITCODE -ne 0) { throw "go test failed in $rel" }
    }
    finally {
        Pop-Location
    }

    if ($Tag) {
        $tagName = "$rel/$Version"
        Write-Host ">>> Tagging $tagName"
        & git tag -a $tagName -m "Release $tagName"
        if ($LASTEXITCODE -ne 0) { throw "git tag failed for $tagName" }
    }
}
