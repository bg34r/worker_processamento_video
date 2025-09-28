# Script PowerShell para gerenciar workers de processamento de vídeo
# Uso: .\manage-workers.ps1 [start|stop|scale|status] [numero_workers]

param(
    [Parameter(Position=0)]
    [ValidateSet('start', 'stop', 'restart', 'status', 'logs', 'scale', 'monitor', 'help')]
    [string]$Command = 'help',
    
    [Parameter(Position=1)]
    [int]$Workers = 3,
    
    [Parameter(Position=1)]
    [string]$WorkerId
)

$ComposeFile = "docker-compose.yml"
$ProjectName = "video-processer-worker"

function Show-Help {
    Write-Host "🎬 Gerenciador de Video Processer Workers" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Uso: .\manage-workers.ps1 [comando] [opções]" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Comandos:" -ForegroundColor Green
    Write-Host "  start           - Inicia todos os workers (padrão: 3 workers)"
    Write-Host "  stop            - Para todos os workers"
    Write-Host "  restart         - Reinicia todos os workers"
    Write-Host "  status          - Mostra status dos workers"
    Write-Host "  logs [worker]   - Mostra logs dos workers (ex: logs -WorkerId 1)"
    Write-Host "  scale [num]     - Escala para um número específico de workers"
    Write-Host "  monitor         - Monitora a fila SQS"
    Write-Host ""
    Write-Host "Exemplos:" -ForegroundColor Yellow
    Write-Host "  .\manage-workers.ps1 start"
    Write-Host "  .\manage-workers.ps1 scale -Workers 5"
    Write-Host "  .\manage-workers.ps1 logs -WorkerId 2"
    Write-Host "  .\manage-workers.ps1 status"
    Write-Host ""
}

function Start-Workers {
    param([int]$NumWorkers = 3)
    
    Write-Host "🚀 Iniciando $NumWorkers workers de processamento de vídeo..." -ForegroundColor Green
    
    if ($NumWorkers -gt 3) {
        Write-Warning "⚠️  Escalando além dos 3 workers padrão definidos no docker-compose.yml"
        Write-Warning "   Será necessário usar docker-compose scale ou configurar workers adicionais"
    }
    
    try {
        docker-compose -p $ProjectName up -d
        Write-Host "✅ Workers iniciados!" -ForegroundColor Green
    }
    catch {
        Write-Error "❌ Erro ao iniciar workers: $_"
    }
}

function Stop-Workers {
    Write-Host "🛑 Parando todos os workers..." -ForegroundColor Yellow
    
    try {
        docker-compose -p $ProjectName down
        Write-Host "✅ Workers parados!" -ForegroundColor Green
    }
    catch {
        Write-Error "❌ Erro ao parar workers: $_"
    }
}

function Restart-Workers {
    Write-Host "🔄 Reiniciando workers..." -ForegroundColor Yellow
    
    try {
        docker-compose -p $ProjectName restart
        Write-Host "✅ Workers reiniciados!" -ForegroundColor Green
    }
    catch {
        Write-Error "❌ Erro ao reiniciar workers: $_"
    }
}

function Show-Status {
    Write-Host "📊 Status dos Workers:" -ForegroundColor Cyan
    Write-Host ""
    
    try {
        docker-compose -p $ProjectName ps
        Write-Host ""
        Write-Host "📈 Estatísticas de containers:" -ForegroundColor Cyan
        
        $containers = docker-compose -p $ProjectName ps -q
        if ($containers) {
            docker stats --no-stream --format "table {{.Container}}`t{{.CPUPerc}}`t{{.MemUsage}}`t{{.NetIO}}" $containers
        } else {
            Write-Warning "Nenhum worker em execução"
        }
    }
    catch {
        Write-Error "❌ Erro ao obter status: $_"
    }
}

function Show-Logs {
    param([string]$WorkerId)
    
    if ([string]::IsNullOrEmpty($WorkerId)) {
        Write-Host "📋 Logs de todos os workers:" -ForegroundColor Cyan
        docker-compose -p $ProjectName logs --tail=50 -f
    } else {
        Write-Host "📋 Logs do worker $WorkerId:" -ForegroundColor Cyan
        docker-compose -p $ProjectName logs --tail=50 -f "video-processer-worker-$WorkerId"
    }
}

function Scale-Workers {
    param([int]$TargetWorkers)
    
    if ($TargetWorkers -lt 1) {
        Write-Error "❌ Número de workers deve ser pelo menos 1"
        return
    }
    
    Write-Host "⚖️  Escalando para $TargetWorkers workers..." -ForegroundColor Yellow
    
    try {
        if ($TargetWorkers -le 3) {
            # Parar todos primeiro
            docker-compose -p $ProjectName down
            
            # Iniciar apenas os workers necessários
            $services = @()
            for ($i = 1; $i -le $TargetWorkers; $i++) {
                $services += "video-processer-worker-$i"
            }
            
            docker-compose -p $ProjectName up -d $services
            Write-Host "✅ Escalado para $TargetWorkers workers!" -ForegroundColor Green
        } else {
            Write-Warning "⚠️  Para mais de 3 workers, configure workers adicionais no docker-compose.yml"
            Write-Warning "   Ou use: docker-compose up --scale video-processer-worker=$TargetWorkers"
        }
    }
    catch {
        Write-Error "❌ Erro ao escalar workers: $_"
    }
}

function Monitor-Queue {
    Write-Host "📊 Monitorando fila SQS..." -ForegroundColor Cyan
    Write-Host "   (Esta funcionalidade requer configuração do AWS CLI)" -ForegroundColor Yellow
    Write-Host ""
    
    # Exemplo de monitoramento (requer AWS CLI configurado)
    try {
        $queueUrl = "http://localhost:4566/000000000000/video-processing-queue"
        Write-Host "🔍 Verificando mensagens na fila..." -ForegroundColor Yellow
        
        # Uncomment when AWS CLI is configured:
        # aws sqs get-queue-attributes --queue-url $queueUrl --attribute-names ApproximateNumberOfMessages --endpoint-url http://localhost:4566
        
        Write-Host "💡 Configure o AWS CLI para monitoramento automático da fila" -ForegroundColor Cyan
    }
    catch {
        Write-Warning "⚠️  AWS CLI não configurado ou LocalStack não disponível"
    }
}

# Função principal
switch ($Command.ToLower()) {
    'start' { Start-Workers -NumWorkers $Workers }
    'stop' { Stop-Workers }
    'restart' { Restart-Workers }
    'status' { Show-Status }
    'logs' { Show-Logs -WorkerId $WorkerId }
    'scale' { Scale-Workers -TargetWorkers $Workers }
    'monitor' { Monitor-Queue }
    default { Show-Help }
}