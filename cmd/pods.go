package cmd

import (
	"context"
	"fmt"

	"ck/k8s"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var podsCmd = &cobra.Command{
	Use:   "pods",
	Short: "Lista pods com problema",
	Run: func(cmd *cobra.Command, args []string) {
		listProblematicPods()
	},
}

func init() {
	RootCmd.AddCommand(podsCmd)
}

func listProblematicPods() {
	ns := viper.GetString("namespace")

	// 1. Conecta no cluster
	clientset, err := k8s.GetClient()
	if err != nil {
		fmt.Println("Erro ao conectar no cluster:", err)
		return
	}

	ctx := context.Background()

	// 2. Lista pods (ns="" = todos os namespaces, igual -A)
	podList, err := clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Println("Erro ao listar pods:", err)
		return
	}

	// 3. Filtra e mostra pods com problema
	fmt.Println("NAMESPACE\t\tNAME\t\t\t\t\tSTATUS")
	fmt.Println("---------\t\t----\t\t\t\t\t------")

	problemCount := 0
	for _, pod := range podList.Items {
		status := string(pod.Status.Phase)
		hasProblem := false

		// Checa fase do pod
		if pod.Status.Phase != corev1.PodRunning &&
			pod.Status.Phase != corev1.PodSucceeded {
			hasProblem = true
		}

		// Checa containers
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
				status = cs.State.Waiting.Reason
				hasProblem = true
			}
			if cs.RestartCount > 5 {
				status = fmt.Sprintf("%s (restarts: %d)", status, cs.RestartCount)
				hasProblem = true
			}
		}

		if hasProblem {
			fmt.Printf("%s\t\t%s\t\t%s\n",
				pod.Namespace, pod.Name, status)
			problemCount++
		}
	}

	fmt.Println("---------")
	fmt.Printf("Total: %d pods com problema\n", problemCount)
}
