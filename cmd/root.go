package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

// Variaveis globais para flags
var Namespace string
var TailLines string

// RootCmd - Comando raiz
var RootCmd = &cobra.Command{
    Use:   "ck",
    Short: "CLI para troubleshooting Kubernetes",
    Long:  "ck e uma ferramenta de linha de comando para facilitar troubleshooting no Kubernetes.",
}

// Execute executa o comando raiz
func Execute() {
    if err := RootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
