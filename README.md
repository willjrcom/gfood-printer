# GFood Printer Agent

Agente local para gerenciamento de impressões térmicas (ESC/POS) via WebSocket e RabbitMQ.

## 🚀 Como Gerar os Executáveis (Build)

Este projeto utiliza Go e suporta compilação cruzada. Você pode gerar o executável para Windows, Linux ou macOS diretamente da sua máquina.

### 1. Pré-requisitos
- Go 1.20 ou superior instalado.

### 2. Gerar para Windows (64-bit)
Este é o formato mais comum para computadores que controlam impressoras térmicas.
```bash
GOOS=windows GOARCH=amd64 go build -o gfood-printer-x64.exe
```

### 3. Gerar para Windows (32-bit/x86)
Para máquinas Windows muito antigas ou sistemas de 32 bits:
```bash
GOOS=windows GOARCH=386 go build -o gfood-printer-x86.exe
```

### 4. Gerar para macOS
```bash
# Para Macs com Intel
GOOS=darwin GOARCH=amd64 go build -o gfood-printer-mac-intel

# Para Macs com Apple Silicon (M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -o gfood-printer-mac-arm
```

### 5. Gerar para Linux
```bash
GOOS=linux GOARCH=amd64 go build -o gfood-printer-linux
```

---

## 🛠️ Comandos Úteis

### Rodar em modo de desenvolvimento
```bash
go run .
```

### Limpar e atualizar dependências
```bash
go mod tidy
```

## 📝 Observações
- O executável gerado para Windows utiliza a API `winspool.drv` para enviar comandos RAW (ESC/POS).
- O executável gerado para Unix (Mac/Linux) utiliza o sistema `CUPS` com o flag `-o raw`.
