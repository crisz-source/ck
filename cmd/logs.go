package cmd

import (
    "fmt"
    "os"
    "os/exec"

    "github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
    Use:   "logs [pod ( apenas o nome do pod ) ]",
    Short: "Mostra logs de um pod",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        podName := args[0]
        showLogs(podName)
    },
}

func init() {
    RootCmd.AddCommand(logsCmd)
    logsCmd.Flags().StringVarP(&Namespace, "namespace", "n", "", "Namespace do pod")
    logsCmd.Flags().StringVarP(&TailLines, "tail", "t", "", "Numero de linhas")
}

func showLogs(podName string) {
    var args []string

    if Namespace == "" {
        args = []string{"logs", podName}
    } else {
        args = []string{"logs", podName, "-n", Namespace}
    }

    if TailLines != "" {
        args = append(args, "--tail", TailLines)
    }

    cmdExec := exec.Command("kubectl", args...)
    cmdExec.Stdout = os.Stdout
    cmdExec.Stderr = os.Stderr

    err := cmdExec.Run()
    if err != nil {
        fmt.Println("Erro ao executar kubectl logs:", err)
    }
}
