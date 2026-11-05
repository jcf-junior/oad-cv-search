param(
    [string]$ApiUrl = "http://localhost:8080",
    [string]$Dir = (Join-Path $PSScriptRoot "..\pdfs")
)

$ok = 0
$failed = 0

foreach ($file in Get-ChildItem -Path $Dir -Filter *.pdf) {
    $status = & curl.exe -s -o NUL -w "%{http_code}" -X POST "$ApiUrl/candidates" -F "file=@$($file.FullName)"
    if ($status -eq "201") {
        $ok++
        Write-Host "ok      $($file.Name)"
    } else {
        $failed++
        Write-Host "$status     $($file.Name)"
    }
}

Write-Host "uploaded: $ok, failed: $failed"
