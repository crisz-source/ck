# ck - CLI para Kubernetes

**ck** = **C**risthian + **K**ubernetes

Uma ferramenta de linha de comando para facilitar o troubleshooting e operações do dia-a-dia no Kubernetes.

---

## Por que criar o ck?

O `kubectl` é poderoso, mas verboso. Para tarefas simples do dia-a-dia, precisamos digitar comandos longos e repetitivos.

O **ck** foi criado para:

- **Simplificar comandos frequentes** - menos digitação, mais produtividade
- **Formatar outputs** - informações organizadas e fáceis de ler
- **Filtrar o que importa** - mostrar apenas pods com problema, não todos
- **Unificar ferramentas** - kubectl + trivy + supervisorctl em um só lugar

---

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

---

## Sobre o comando `ck workers`

O comando `workers` foi criado para um caso de uso específico do meu ambiente de trabalho.

### Contexto

No sistema SUPP (Sistema Único de Processo e Protocolo), utilizamos pods PHP que rodam múltiplos workers gerenciados pelo **Supervisor** (um gerenciador de processos). Cada pod pode ter 50+ workers rodando simultaneamente, processando filas como:

- Indexação de documentos
- Envio de processos
- Sincronização com tribunais
- Processamento de relatórios
- E muitos outros...

### O problema

Verificar o status desses workers manualmente era trabalhoso:
```bash
# Antes: precisava entrar em cada pod e rodar supervisorctl
kubectl exec pod-xyz -n php-worker -- supervisorctl status
# Output gigante, difícil de ler
```

### A solução
```bash
# Agora: um comando mostra tudo
ck workers -n php-worker
```

Output:
```
=== WORKERS STATUS - NAMESPACE: php-worker ===

POD: php-worker-light-597447d7f8-wwt5l
  FATAL (45): indexacao_processo, download_processo, assistente_ia ... (+40 mais)
  RUNNING (14)
  Status: 14 OK | 45 FATAL | 0 STOPPED

POD: php-worker-heavy-8cf6d959b-d5fjp
  FATAL (51): indexacao_processo, populate_pessoa ... (+46 mais)
  RUNNING (8)
  Status: 8 OK | 51 FATAL | 0 STOPPED

==================================================
TOTAL GERAL: 22 OK | 96 FATAL | 0 STOPPED

ATENCAO: Workers em FATAL precisam de investigacao!
```

### Uso
```bash
# Todos os pods do namespace
ck workers -n php-worker

# Pod específico (mostra detalhes completos)
ck workers php-worker-light-xyz -n php-worker
```

> **Nota:** Este comando só funciona em pods que utilizam Supervisor. Em outros ambientes, ele mostrará "Erro: sem Supervisor".

---

## Estrutura do projeto
```
ck/
├── main.go                 # Entrada do programa
├── go.mod                  # Dependências Go
├── go.sum                  # Lock das dependências
├── build.sh                # Script para gerar binários
├── README.md               # Este arquivo
├── cmd/                    # Comandos da CLI
│   ├── root.go             # Comando raiz e variáveis globais
│   ├── version.go          # ck version
│   ├── pods.go             # ck pods
│   ├── logs.go             # ck logs
│   ├── describe.go         # ck describe
│   ├── exec.go             # ck exec
│   ├── top.go              # ck top
│   ├── workers.go          # ck workers
│   └── scan.go             # ck scan
├── types/                  # Structs compartilhadas
│   └── types.go            # Pod, Event, PodMetrics, etc
├── dist/                   # Binários compilados
│   ├── ck-linux-amd64      # Linux 64-bit
│   ├── ck-linux-arm64      # Linux ARM
│   ├── ck-darwin-amd64     # macOS Intel
│   └── ck-windows-amd64.exe # Windows
└── pratica/                # Exercícios de estudo (ver abaixo)
    ├── ponteiros.go
    ├── maps.go
    ├── loops.go
    ├── errors.go
    ├── bug1.go
    ├── bug2.go
    ├── bug3.go
    └── bug4.go
```

---

## Sobre a pasta `pratica/`

A pasta `pratica/` contém exercícios de estudo de Go criados durante o desenvolvimento do ck. São desafios de debug onde a IA gerava código com bugs propositais para eu encontrar e corrigir.

Exercícios incluem:

- **ponteiros.go** - Entender ponteiros e referências em Go
- **maps.go** - Trabalhar com maps e verificar existência de chaves
- **loops.go** - Loops com range e manipulação de slices
- **errors.go** - Tratamento de erros em Go
- **bug1.go a bug4.go** - Desafios de debug com níveis de dificuldade crescente

Esses exercícios ajudaram a fixar conceitos fundamentais de Go como ponteiros, slices, maps, structs e tratamento de erros.

---

## Requisitos

- `kubectl` configurado e com acesso ao cluster
- `trivy` instalado (apenas para o comando `ck scan`)

---


## Autor

Cristhian - 2026