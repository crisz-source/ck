package cmd

import (
    "fmt"
    "os"
    "os/exec"

    "github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
    Use:   "exec [pod ( apenas o nome do pod )] -- [comando]",
    Short: "Executa comando dentro do pod",
    Example: `  ck exec meu-pod -- bash
  ck exec meu-pod -n default -- sh
  ck exec meu-pod -- ls -la`,
    Args: cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        podName := args[0]

        shellCmd := []string{"sh"}
        for i, arg := range args {
            if arg == "--" && i+1 < len(args) {
                shellCmd = args[i+1:]
                break
            }
        }

        execPod(podName, shellCmd)
    },
}

func init() {
    RootCmd.AddCommand(execCmd)
    execCmd.Flags().StringVarP(&Namespace, "namespace", "n", "", "Namespace do pod")
}

func execPod(podName string, shellCmd []string) {
    var args []string

    if Namespace == "" {
        args = []string{"exec", "-it", podName, "--"}
    } else {
        args = []string{"exec", "-it", podName, "-n", Namespace, "--"}
    }

    args = append(args, shellCmd...)

    cmdExec := exec.Command("kubectl", args...)
    cmdExec.Stdin = os.Stdin
    cmdExec.Stdout = os.Stdout
    cmdExec.Stderr = os.Stderr

    err := cmdExec.Run()
    if err != nil {
        fmt.Println("Erro ao executar comando no pod:", err)
    }
}
