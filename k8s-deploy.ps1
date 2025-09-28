#!/usr/bin/env pwsh
# Deploy para Kubernetes com Auto-Scaling Inteligente

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("deploy", "destroy", "status", "scale", "logs", "install-keda")]
    [string]$Action,
    
    [int]$Replicas = 3,
    [string]$Namespace = "video-processing",
    [switch]$UseKeda,
    [switch]$WatchLogs
)

function Install-KEDA {
    Write-Host "🎯 Instalando KEDA (Kubernetes Event-driven Autoscaling)..." -ForegroundColor Yellow
    
    # Instalar KEDA via Helm
    helm repo add kedacore https://kedacore.github.io/charts
    helm repo update
    helm install keda kedacore/keda --namespace keda-system --create-namespace
    
    Write-Host "✅ KEDA instalado com sucesso!" -ForegroundColor Green
}

function Deploy-Application {
    Write-Host "🚀 Fazendo deploy da aplicação no Kubernetes..." -ForegroundColor Green
    
    # Build da imagem Docker
    Write-Host "📦 Construindo imagem Docker..." -ForegroundColor Cyan
    docker build -t video-worker:latest .
    
    # Aplicar manifestos
    Write-Host "📋 Aplicando manifestos Kubernetes..." -ForegroundColor Cyan
    kubectl apply -f k8s/rbac.yaml
    kubectl apply -f k8s/configmap.yaml
    kubectl apply -f k8s/deployment.yaml
    kubectl apply -f k8s/service.yaml
    
    if ($UseKeda) {
        Write-Host "🎭 Aplicando KEDA Scaler..." -ForegroundColor Magenta
        kubectl apply -f k8s/keda-scaler.yaml
    } else {
        Write-Host "📊 Aplicando HPA tradicional..." -ForegroundColor Blue
        kubectl apply -f k8s/hpa.yaml
    }
    
    Write-Host "✅ Deploy concluído!" -ForegroundColor Green
    Get-ApplicationStatus
}

function Remove-Application {
    Write-Host "🗑️  Removendo aplicação do Kubernetes..." -ForegroundColor Red
    
    kubectl delete -f k8s/ --ignore-not-found=true
    kubectl delete namespace $Namespace --ignore-not-found=true
    
    Write-Host "✅ Aplicação removida!" -ForegroundColor Green
}

function Get-ApplicationStatus {
    Write-Host "📊 Status da aplicação:" -ForegroundColor Cyan
    Write-Host ""
    
    # Pods
    Write-Host "🔸 Pods:" -ForegroundColor Yellow
    kubectl get pods -n $Namespace -l app=video-worker
    
    Write-Host ""
    
    # HPA Status
    Write-Host "🔸 Horizontal Pod Autoscaler:" -ForegroundColor Yellow
    kubectl get hpa -n $Namespace 2>$null
    
    Write-Host ""
    
    # KEDA Status (se disponível)
    Write-Host "🔸 KEDA Scaled Objects:" -ForegroundColor Yellow
    kubectl get scaledobjects -n $Namespace 2>$null
    
    Write-Host ""
    
    # Services
    Write-Host "🔸 Services:" -ForegroundColor Yellow
    kubectl get services -n $Namespace
}

function Scale-Application {
    Write-Host "⚖️  Escalando aplicação para $Replicas replicas..." -ForegroundColor Blue
    
    kubectl scale deployment video-worker --replicas=$Replicas -n $Namespace
    
    Write-Host "✅ Scaling solicitado!" -ForegroundColor Green
    Get-ApplicationStatus
}

function Watch-Logs {
    Write-Host "📜 Monitorando logs dos workers..." -ForegroundColor Cyan
    
    if ($WatchLogs) {
        kubectl logs -f -l app=video-worker -n $Namespace --all-containers=true
    } else {
        kubectl logs -l app=video-worker -n $Namespace --all-containers=true --tail=100
    }
}

# Verificar se kubectl está disponível
if (-not (Get-Command kubectl -ErrorAction SilentlyContinue)) {
    Write-Error "❌ kubectl não encontrado. Instale o Kubernetes CLI primeiro."
    exit 1
}

# Executar ação solicitada
switch ($Action) {
    "install-keda" {
        Install-KEDA
    }
    
    "deploy" {
        Deploy-Application
    }
    
    "destroy" {
        Remove-Application
    }
    
    "status" {
        Get-ApplicationStatus
    }
    
    "scale" {
        Scale-Application
    }
    
    "logs" {
        Watch-Logs
    }
}

Write-Host ""
Write-Host "🎉 Operação '$Action' concluída!" -ForegroundColor Green