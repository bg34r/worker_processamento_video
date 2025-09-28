#!/bin/bash
# Script para executar testes locais com cobertura e SonarQube

set -e

# Parâmetros
SKIP_SONAR=false
ONLY_TESTS=false
SONAR_TOKEN="${SONAR_TOKEN:-}"
SONAR_URL="${SONAR_URL:-http://localhost:9000}"

# Parse argumentos
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-sonar)
            SKIP_SONAR=true
            shift
            ;;
        --only-tests)
            ONLY_TESTS=true
            shift
            ;;
        --sonar-token)
            SONAR_TOKEN="$2"
            shift 2
            ;;
        --sonar-url)
            SONAR_URL="$2"
            shift 2
            ;;
        --help)
            echo "Uso: $0 [--skip-sonar] [--only-tests] [--sonar-token TOKEN] [--sonar-url URL]"
            exit 0
            ;;
        *)
            echo "Argumento desconhecido: $1"
            exit 1
            ;;
    esac
done

echo "🧪 Executando testes com cobertura..."

# Verificar se LocalStack está rodando
if curl -s http://localhost:4566/_localstack/health > /dev/null 2>&1; then
    echo "✅ LocalStack está rodando"
else
    echo "⚠️  LocalStack não detectado. Alguns testes podem falhar."
fi

# Executar testes com cobertura
echo "📊 Executando testes unitários..."
if go test ./internal/infrastructure/... -coverprofile=coverage.out -covermode=count -v; then
    echo "✅ Todos os testes passaram!"
else
    echo "❌ Alguns testes falharam!"
    exit 1
fi

# Gerar relatório de cobertura
echo "📈 Gerando relatório de cobertura..."
go tool cover -func=coverage.out | tee coverage_report.txt

# Extrair cobertura total
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

echo ""
echo "📊 RELATÓRIO DE COBERTURA:"
echo "═══════════════════════════"
echo "Cobertura Total: ${TOTAL_COVERAGE}%"

if (( $(echo "${TOTAL_COVERAGE} >= 60" | bc -l) )); then
    echo "✅ Cobertura adequada (≥60%)"
else
    echo "❌ Cobertura insuficiente (<60%)"
    echo "💡 Meta: ≥60% de cobertura total"
fi

# Gerar HTML de cobertura
echo "🌐 Gerando relatório HTML..."
go tool cover -html=coverage.out -o coverage.html
echo "📁 Relatório HTML salvo em: coverage.html"

if [ "$ONLY_TESTS" = true ]; then
    echo "🎉 Execução de testes concluída!"
    exit 0
fi

# Executar SonarQube se não foi pulado
if [ "$SKIP_SONAR" = false ]; then
    echo ""
    echo "🔍 Executando análise SonarQube..."
    
    if [ -z "$SONAR_TOKEN" ]; then
        echo "❌ SonarQube Token necessário para análise"
        echo "💡 Configure: export SONAR_TOKEN=seu_token ou use --sonar-token"
    else
        # Verificar se sonar-scanner está disponível
        if command -v sonar-scanner > /dev/null 2>&1; then
            # Executar SonarQube análise
            sonar-scanner \
                -Dsonar.projectKey=worker_processamento_video \
                -Dsonar.sources=. \
                -Dsonar.host.url="$SONAR_URL" \
                -Dsonar.login="$SONAR_TOKEN" \
                -Dsonar.go.coverage.reportPaths=coverage.out
                
            if [ $? -eq 0 ]; then
                echo "✅ Análise SonarQube concluída!"
            else
                echo "❌ Falha na análise SonarQube"
            fi
        else
            echo "❌ sonar-scanner não encontrado"
            echo "💡 Instale o SonarQube Scanner: https://docs.sonarsource.com/sonarqube/latest/analyzing-source-code/scanners/sonarscanner/"
        fi
    fi
fi

echo ""
echo "📊 RESUMO:"
echo "═════════"
echo "✅ Testes executados"
echo "📈 Cobertura: ${TOTAL_COVERAGE}%"
echo "📁 Relatórios gerados:"
echo "   • coverage.out (para SonarQube)"
echo "   • coverage.html (visualização)"
echo "   • coverage_report.txt (texto)"

if [ "$SKIP_SONAR" = false ] && [ -n "$SONAR_TOKEN" ]; then
    echo "🔍 Análise SonarQube executada"
fi

echo ""
echo "🎉 Pipeline local concluído!"