package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Mostra a configuração ativa do ck",
	Example: `  ck config        # Mostra toda a config ativa
  ck config path   # Mostra o caminho do arquivo de config`,
	Run: func(cmd *cobra.Command, args []string) {
		showConfig()
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Mostra o caminho do arquivo de configuração",
	Run: func(cmd *cobra.Command, args []string) {
		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			fmt.Println("Nenhum arquivo de configuração encontrado")
			fmt.Println("Crie um em: ~/.ck.yaml")
		} else {
			fmt.Println(configFile)
		}
	},
}

func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configPathCmd)
}

func showConfig() {
	configFile := viper.ConfigFileUsed()

	fmt.Println("=== CK CONFIG ===")
	fmt.Println()

	if configFile != "" {
		fmt.Printf("Arquivo:    %s\n", configFile)
	} else {
		fmt.Println("Arquivo:    (nenhum encontrado)")
	}

	fmt.Println()
	fmt.Println("Valores ativos:")
	fmt.Println("  (prioridade: flag > env > arquivo > default)")
	fmt.Println()

	// Mostra as configs principais
	ns := viper.GetString("namespace")
	if ns == "" {
		ns = "(vazio → usa -A todos namespaces)"
	}
	fmt.Printf("  namespace:       %s\n", ns)
	fmt.Printf("  tail:            %s\n", viper.GetString("tail"))
	fmt.Printf("  scan.severity:   %s\n", viper.GetString("scan.severity"))

	// Mostra todas as chaves que o Viper conhece
	allKeys := viper.AllKeys()
	if len(allKeys) > 3 {
		fmt.Println()
		fmt.Println("Todas as chaves:")
		for _, key := range allKeys {
			fmt.Printf("  %-20s %v\n", key+":", viper.Get(key))
		}
	}
}