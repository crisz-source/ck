# Changelog

Todas as mudanças notáveis do projeto ck serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/).

---

## [0.3.0] - 2026-02-23\

### Adicionado
- **`ck watch`** — Monitoramento de pods em tempo real usando Kubernetes Informers
  - Detecta restarts, CrashLoopBackOff, OOMKilled e pods deletados
  - Envia alertas por email via Azure Communication Services (REST API com HMAC-SHA256)
  - Alertas no terminal em tempo real com emojis indicando severidade
  - Threshold configurável para envio de email (`watch.restart_threshold`)
  - Graceful shutdown com captura de SIGINT/SIGTERM
- **Pacote `notify/`** — Sistema de notificações desacoplado
  - `notify/email.go` — Envio de email via Azure Communication Services
  - `notify.PodEvent` — Struct compartilhada para eventos de pod
  - `notify.PrintAlert()` — Exibição de alertas formatados no terminal
- **`k8s/watcher.go`** — Informer com EventHandlers (AddFunc, UpdateFunc, DeleteFunc)
  - Comparação de estado antigo vs novo para detectar restarts
  - Classificação automática de eventos (RESTART, CRASHLOOP, OOM_KILLED)
  - Resync period de 30 segundos para garantir consistência do cache

### Alterado
- `~/.ck.yaml` — Novas seções `watch` e `notify.email`
- Requisito mínimo de Go atualizado para 1.21+

---

## [0.2.0] - 2026-02-23

### Adicionado
- **Viper** — Gerenciamento de configuração com hierarquia de prioridade
  - Arquivo de configuração `~/.ck.yaml`
  - Variáveis de ambiente com prefixo `CK_` (ex: `CK_NAMESPACE`)
  - Hierarquia: flag > env > arquivo > default
  - Bind automático de flags com Viper
- **`ck config`** — Mostra configuração ativa e origem dos valores
- **`ck config path`** — Mostra caminho do arquivo de configuração
- **client-go** — Comunicação direta com a API do Kubernetes
  - `k8s/client.go` — Helper de conexão (kubeconfig + in-cluster)
  - `ck pods` reescrito usando client-go (sem depender de kubectl)
  - Suporte a autenticação in-cluster (para rodar dentro de pods)
- **Pacote `k8s/`** — Separação da lógica de conexão com Kubernetes

### Alterado
- `cmd/root.go` — Reescrito com Viper (removidas variáveis globais)
- `cmd/workers.go` — Migrado para usar `viper.GetString("namespace")`
- `cmd/pods.go` — Reescrito com client-go (acesso direto à API K8s)

### Removido
- Variável global `Namespace` (substituída por `viper.GetString("namespace")`)
- Dependência de `kubectl` no comando `pods` (agora usa client-go)

---

## [0.1.0] - 2026-02-23

### Adicionado
- **Estrutura inicial** do projeto com Cobra
- **`ck version`** — Mostra versão do ck
- **`ck pods`** — Lista pods com problema (CrashLoopBackOff, ImagePullBackOff, restarts > 5)
- **`ck logs`** — Mostra logs de um pod com flag `-t` para limitar linhas
- **`ck describe`** — Detalhes resumidos de um pod com eventos
- **`ck exec`** — Executa comando dentro de um pod
- **`ck top`** — Lista pods ordenados por consumo de CPU ou memória (`-m`)
- **`ck workers`** — Status dos workers do Supervisor (específico para SUPP)
- **`ck scan`** — Scan de vulnerabilidades em imagens Docker via Trivy
- **`ck ingress`** — Lista ingresses com URLs
- **`ck nodes`** — Status dos nodes com CPU/memória
- **Pasta `pratica/`** — Exercícios de estudo de Go (ponteiros, maps, loops, errors, bugs)
- **`build.sh`** — Script de build multi-plataforma (Linux, macOS, Windows)