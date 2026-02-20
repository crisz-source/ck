## Comandos

- `ck version` - Mostra a versão do ck
- `ck pods` - Lista apenas pods com problema (CrashLoopBackOff, ImagePullBackOff, restarts > 5)
- `ck pods -n <namespace>` - Filtra por namespace
- `ck logs <pod> -n <namespace>` - Mostra logs de um pod
- `ck logs <pod> -n <namespace> -t 100` - Mostra últimas 100 linhas
- `ck describe <pod> -n <namespace>` - Mostra detalhes resumidos de um pod com eventos
- `ck exec <pod> -n <namespace> -- <cmd>` - Executa comando dentro do pod
- `ck top` - Lista pods ordenados por consumo de CPU
- `ck top -m` - Lista pods ordenados por consumo de memória
- `ck top -n <namespace>` - Filtra por namespace
- `ck workers -n <namespace>` - Mostra status dos workers do Supervisor
- `ck scan <imagem>` - Scan de vulnerabilidades em imagens Docker
- `ck scan <imagem> -s CRITICAL` - Filtra por severidade
- `ck ingress` - Lista ingresses com URLs
- `ck nodes` - Status dos nodes com CPU/memoria

Para ver todos os comandos e opções:
```bash
ck --help
ck <comando> --help
```

---

## Instalação

### Linux
```bash
# Baixa o binário
curl -L https://github.com/crisz-source/ck/releases/latest/download/ck-linux-amd64 -o ck

# Dá permissão de execução
chmod +x ck

# Move para o PATH
sudo mv ck /usr/local/bin/

# Verifica instalação
ck version
```

### macOS
```bash
# Baixa o binário
curl -L https://github.com/crisz-source/ck/releases/latest/download/ck-darwin-amd64 -o ck

# Dá permissão de execução
chmod +x ck

# Move para o PATH
sudo mv ck /usr/local/bin/

# Verifica instalação
ck version
```

### Windows
```powershell
# Baixa o binário (PowerShell)
Invoke-WebRequest -Uri "https://github.com/crisz-source/ck/releases/latest/download/ck-windows-amd64.exe" -OutFile "ck.exe"

# Move para uma pasta no PATH (exemplo: C:\Windows)
Move-Item ck.exe C:\Windows\

# Verifica instalação
ck version
```

### Compilar do código fonte

Requer Go 1.18+
```bash
git clone https://github.com/crisz-source/ck.git
cd ck
go build -o ck .
sudo mv ck /usr/local/bin/
```

---

## Exemplos de uso

### Listar pods com problema
```bash
# Todos os namespaces
ck pods

# Namespace específico
ck pods -n php-worker
```

### Ver logs
```bash
# Logs completos
ck logs meu-pod -n default

# Últimas 100 linhas
ck logs meu-pod -n default -t 100
```

### Entrar no pod
```bash
# Shell interativo
ck exec meu-pod -n default -- bash

# Comando específico
ck exec meu-pod -n default -- ls -la
```

### Ver consumo de recursos
```bash
# Por CPU (padrão)
ck top -n php-worker

# Por memória
ck top -n php-worker -m
```

### Scan de vulnerabilidades

Requer [Trivy](https://aquasecurity.github.io/trivy) instalado.
```bash
# Scan básico (HIGH e CRITICAL)
ck scan nginx:latest

# Apenas CRITICAL
ck scan nginx:latest -s CRITICAL

# Todas as severidades
ck scan minha-imagem:v1.0 -s CRITICAL,HIGH,MEDIUM,LOW
```
