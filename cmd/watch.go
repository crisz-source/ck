package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ck/k8s"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Monitora pods em tempo real com alertas por email",
	Long: `Monitora pods em tempo real usando Kubernetes Informers.
Detecta restarts, CrashLoopBackOff, OOMKilled e outros problemas.
Envia alertas por email via Azure Communication Services.

Configure no ~/.ck.yaml:
  watch:
    restart_threshold: 3
  notify:
    email:
      connection_string: "endpoint=https://...;accesskey=..."
      from: "DoNotReply@pgmbh.org"
      to: "seu-email@destino.com"`,
	Example: `  ck watch                    # Monitora namespace do config
  ck watch -n php-worker      # Monitora namespace específico
  ck watch -n ""              # Monitora TODOS os namespaces`,
	Run: func(cmd *cobra.Command, args []string) {
		runWatch()
	},
}

func init() {
	RootCmd.AddCommand(watchCmd)
}

func runWatch() {
	ns := viper.GetString("namespace")

	// Conecta no cluster
	clientset, err := k8s.GetClient()
	if err != nil {
		fmt.Println("Erro ao conectar no cluster:", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Roda o watcher em goroutine
	go k8s.WatchPods(ctx, clientset, ns)

	// Espera sinal de parada (Ctrl+C)
	sig := <-sigCh
	fmt.Printf("\n🛑 Recebido sinal %v. Parando monitoramento...\n", sig)
	cancel() // Cancela o contexto → para o Informer
}
