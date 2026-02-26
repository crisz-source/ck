package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scanCmd = &cobra.Command{
	Use:   "scan [imagem]",
	Short: "Scan de vulnerabilidades em imagem Docker",
	Example: `  ck scan nginx:latest
  ck scan nginx:latest -s CRITICAL,HIGH
  ck scan nginx:latest --no-secrets`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		image := args[0]
		scanImage(image)
	},
}

func init() {
	RootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringP("severity", "s", "", "Filtrar por severidade (ex: CRITICAL,HIGH)")
	scanCmd.Flags().Bool("no-secrets", false, "Desabilitar scan de secrets")

	viper.BindPFlag("scan.severity", scanCmd.Flags().Lookup("severity"))
	viper.BindPFlag("scan.no_secrets", scanCmd.Flags().Lookup("no-secrets"))
}

// Structs para parsear JSON do Trivy

type TrivyReport struct {
	Results []TrivyResult `json:"Results"`
}

type TrivyResult struct {
	Target          string        `json:"Target"`
	Class           string        `json:"Class"`
	Vulnerabilities []TrivyVuln   `json:"Vulnerabilities"`
	Secrets         []TrivySecret `json:"Secrets"`
}

type TrivyVuln struct {
	VulnerabilityID  string `json:"VulnerabilityID"`
	PkgName          string `json:"PkgName"`
	InstalledVersion string `json:"InstalledVersion"`
	FixedVersion     string `json:"FixedVersion"`
	Severity         string `json:"Severity"`
	Title            string `json:"Title"`
	PrimaryURL       string `json:"PrimaryURL"`
}

type TrivySecret struct {
	RuleID    string `json:"RuleID"`
	Category  string `json:"Category"`
	Severity  string `json:"Severity"`
	Title     string `json:"Title"`
	Match     string `json:"Match"`
	StartLine int    `json:"StartLine"`
	EndLine   int    `json:"EndLine"`
}

type secretWithTarget struct {
	TrivySecret
	Target string
}

func scanImage(image string) {
	_, err := exec.LookPath("trivy")
	if err != nil {
		fmt.Println("Erro: Trivy nao esta instalado")
		fmt.Println("")
		fmt.Println("Instale com:")
		fmt.Println("  Ubuntu: sudo apt-get install trivy")
		fmt.Println("  Mac: brew install trivy")
		fmt.Println("  Docs: https://aquasecurity.github.io/trivy")
		return
	}

	severity := viper.GetString("scan.severity")
	noSecrets := viper.GetBool("scan.no_secrets")

	scanners := "vuln,secret"
	if noSecrets {
		scanners = "vuln"
	}

	// Header
	fmt.Println(strings.Repeat("=", 71))
	fmt.Printf("  SCAN: %s\n", image)
	fmt.Println(strings.Repeat("=", 71))
	fmt.Println("")

	if severity != "" {
		fmt.Printf("Severidade: %s\n", severity)
	}
	fmt.Println("Escaneando... (pode demorar alguns segundos)")
	fmt.Println("")

	// Build trivy args
	args := []string{
		"image",
		"--format", "json",
		"--scanners", scanners,
		"--quiet",
	}
	if severity != "" {
		args = append(args, "--severity", severity)
	}
	args = append(args, image)

	out, err := exec.Command("trivy", args...).Output()
	if err != nil {
		if len(out) == 0 {
			fmt.Println("Erro ao executar trivy:", err)
			return
		}
	}

	var report TrivyReport
	if err := json.Unmarshal(out, &report); err != nil {
		fmt.Println("Erro ao parsear resultado:", err)
		return
	}

	// Coletar vulns e secrets
	vulnCounts := map[string]int{"CRITICAL": 0, "HIGH": 0, "MEDIUM": 0, "LOW": 0}
	secretCounts := map[string]int{"CRITICAL": 0, "HIGH": 0, "MEDIUM": 0, "LOW": 0}

	var allVulns []TrivyVuln
	var allSecrets []secretWithTarget

	for _, result := range report.Results {
		for _, vuln := range result.Vulnerabilities {
			vulnCounts[vuln.Severity]++
			allVulns = append(allVulns, vuln)
		}
		for _, secret := range result.Secrets {
			secretCounts[secret.Severity]++
			allSecrets = append(allSecrets, secretWithTarget{secret, result.Target})
		}
	}

	totalVulns := vulnCounts["CRITICAL"] + vulnCounts["HIGH"] + vulnCounts["MEDIUM"] + vulnCounts["LOW"]
	totalSecrets := secretCounts["CRITICAL"] + secretCounts["HIGH"] + secretCounts["MEDIUM"] + secretCounts["LOW"]

	if totalVulns == 0 && totalSecrets == 0 {
		fmt.Println("[OK] Nenhuma vulnerabilidade ou secret encontrado!")
		return
	}

	// RESUMO
	fmt.Println("RESUMO")
	boxSep := "+" + strings.Repeat("-", 37) + "+"

	fmt.Println(boxSep)
	fmt.Printf("| %-35s |\n", "VULNERABILIDADES")
	if totalVulns == 0 {
		fmt.Printf("| %-35s |\n", "  Nenhuma encontrada")
	} else {
		for _, sev := range []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"} {
			if vulnCounts[sev] > 0 {
				indicator := "[-]"
				if sev == "CRITICAL" || sev == "HIGH" {
					indicator = "[!]"
				}
				line := fmt.Sprintf("  %s %-8s: %d", indicator, sev, vulnCounts[sev])
				fmt.Printf("| %-35s |\n", line)
			}
		}
	}

	if !noSecrets {
		fmt.Println(boxSep)
		fmt.Printf("| %-35s |\n", "SECRETS EXPOSTOS")
		if totalSecrets == 0 {
			fmt.Printf("| %-35s |\n", "  Nenhum encontrado")
		} else {
			for _, sev := range []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"} {
				if secretCounts[sev] > 0 {
					indicator := "[-]"
					if sev == "CRITICAL" || sev == "HIGH" {
						indicator = "[!]"
					}
					line := fmt.Sprintf("  %s %-8s: %d", indicator, sev, secretCounts[sev])
					fmt.Printf("| %-35s |\n", line)
				}
			}
		}
	}
	fmt.Println(boxSep)
	fmt.Println("")

	// Tabela de VULNERABILIDADES (larguras dinamicas, sem cortar nada)
	if totalVulns > 0 {
		fmt.Println("VULNERABILIDADES")

		// Calcular largura de cada coluna baseado nos dados reais
		sevW := len("SEVER.")
		cveW := len("CVE")
		pkgW := len("PACOTE")
		verW := len("VERSAO")
		fixW := len("FIX")
		linkW := len("LINK")

		limit := len(allVulns)
		if limit > 20 {
			limit = 20
		}

		for i := 0; i < limit; i++ {
			vuln := allVulns[i]
			sevW = maxInt(sevW, len(vuln.Severity))
			cveW = maxInt(cveW, len(vuln.VulnerabilityID))
			pkgW = maxInt(pkgW, len(vuln.PkgName))
			verW = maxInt(verW, len(vuln.InstalledVersion))
			fix := vuln.FixedVersion
			if fix == "" {
				fix = "sem fix"
			}
			fixW = maxInt(fixW, len(fix))
			linkW = maxInt(linkW, len(vuln.PrimaryURL))
		}

		vulnSep := fmt.Sprintf("+%s+%s+%s+%s+%s+%s+",
			strings.Repeat("-", sevW+2),
			strings.Repeat("-", cveW+2),
			strings.Repeat("-", pkgW+2),
			strings.Repeat("-", verW+2),
			strings.Repeat("-", fixW+2),
			strings.Repeat("-", linkW+2))

		rowFmt := fmt.Sprintf("| %%-%ds | %%-%ds | %%-%ds | %%-%ds | %%-%ds | %%-%ds |\n",
			sevW, cveW, pkgW, verW, fixW, linkW)

		fmt.Println(vulnSep)
		fmt.Printf(rowFmt, "SEVER.", "CVE", "PACOTE", "VERSAO", "FIX", "LINK")
		fmt.Println(vulnSep)

		for i := 0; i < limit; i++ {
			vuln := allVulns[i]
			fix := vuln.FixedVersion
			if fix == "" {
				fix = "sem fix"
			}
			fmt.Printf(rowFmt,
				vuln.Severity,
				vuln.VulnerabilityID,
				vuln.PkgName,
				vuln.InstalledVersion,
				fix,
				vuln.PrimaryURL)
		}

		fmt.Println(vulnSep)
		if len(allVulns) > 20 {
			fmt.Printf("... e mais %d vulnerabilidades\n", len(allVulns)-20)
		}
		fmt.Println("")
	}

	// Tabela de SECRETS (larguras dinamicas)
	if totalSecrets > 0 && !noSecrets {
		fmt.Println("SECRETS EXPOSTOS")

		sevW := len("SEVER.")
		tipoW := len("TIPO")
		arqW := len("ARQUIVO")

		for _, s := range allSecrets {
			sevW = maxInt(sevW, len(s.Severity))
			tipoW = maxInt(tipoW, len(s.Title))
			arqW = maxInt(arqW, len(s.Target))
		}

		secretSep := fmt.Sprintf("+%s+%s+%s+",
			strings.Repeat("-", sevW+2),
			strings.Repeat("-", tipoW+2),
			strings.Repeat("-", arqW+2))

		rowFmt := fmt.Sprintf("| %%-%ds | %%-%ds | %%-%ds |\n",
			sevW, tipoW, arqW)

		fmt.Println(secretSep)
		fmt.Printf(rowFmt, "SEVER.", "TIPO", "ARQUIVO")
		fmt.Println(secretSep)

		for _, s := range allSecrets {
			fmt.Printf(rowFmt, s.Severity, s.Title, s.Target)
		}

		fmt.Println(secretSep)
		fmt.Println("")
	}

	// Alerta final
	if vulnCounts["CRITICAL"] > 0 || secretCounts["CRITICAL"] > 0 {
		parts := []string{}
		if vulnCounts["CRITICAL"] > 0 {
			parts = append(parts, fmt.Sprintf("%d vulnerabilidade(s) CRITICAL", vulnCounts["CRITICAL"]))
		}
		if secretCounts["CRITICAL"] > 0 {
			parts = append(parts, fmt.Sprintf("%d secret(s) CRITICAL", secretCounts["CRITICAL"]))
		}
		fmt.Printf("[ATENCAO] Encontrados %s!\n", strings.Join(parts, " + "))
		fmt.Println("          Recomenda-se atualizar os pacotes afetados.")
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
