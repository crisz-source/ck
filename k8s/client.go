package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClient cria e retorna um clientset autenticado com o cluster K8s.
// Tenta 2 formas de autenticação:
//   1. In-cluster (se o ck estiver rodando DENTRO de um pod)
//   2. Kubeconfig (arquivo ~/.kube/config — o caso mais comum)
func GetClient() (*kubernetes.Clientset, error) {
	// Tenta in-cluster primeiro (pra quando ck rodar dentro do cluster)
	config, err := rest.InClusterConfig()
	if err != nil {
		// Não está dentro do cluster → usa kubeconfig
		kubeconfig := getKubeconfigPath()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("erro ao carregar kubeconfig: %w", err)
		}
	}

	// Cria o clientset (a "conexão" com o cluster)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente K8s: %w", err)
	}

	return clientset, nil
}

// getKubeconfigPath retorna o caminho do kubeconfig.
// Prioridade: KUBECONFIG env > ~/.kube/config
func getKubeconfigPath() string {
	// 1. Variável de ambiente KUBECONFIG (padrão do kubectl)
	if kc := os.Getenv("KUBECONFIG"); kc != "" {
		return kc
	}

	// 2. Caminho padrão ~/.kube/config
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".kube", "config")
}