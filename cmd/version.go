package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Mostra a versao do ck",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("ck version 0.2.0")
    },
}

func init() {
    RootCmd.AddCommand(versionCmd)
}
