# GFood Printer Agent

Agente local para gerenciamento de impressões térmicas (ESC/POS) via WebSocket e RabbitMQ.

---

## ▶️ Como Rodar

### 1. Configure o `config.json`
Crie um arquivo `config.json` na mesma pasta do executável:

```json
{
  "rabbitmq_url": "amqp://user:pass@host:5672/",
  "schema_name": "nome_do_schema",
  "api_url": "https://api.exemplo.com"
}
```

### 2. Execute o binário da sua plataforma

**macOS (Apple Silicon — M1/M2/M3):**
```bash
./gfood-printer-mac-arm
```

**macOS (Intel):**
```bash
./gfood-printer-mac-intel
```

**Linux:**
```bash
./gfood-printer-linux
```

**Windows:** dê duplo clique em `gfood-printer-x64.exe` ou execute no terminal:
```cmd
gfood-printer-x64.exe
```

---

## 🔨 Gerar Executáveis (Build)

### Opção 1 — Script automático (recomendado)
Gera todos os binários de uma vez na pasta `versions/`:
```bash
./build.sh
```

### Opção 2 — Manual por plataforma
```bash
# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o versions/gfood-printer-mac-arm

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o versions/gfood-printer-mac-intel

# Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o versions/gfood-printer-linux

# Linux ARM64 (Raspberry Pi)
GOOS=linux GOARCH=arm64 go build -o versions/gfood-printer-linux-arm64

# Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o versions/gfood-printer-x64.exe

# Windows 32-bit
GOOS=windows GOARCH=386 go build -o versions/gfood-printer-x86.exe
```

### Rodar em modo desenvolvimento
```bash
go run .
```

---

## 📝 Observações
- **Windows**: usa a API `winspool.drv` para enviar comandos RAW (ESC/POS).
- **macOS / Linux**: usa o sistema `CUPS` com o flag `-o raw`.
- Erros de impressão descartam a mensagem da fila imediatamente (sem retry).
