#!/bin/bash

# Script Bash para teste de carga - Upload múltiplo de vídeos
# Uso: ./test-stress.sh [num_rounds] [videos_folder]

NUM_ROUNDS=${1:-1}
VIDEOS_FOLDER=${2:-"/mnt/c/Users/Bruno/Documents/FIAP/Hackaton/Videos Hacka/Samples"}
UPLOAD_URL="http://localhost:8094/videos/upload"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJicnVub2FyYXVqbyIsImVtYWlsIjoiYnJ1bm9saW1hMDczQGdtYWlsLmNvbSIsInVzZXJfaWQiOjEsInJvbGUiOiJhZG1pbiIsImV4cCI6MTc1OTA4NTQwMCwiaWF0IjoxNzU5MDgzNjAwfQ.i_5UcUw9A7kp0eUdyrfKZBHPNE_C2GO7-CLmrhCDty8"

echo "🚀 Iniciando teste de carga com vídeos da pasta:"
echo "📁 Pasta: $VIDEOS_FOLDER"
echo "🔄 Rodadas: $NUM_ROUNDS"
echo "🎯 Endpoint: $UPLOAD_URL"
echo ""

# Verificar se a pasta existe
if [ ! -d "$VIDEOS_FOLDER" ]; then
    echo "❌ Pasta não encontrada: $VIDEOS_FOLDER"
    exit 1
fi

# Buscar todos os arquivos de vídeo na pasta
VIDEO_FILES=($(find "$VIDEOS_FOLDER" -name "*.mp4" | sort))
NUM_VIDEOS=${#VIDEO_FILES[@]}

if [ $NUM_VIDEOS -eq 0 ]; then
    echo "❌ Nenhum arquivo .mp4 encontrado em: $VIDEOS_FOLDER"
    exit 1
fi

echo "📹 Encontrados $NUM_VIDEOS vídeos na pasta:"
for video in "${VIDEO_FILES[@]}"; do
    filename=$(basename "$video")
    size=$(du -h "$video" | cut -f1)
    echo "   • $filename ($size)"
done
echo ""

TOTAL_UPLOADS=$((NUM_VIDEOS * NUM_ROUNDS))
echo "📊 Total de uploads que serão realizados: $TOTAL_UPLOADS"
echo ""

# Arrays para estatísticas
declare -a RESULTS_SUCCESS
declare -a RESULTS_DURATION
declare -a RESULTS_FILENAME

# Função para upload individual
upload_video() {
    local index=$1
    local video_path=$2
    local filename=$3
    local round=$4
    
    local titulo="Stress Test - $filename (Round $round)"
    local autor="Load Test Worker #$index"
    
    echo "📤 [$index] Uploading: $filename (Round $round)"
    
    start_time=$(date +%s.%N)
    
    response=$(curl -s -X POST \
        -H "Authorization: Bearer $TOKEN" \
        -F "file=@$video_path;type=video/mp4" \
        -F "titulo=$titulo" \
        -F "autor=$autor" \
        "$UPLOAD_URL" 2>&1)
    
    curl_exit_code=$?
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0")
    
    # Armazenar resultados
    RESULTS_SUCCESS[$index]=$curl_exit_code
    RESULTS_DURATION[$index]=$duration
    RESULTS_FILENAME[$index]=$filename
    
    if [ $curl_exit_code -eq 0 ]; then
        printf "✅ [%d] %s uploaded in %.2fs\n" $index "$filename" "$duration"
        # Tentar extrair ID do JSON
        video_id=$(echo "$response" | grep -o '"id_video":"[^"]*"' | cut -d'"' -f4)
        if [ ! -z "$video_id" ]; then
            echo "   📋 ID: $video_id"
        fi
    else
        printf "❌ [%d] %s failed after %.2fs\n" $index "$filename" "$duration"
        echo "   📋 Error: $response"
    fi
}

# Registrar tempo de início
START_TIME=$(date +%s.%N)

# Executar uploads em paralelo
upload_index=1
for round in $(seq 1 $NUM_ROUNDS); do
    echo "🔄 Iniciando Rodada $round de $NUM_ROUNDS..."
    
    for video_path in "${VIDEO_FILES[@]}"; do
        filename=$(basename "$video_path")
        upload_video $upload_index "$video_path" "$filename" $round &
        upload_index=$((upload_index + 1))
        
        # Pequeno delay entre inicializações
        sleep 0.3
    done
    
    # Delay maior entre rodadas
    if [ $round -lt $NUM_ROUNDS ]; then
        echo "⏸️  Pausa entre rodadas..."
        sleep 2
    fi
done

echo "⏳ Aguardando conclusão de $TOTAL_UPLOADS uploads..."
echo ""

# Aguardar todos os processos em background
wait

# Calcular estatísticas
END_TIME=$(date +%s.%N)
TOTAL_TIME=$(echo "$END_TIME - $START_TIME" | bc -l)

SUCCESS_COUNT=0
TOTAL_DURATION=0
for i in $(seq 1 $TOTAL_UPLOADS); do
    if [ "${RESULTS_SUCCESS[$i]}" -eq 0 ] 2>/dev/null; then
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    fi
    if [ ! -z "${RESULTS_DURATION[$i]}" ]; then
        TOTAL_DURATION=$(echo "$TOTAL_DURATION + ${RESULTS_DURATION[$i]}" | bc -l)
    fi
done

FAIL_COUNT=$((TOTAL_UPLOADS - SUCCESS_COUNT))
AVG_DURATION=$(echo "scale=2; $TOTAL_DURATION / $TOTAL_UPLOADS" | bc -l)
SUCCESS_RATE=$(echo "scale=1; $SUCCESS_COUNT * 100 / $TOTAL_UPLOADS" | bc -l)
THROUGHPUT=$(echo "scale=2; $SUCCESS_COUNT / $TOTAL_TIME" | bc -l)

# Exibir relatório
echo ""
echo "📊 RELATÓRIO DO TESTE DE CARGA"
echo "═══════════════════════════════════════════════════"
printf "⏱️  Tempo total do teste: %.2fs\n" "$TOTAL_TIME"
echo "📹 Vídeos únicos testados: $NUM_VIDEOS"
echo "🔄 Rodadas executadas: $NUM_ROUNDS"
echo "📁 Total de uploads: $TOTAL_UPLOADS"
echo "✅ Sucessos: $SUCCESS_COUNT ($SUCCESS_RATE%)"
echo "❌ Falhas: $FAIL_COUNT"
echo ""
echo "⏲️  TEMPOS DE RESPOSTA:"
printf "   Médio: %.2fs\n" "$AVG_DURATION"
printf "🚀 Throughput: %.2f uploads/segundo\n" "$THROUGHPUT"
echo ""
echo "💡 PRÓXIMOS PASSOS:"
echo "   1. Verifique os logs dos workers: docker compose logs -f"
echo "   2. Monitore o S3: docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 video-upload-service-hackaton-localstack-1 aws --endpoint-url=http://localhost:4566 s3 ls s3://video-service-bucket/ --recursive"
echo "   3. Verifique arquivos processados: ls s3://video-service-bucket/processed/"
echo "   4. Monitore recursos dos containers: docker stats"
echo ""
echo "🎉 Teste de carga com $NUM_VIDEOS vídeos diferentes concluído!"