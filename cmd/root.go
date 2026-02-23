package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cfgFile guarda o caminho se o usuário passar --config
var cfgFile string


var RootCmd = &cobra.Command{
	Use:   "ck",
	Short: "CLI para troubleshooting Kubernetes",
	Long: `ck (Cristhian + Kubernetes)
Ferramenta de linha de comando para facilitar troubleshooting no Kubernetes.

Configuração: ~/.ck.yaml | Variáveis de ambiente: CK_NAMESPACE, CK_TAIL`,
}

// Execute executa o comando raiz
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Registra initConfig pra rodar ANTES de qualquer comando
	// (depois que o Cobra já parseou as flags)
	cobra.OnInitialize(initConfig)

	// --config: aponta pra um arquivo de config específico
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"arquivo de configuração (padrão: $HOME/.ck.yaml)")

	// --namespace / -n: AGORA É PERSISTENT (disponível em TODOS os subcomandos)
	RootCmd.PersistentFlags().StringP("namespace", "n", "",
		"namespace do Kubernetes")

	// Conecta a flag --namespace com o Viper
	// Isso faz viper.GetString("namespace") retornar o valor da flag
	// SE ela foi passada, senão cai pro env → arquivo → default
	viper.BindPFlag("namespace", RootCmd.PersistentFlags().Lookup("namespace"))
}

// initConfig lê configurações do arquivo, env vars e defaults
func initConfig() {
	if cfgFile != "" {
		// Usuário passou --config /path/to/config.yaml
		viper.SetConfigFile(cfgFile)
	} else {
		// Procura ~/.ck.yaml
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erro ao encontrar home:", err)
			os.Exit(1)
		}

		viper.SetConfigName(".ck")    // nome do arquivo (sem extensão)
		viper.SetConfigType("yaml")   // formato
		viper.AddConfigPath(home)     // procura em ~/
		viper.AddConfigPath(".")      // procura no diretório atual
	}

	// Variáveis de ambiente: CK_NAMESPACE, CK_TAIL, etc.
	viper.SetEnvPrefix("CK")
	viper.AutomaticEnv()

	// Defaults (menor prioridade)
	// namespace: SEM default → se não tiver config, fica "" → mantém o -A
	viper.SetDefault("tail", "100")
	viper.SetDefault("scan.severity", "CRITICAL,HIGH")

	// Lê o arquivo de config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Arquivo existe mas tem erro de sintaxe
			fmt.Fprintf(os.Stderr, "Erro no arquivo de config: %v\n", err)
		}
		// Se não encontrou arquivo, silêncio — usa defaults e flags
	}
}