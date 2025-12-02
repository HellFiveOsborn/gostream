package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Configuração
const (
	// Pasta física na RAM onde os arquivos ficam
	HlsDir     = "/dev/shm/hls"
	SegmentTime = "4"
	ListSize    = "5"
)

type Response struct {
	Status  string `json:"status"`
	Stream  string `json:"stream,omitempty"` // URL limpa .m3u8
	Message string `json:"message,omitempty"`
}

// Estrutura do Processo
type StreamProcess struct {
	Cmd        *exec.Cmd
	URLs       []string
	CurrentIdx int
	ID         string
	ManualStop bool
	mu         sync.Mutex
}

var (
	streams = make(map[string]*StreamProcess)
	mu      sync.Mutex
)

func main() {
	os.RemoveAll(HlsDir)
	os.MkdirAll(HlsDir, 0777)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", apiHandler)

	fmt.Printf("API Restreamer iniciada na porta %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	query := r.URL.Query()
	streamParam := query.Get("stream")
	stopParam := query.Get("stop")
	restartParam := query.Get("restart")

	// Determina o protocolo (http ou https) baseado no header (útil se usar Cloudflare/Proxy)
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	// Se não tiver stream param, apenas retorna erro se não for um comando de stop global (opcional)
	if streamParam == "" {
		json.NewEncoder(w).Encode(Response{Status: "error", Message: "Parâmetro ?stream= é obrigatório"})
		return
	}

	// 1. Gera o HASH (ID da stream)
	hash := fmt.Sprintf("%x", md5.Sum([]byte(streamParam)))

	// 2. Monta a URL Final Limpa (Ex: http://localhost/live/{HASH}/index.m3u8)
	publicStreamURL := fmt.Sprintf("%s://%s/live/%s/index.m3u8", scheme, r.Host, hash)

	mu.Lock()
	proc, exists := streams[hash]
	mu.Unlock()

	// --- STOP ---
	if stopParam == "true" {
		if exists {
			stopStream(hash, proc)
			json.NewEncoder(w).Encode(Response{Status: "stopped", Message: "Stream parada."})
		} else {
			json.NewEncoder(w).Encode(Response{Status: "error", Message: "Stream não encontrada."})
		}
		return
	}

	// --- RESTART ---
	if restartParam == "true" && exists {
		stopStream(hash, proc)
		exists = false // Força recriação abaixo
	}

	// --- VERIFICA SE JÁ RODA ---
	if exists {
		json.NewEncoder(w).Encode(Response{
			Status: "started", // Mantive "started" ou "running" conforme preferir
			Stream: publicStreamURL,
		})
		return
	}

	// --- INICIA NOVA STREAM ---
	urls := strings.Split(streamParam, ",")
	newProc := &StreamProcess{
		URLs:       urls,
		CurrentIdx: 0,
		ID:         hash,
		ManualStop: false,
	}

	outputDir := filepath.Join(HlsDir, hash)
	os.MkdirAll(outputDir, 0777)

	go startFFmpeg(newProc)

	mu.Lock()
	streams[hash] = newProc
	mu.Unlock()

	// Delay para garantir criação do arquivo
	time.Sleep(3 * time.Second)

	// Retorno JSON limpo
	json.NewEncoder(w).Encode(Response{
		Status: "started",
		Stream: publicStreamURL,
	})
}

func startFFmpeg(proc *StreamProcess) {
	proc.mu.Lock()
	if proc.ManualStop {
		proc.mu.Unlock()
		return
	}
	url := proc.URLs[proc.CurrentIdx]
	outputDir := filepath.Join(HlsDir, proc.ID)
	playlist := filepath.Join(outputDir, "index.m3u8")
	segment := filepath.Join(outputDir, "seg_%03d.ts")
	proc.mu.Unlock()

	log.Printf("[%s] Iniciando encode: %s", proc.ID, url)

	cmd := exec.Command("ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-reconnect", "1", "-reconnect_at_eof", "1", "-reconnect_streamed", "1", "-reconnect_delay_max", "2",
		"-i", url,
		"-vf", "scale=-2:480",
		"-r", "30",
		"-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency",
		"-b:v", "800k", "-maxrate", "1000k", "-bufsize", "2000k", "-g", "60",
		"-c:a", "aac", "-b:a", "64k", "-ar", "44100",
		"-f", "hls",
		"-hls_time", SegmentTime,
		"-hls_list_size", ListSize,
		"-hls_flags", "delete_segments+append_list",
		"-hls_segment_filename", segment,
		playlist,
	)

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	proc.mu.Lock()
	proc.Cmd = cmd
	proc.mu.Unlock()

	err := cmd.Run()

	proc.mu.Lock()
	isManualStop := proc.ManualStop
	proc.mu.Unlock()

	if isManualStop {
		return
	}

	log.Printf("[%s] Falha/Fim: %v. Tentando fallback...", proc.ID, err)

	// Fallback
	proc.mu.Lock()
	if proc.CurrentIdx < len(proc.URLs)-1 {
		proc.CurrentIdx++
		proc.mu.Unlock()
		time.Sleep(1 * time.Second)
		startFFmpeg(proc)
	} else {
		proc.CurrentIdx = 0 // Loop infinito nas URLs
		proc.mu.Unlock()
		time.Sleep(5 * time.Second)
		startFFmpeg(proc)
	}
}

func stopStream(id string, proc *StreamProcess) {
	proc.mu.Lock()
	proc.ManualStop = true
	if proc.Cmd != nil && proc.Cmd.Process != nil {
		syscall.Kill(-proc.Cmd.Process.Pid, syscall.SIGKILL)
	}
	proc.mu.Unlock()

	mu.Lock()
	delete(streams, id)
	mu.Unlock()

	os.RemoveAll(filepath.Join(HlsDir, id))
}
