package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"   
)

var logsCmd = &cobra.Command{
	Use:   "logs [pod]",
	Short: "Mostra logs de um pod",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		podName := args[0]
		showLogs(podName)
	},
}

func init() {
	RootCmd.AddCommand(logsCmd)

	
	// logsCmd.Flags().StringVarP(&Namespace, "namespace", "n", "", "Namespace do pod")

	
	// logsCmd.Flags().StringVarP(&TailLines, "tail", "t", "", "Numero de linhas")
	logsCmd.Flags().StringP("tail", "t", "", "Numero de linhas")
	viper.BindPFlag("tail", logsCmd.Flags().Lookup("tail"))
}

func showLogs(podName string) {
	ns := viper.GetString("namespace")
	tail := viper.GetString("tail")

	var args []string
	if ns == "" {
		args = []string{"logs", podName}
	} else {
		args = []string{"logs", podName, "-n", ns}
	}

	// tail agora pode vir da flag -t OU do ~/.ck.yaml OU do default "100"
	if tail != "" {
		args = append(args, "--tail", tail)
	}

	cmdExec := exec.Command("kubectl", args...)
	cmdExec.Stdout = os.Stdout
	cmdExec.Stderr = os.Stderr

	err := cmdExec.Run()
	if err != nil {
		fmt.Println("Erro ao executar kubectl logs:", err)
	}
}