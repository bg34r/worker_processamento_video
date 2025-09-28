#!/usr/bin/env pwsh
# Script para executar testes locais com cobertura e SonarQube

param(
    [switch]$SkipSonar,
    [switch]$OnlyTests,
    [string]$SonarToken = "",
    [string]$SonarUrl = "http://localhost:9000"
)

Write-Host "🧪 Executando testes com cobertura..." -ForegroundColor Cyan

# Verificar se LocalStack está rodando
try {
    $response = Invoke-WebRequest -Uri "http://localhost:4566/_localstack/health" -UseBasicParsing -TimeoutSec 5
    Write-Host "✅ LocalStack está rodando" -ForegroundColor Green
} catch {
    Write-Host "⚠️  LocalStack não detectado. Alguns testes podem falhar." -ForegroundColor Yellow
}

# Executar testes com cobertura
Write-Host "📊 Executando testes unitários..." -ForegroundColor Yellow
$testResult = go test ./internal/infrastructure/... -coverprofile=coverage.out -covermode=count -v

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Alguns testes falharam!" -ForegroundColor Red
    exit 1
} else {
    Write-Host "✅ Todos os testes passaram!" -ForegroundColor Green
}

# Gerar relatório de cobertura
Write-Host "📈 Gerando relatório de cobertura..." -ForegroundColor Yellow
$coverageOutput = go tool cover -func coverage
$coverageOutput | Out-File -FilePath coverage_report.txt -Encoding UTF8

# Extrair cobertura total
$totalLine = $coverageOutput | Where-Object { $_ -like "*total*" }
if ($totalLine) {
    $totalCoverage = ($totalLine -split '\s+')[2] -replace '%', ''
    Write-Host ""
    Write-Host "📊 RELATÓRIO DE COBERTURA:" -ForegroundColor Cyan
    Write-Host "═══════════════════════════" -ForegroundColor Cyan
    Write-Host "Cobertura Total: $totalCoverage%" -ForegroundColor White
    
    if ([double]$totalCoverage -ge 60) {
        Write-Host "✅ Cobertura adequada (≥60%)" -ForegroundColor Green
    } else {
        Write-Host "❌ Cobertura insuficiente (<60%)" -ForegroundColor Red
        Write-Host "💡 Meta: ≥60% de cobertura total" -ForegroundColor Yellow
    }
} else {
    Write-Host "⚠️  Não foi possível extrair cobertura total" -ForegroundColor Yellow
    $totalCoverage = "0"
}

# Gerar HTML de cobertura
Write-Host "🌐 Gerando relatório HTML..." -ForegroundColor Yellow
go tool cover "-html=coverage" "-o=coverage.html"
Write-Host "📁 Relatório HTML salvo em: coverage.html" -ForegroundColor Green

if ($OnlyTests) {
    Write-Host "🎉 Execução de testes concluída!" -ForegroundColor Green
    exit 0
}

# Executar SonarQube se não foi pulado
if (-not $SkipSonar) {
    Write-Host ""
    Write-Host "🔍 Executando análise SonarQube..." -ForegroundColor Cyan
    
    if ($SonarToken -eq "") {
        Write-Host "⚠️  SONAR_TOKEN não fornecido. Use -SonarToken ou configure a variável de ambiente." -ForegroundColor Yellow
        $SonarToken = $env:SONAR_TOKEN
    }
    
    if ($SonarToken -eq "") {
        Write-Host "❌ SonarQube Token necessário para análise" -ForegroundColor Red
        Write-Host "💡 Configure: `$env:SONAR_TOKEN ou use -SonarToken" -ForegroundColor Yellow
    } else {
        # Verificar se sonar-scanner está disponível
        try {
            & sonar-scanner --version | Out-Null
            
            # Executar SonarQube análise
            & sonar-scanner `
                "-Dsonar.projectKey=worker_processamento_video" `
                "-Dsonar.sources=." `
                "-Dsonar.host.url=$SonarUrl" `
                "-Dsonar.login=$SonarToken" `
                "-Dsonar.go.coverage.reportPaths=coverage.out"
                
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ Análise SonarQube concluída!" -ForegroundColor Green
            } else {
                Write-Host "❌ Falha na análise SonarQube" -ForegroundColor Red
            }
        } catch {
            Write-Host "❌ sonar-scanner não encontrado" -ForegroundColor Red
            Write-Host "💡 Instale o SonarQube Scanner: https://docs.sonarsource.com/sonarqube/latest/analyzing-source-code/scanners/sonarscanner/" -ForegroundColor Yellow
        }
    }
}

Write-Host ""
Write-Host "📊 RESUMO:" -ForegroundColor Magenta
Write-Host "═════════" -ForegroundColor Magenta
Write-Host "✅ Testes executados" -ForegroundColor Green
Write-Host "📈 Cobertura: $totalCoverage%" -ForegroundColor White
Write-Host "📁 Relatórios gerados:" -ForegroundColor Yellow
Write-Host "   • coverage.out (para SonarQube)" -ForegroundColor Gray
Write-Host "   • coverage.html (visualização)" -ForegroundColor Gray
Write-Host "   • coverage_report.txt (texto)" -ForegroundColor Gray

if (-not $SkipSonar -and $SonarToken -ne "") {
    Write-Host "🔍 Análise SonarQube executada" -ForegroundColor Green
}

Write-Host ""
Write-Host "🎉 Pipeline local concluído!" -ForegroundColor Green