$baseUrl = "http://localhost:3000"

Write-Host "Testing Max Demo API..." -ForegroundColor Cyan

# 1. Get Token
try {
    $val = Invoke-RestMethod -Uri "$baseUrl/api/v1/test-token" -Method Get -ErrorAction Stop
} catch {
    Write-Host "[ERROR] Failed to connect to server at $baseUrl. Is 'zeno run src/main.zl' running?" -ForegroundColor Red
    exit 1
}

$token = $val.token
if (-not $token) {
     Write-Host "[ERROR] Token not found in response." -ForegroundColor Red
     exit 1
}

Write-Host "[OK] Token Acquired: $($token.Substring(0, 15))..." -ForegroundColor Green

$headers = @{
    Authorization = "Bearer $token"
}

# 2. Get Tasks
try {
    Write-Host "`n[TASKS] Fetching Tasks..." -ForegroundColor Yellow
    $tasksResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/tasks" -Method Get -Headers $headers
    if ($tasksResponse.data) {
        Write-Host "[OK] Success! Found $($tasksResponse.data.Count) tasks." -ForegroundColor Green
        $tasksResponse.data | Select-Object id, title, status, priority | Format-Table -AutoSize
    } else {
        Write-Host "[WARN] No data returned for tasks." -ForegroundColor Yellow
    }
} catch {
    Write-Host "[ERROR] Failed to fetch tasks: $_" -ForegroundColor Red
}

# 3. Get Teams
try {
    Write-Host "`n[TEAMS] Fetching Teams..." -ForegroundColor Yellow
    $teamsResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/teams" -Method Get -Headers $headers
    if ($teamsResponse.data) {
        Write-Host "[OK] Success! Found $($teamsResponse.data.Count) teams." -ForegroundColor Green
        $teamsResponse.data | Select-Object id, name, description | Format-Table -AutoSize
    } else {
        Write-Host "[WARN] No data returned for teams." -ForegroundColor Yellow
    }
} catch {
    Write-Host "[ERROR] Failed to fetch teams: $_" -ForegroundColor Red
}

Write-Host "`n[DONE] Test Complete!" -ForegroundColor Cyan
