# GoRestream

Servidor de restreaming HLS de alta performance escrito em Go, com Nginx como proxy reverso e FFmpeg para transcodificação.

## 🚀 Características

- **Restreaming HLS**: Converte qualquer stream de vídeo para HLS
- **Fallback Automático**: Suporte a múltiplas URLs de origem com failover
- **Transcodificação**: Redimensiona para 480p com bitrate otimizado
- **Baixa Latência**: Configurado com `ultrafast` e `zerolatency`
- **RAM Storage**: Segmentos armazenados em `/dev/shm` para máxima performance
- **API REST**: Interface simples para controle de streams

## 📋 Requisitos

- Docker e Docker Compose
- Mínimo 256MB RAM (recomendado 2GB para múltiplas streams)
- Acesso ao `/dev/shm` (memória compartilhada)

## 🐳 Deploy com Dokploy

### Opção 1: Deploy via Compose

1. No Dokploy, crie um novo projeto
2. Selecione **Compose** como tipo de deploy
3. Conecte seu repositório Git ou faça upload dos arquivos
4. O Dokploy irá detectar automaticamente o `docker-compose.yml`
5. Configure o domínio desejado no painel
6. Clique em **Deploy**

### Opção 2: Deploy via Dockerfile

1. No Dokploy, crie um novo projeto
2. Selecione **Application** como tipo de deploy
3. Conecte seu repositório Git
4. Configure:
   - **Build Type**: Dockerfile
   - **Port**: `80`
   - **Exposed Port**: `8080`
5. Clique em **Deploy**

### Configuração de Rede

O `docker-compose.yml` já está configurado para usar a rede `dokploy-network`. Se você usa uma rede diferente, ajuste no arquivo.

## ⚙️ Configuração

### Variáveis de Ambiente

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `TZ` | `America/Sao_Paulo` | Timezone do container |

### Portas

| Porta Interna | Porta Externa | Descrição |
|---------------|---------------|-----------|
| 80 | 8080 | Nginx (API + HLS) |

## 📡 API

### Iniciar Stream

```
GET /?stream=<URL_DA_STREAM>
```

**Parâmetros:**
- `stream` (obrigatório): URL da stream de origem. Suporta múltiplas URLs separadas por vírgula para fallback.

**Exemplo:**
```bash
curl "http://localhost:8080/?stream=https://exemplo.com/live/stream.m3u8"
```

**Resposta:**
```json
{
  "status": "started",
  "stream": "http://localhost:8080/live/abc123def456/index.m3u8"
}
```

### Iniciar Stream com Fallback

```bash
curl "http://localhost:8080/?stream=https://principal.com/live.m3u8,https://backup.com/live.m3u8"
```

### Parar Stream

```
GET /?stream=<URL_DA_STREAM>&stop=true
```

**Exemplo:**
```bash
curl "http://localhost:8080/?stream=https://exemplo.com/live/stream.m3u8&stop=true"
```

**Resposta:**
```json
{
  "status": "stopped",
  "message": "Stream parada."
}
```

### Reiniciar Stream

```
GET /?stream=<URL_DA_STREAM>&restart=true
```

**Exemplo:**
```bash
curl "http://localhost:8080/?stream=https://exemplo.com/live/stream.m3u8&restart=true"
```

## 🎥 Reprodução

A URL retornada pela API pode ser usada em qualquer player HLS:

- **VLC**: Mídia → Abrir Fluxo de Rede
- **hls.js**: Players web com suporte HLS
- **Video.js**: Player HTML5 com plugin HLS
- **FFplay**: `ffplay http://localhost:8080/live/{hash}/index.m3u8`

### Exemplo HTML com hls.js

```html
<script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
<video id="video" controls></video>
<script>
  var video = document.getElementById('video');
  var hls = new Hls();
  hls.loadSource('http://localhost:8080/live/{hash}/index.m3u8');
  hls.attachMedia(video);
</script>
```

## 🔧 Especificações Técnicas

### Transcodificação

- **Resolução**: 480p (scale=-2:480)
- **FPS**: 30
- **Codec Vídeo**: H.264 (libx264)
- **Preset**: ultrafast
- **Tune**: zerolatency
- **Bitrate Vídeo**: 800kbps (max 1000kbps)
- **Codec Áudio**: AAC
- **Bitrate Áudio**: 64kbps
- **Sample Rate**: 44100 Hz

### HLS

- **Duração Segmento**: 4 segundos
- **Playlist Size**: 5 segmentos
- **Flags**: delete_segments, append_list

## 🛠️ Desenvolvimento Local

```bash
# Clonar repositório
git clone https://github.com/HellFiveOsborn/gostream.git
cd gorestream

# Build e execução
docker-compose up --build

# Ou apenas build
docker build -t gorestream .
docker run -p 8080:80 -v /dev/shm:/dev/shm gorestream
```

## 📊 Monitoramento

O container possui healthcheck configurado que verifica a cada 30 segundos se o Nginx está respondendo.

### Logs

```bash
# Via Docker
docker logs -f gorestream

# Via Dokploy
# Acesse a aba "Logs" do serviço no painel
```

## ⚠️ Limitações

- Cada stream consome recursos de CPU para transcodificação
- Recomenda-se limitar o número de streams simultâneas baseado nos recursos disponíveis
- O `/dev/shm` precisa ter espaço suficiente para os segmentos HLS

## 📝 Licença

MIT License
