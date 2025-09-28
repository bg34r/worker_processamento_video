# Script PowerShell para teste de carga - Upload múltiplo de vídeos
# Uso: .\test-stress.ps1 -NumRounds 2 -VideosFolder "C:\caminho\pasta"

param(
    [int]$NumRounds = 1,
    [string]$VideosFolder = "C:\Users\Bruno\Documents\FIAP\Hackaton\Videos Hacka\Samples",
    [string]$UploadUrl = "http://localhost:8094/videos/upload",
    [string]$Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJicnVub2FyYXVqbyIsImVtYWlsIjoiYnJ1bm9saW1hMDczQGdtYWlsLmNvbSIsInVzZXJfaWQiOjEsInJvbGUiOiJhZG1pbiIsImV4cCI6MTc1OTA5NzEwNSwiaWF0IjoxNzU5MDk1MzA1fQ.cmLljgtSyWb2NF-11_wTLnypabkjb7h_7biGI8ooI_A"
)

Write-Host "🚀 Iniciando teste de carga com vídeos da pasta:" -ForegroundColor Green
Write-Host "📁 Pasta: $VideosFolder" -ForegroundColor Cyan
Write-Host "🔄 Rodadas: $NumRounds" -ForegroundColor Cyan
Write-Host "🎯 Endpoint: $UploadUrl" -ForegroundColor Cyan
Write-Host ""

# Verificar se a pasta existe
if (!(Test-Path $VideosFolder)) {
    Write-Error "❌ Pasta não encontrada: $VideosFolder"
    exit 1
}

# Buscar todos os arquivos de vídeo na pasta
$videoFiles = Get-ChildItem -Path $VideosFolder -Filter "*.mp4" | Sort-Object Name
if ($videoFiles.Count -eq 0) {
    Write-Error "❌ Nenhum arquivo .mp4 encontrado em: $VideosFolder"
    exit 1
}

Write-Host "📹 Encontrados $($videoFiles.Count) vídeos na pasta:" -ForegroundColor Yellow
foreach ($video in $videoFiles) {
    $sizeKB = [math]::Round($video.Length / 1KB, 1)
    Write-Host "   • $($video.Name) ($sizeKB KB)" -ForegroundColor Gray
}
Write-Host ""

$totalUploads = $videoFiles.Count * $NumRounds
Write-Host "📊 Total de uploads que serão realizados: $totalUploads" -ForegroundColor Magenta
Write-Host ""

# Função para fazer upload individual
function Invoke-VideoUpload {
    param(
        [int]$Index,
        [string]$FilePath,
        [string]$FileName,
        [int]$Round,
        [string]$Url,
        [string]$AuthToken
    )
    
    $titulo = "Stress Test - $FileName (Round $Round)"
    $autor = "Load Test Worker #$Index"
    
    $startTime = Get-Date
    
    try {
        Write-Host "📤 [$Index] Uploading: $FileName (Round $Round)" -ForegroundColor Yellow
        
        # Executar curl upload
        $response = curl -X POST `
            -H "Authorization: Bearer $AuthToken" `
            -F "file=@$FilePath;type=video/mp4" `
            -F "titulo=$titulo" `
            -F "autor=$autor" `
            $Url 2>&1
        
        $endTime = Get-Date
        $duration = ($endTime - $startTime).TotalSeconds
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ [$Index] $FileName uploaded in $([math]::Round($duration, 2))s" -ForegroundColor Green
            
            # Tentar extrair ID do vídeo da resposta JSON
            try {
                $jsonResponse = $response | ConvertFrom-Json
                $videoId = $jsonResponse.id_video
                Write-Host "   📋 ID: $videoId" -ForegroundColor Gray
            } catch {
                Write-Host "   📋 Response: $response" -ForegroundColor Gray
            }
        } else {
            Write-Host "❌ [$Index] $FileName failed after $([math]::Round($duration, 2))s" -ForegroundColor Red
            Write-Host "   📋 Error: $response" -ForegroundColor Red
        }
        
        return @{
            Index = $Index
            FileName = $FileName
            Round = $Round
            Success = ($LASTEXITCODE -eq 0)
            Duration = $duration
            Response = $response
        }
    }
    catch {
        $endTime = Get-Date
        $duration = ($endTime - $startTime).TotalSeconds
        
        Write-Host "❌ [$Index] Exception uploading $FileName : $($_.Exception.Message)" -ForegroundColor Red
        
        return @{
            Index = $Index
            FileName = $FileName
            Round = $Round
            Success = $false
            Duration = $duration
            Response = $_.Exception.Message
        }
    }
}

# Registrar tempo de início do teste
$testStartTime = Get-Date

# Executar uploads em paralelo usando jobs
$jobs = @()
$uploadIndex = 1

for ($round = 1; $round -le $NumRounds; $round++) {
    Write-Host "🔄 Iniciando Rodada $round de $NumRounds..." -ForegroundColor Magenta
    
    foreach ($videoFile in $videoFiles) {
        $job = Start-Job -ScriptBlock ${function:Invoke-VideoUpload} -ArgumentList $uploadIndex, $videoFile.FullName, $videoFile.Name, $round, $UploadUrl, $Token
        $jobs += $job
        $uploadIndex++
        
        # Pequeno delay entre inicializações para evitar sobrecarga instantânea
        Start-Sleep -Milliseconds 300
    }
    
    # Delay maior entre rodadas
    if ($round -lt $NumRounds) {
        Write-Host "⏸️  Pausa entre rodadas..." -ForegroundColor Yellow
        Start-Sleep -Seconds 2
    }
}

Write-Host "⏳ Aguardando conclusão de $($jobs.Count) uploads..." -ForegroundColor Yellow
Write-Host ""

# Aguardar todos os jobs e coletar resultados
$results = @()
$completedJobs = 0
foreach ($job in $jobs) {
    $result = Receive-Job -Job $job -Wait
    $results += $result
    Remove-Job -Job $job
    $completedJobs++
    
    # Mostrar progresso a cada 5 jobs completados
    if ($completedJobs % 5 -eq 0) {
        Write-Host "📈 Progresso: $completedJobs/$($jobs.Count) uploads concluídos..." -ForegroundColor Cyan
    }
}

# Calcular estatísticas
$testEndTime = Get-Date
$totalTestTime = ($testEndTime - $testStartTime).TotalSeconds
$successCount = ($results | Where-Object { $_.Success }).Count
$failCount = $totalUploads - $successCount
$avgDuration = ($results | Measure-Object -Property Duration -Average).Average
$minDuration = ($results | Measure-Object -Property Duration -Minimum).Minimum
$maxDuration = ($results | Measure-Object -Property Duration -Maximum).Maximum

# Exibir relatório
Write-Host ""
Write-Host "📊 RELATÓRIO DO TESTE DE CARGA" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "⏱️  Tempo total do teste: $([math]::Round($totalTestTime, 2))s" -ForegroundColor White
Write-Host "� Vídeos únicos testados: $($videoFiles.Count)" -ForegroundColor White
Write-Host "🔄 Rodadas executadas: $NumRounds" -ForegroundColor White
Write-Host "📁 Total de uploads: $totalUploads" -ForegroundColor White
Write-Host "✅ Sucessos: $successCount ($([math]::Round(($successCount/$totalUploads)*100, 1))%)" -ForegroundColor Green
Write-Host "❌ Falhas: $failCount ($([math]::Round(($failCount/$totalUploads)*100, 1))%)" -ForegroundColor Red
Write-Host ""
Write-Host "⏲️  TEMPOS DE RESPOSTA:" -ForegroundColor Yellow
Write-Host "   Médio: $([math]::Round($avgDuration, 2))s" -ForegroundColor White
Write-Host "   Mínimo: $([math]::Round($minDuration, 2))s" -ForegroundColor Green
Write-Host "   Máximo: $([math]::Round($maxDuration, 2))s" -ForegroundColor Red
Write-Host ""

if ($successCount -gt 0) {
    $throughput = $successCount / $totalTestTime
    Write-Host "🚀 Throughput: $([math]::Round($throughput, 2)) uploads/segundo" -ForegroundColor Cyan
}

# Exibir detalhes dos uploads que falharam
if ($failCount -gt 0) {
    Write-Host ""
    Write-Host "❌ DETALHES DAS FALHAS:" -ForegroundColor Red
    $failures = $results | Where-Object { !$_.Success }
    foreach ($failure in $failures) {
        Write-Host "   [$($failure.Index)] $($failure.Response)" -ForegroundColor Red
    }
}

# Mostrar estatísticas por vídeo
if ($NumRounds -gt 1) {
    Write-Host ""
    Write-Host "📈 ESTATÍSTICAS POR VÍDEO:" -ForegroundColor Yellow
    $videoStats = $results | Where-Object { $_.Success } | Group-Object FileName | Sort-Object Name
    foreach ($stat in $videoStats) {
        $avgTime = ($stat.Group | Measure-Object Duration -Average).Average
        Write-Host "   📹 $($stat.Name): $($stat.Count)/$NumRounds sucessos (avg: $([math]::Round($avgTime, 2))s)" -ForegroundColor White
    }
}

Write-Host ""
Write-Host "💡 PRÓXIMOS PASSOS:" -ForegroundColor Magenta
Write-Host "   1. Verifique os logs dos workers: docker compose logs -f" -ForegroundColor White
Write-Host "   2. Monitore o S3: docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 video-upload-service-hackaton-localstack-1 aws --endpoint-url=http://localhost:4566 s3 ls s3://video-service-bucket/ --recursive" -ForegroundColor White
Write-Host "   3. Verifique arquivos processados: ls s3://video-service-bucket/processed/" -ForegroundColor White
Write-Host "   4. Monitore recursos dos containers: docker stats" -ForegroundColor White
Write-Host ""
Write-Host "🎉 Teste de carga com $($videoFiles.Count) vídeos diferentes concluído!" -ForegroundColor Green