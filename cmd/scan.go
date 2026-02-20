package cmd

import (
    "encoding/json"
    "fmt"
    "os/exec"
    "strings"

    "github.com/spf13/cobra"
)

var severityFilter string

var scanCmd = &cobra.Command{
    Use:   "scan [imagem]",
    Short: "Scan de vulnerabilidades em imagem Docker",
    Example: `  ck scan nginx:latest
  ck scan php-worker:latest -s CRITICAL
  ck scan myregistry/app:v1.0 -s HIGH,CRITICAL`,
    Args: cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        image := args[0]
        scanImage(image)
    },
}

func init() {
    RootCmd.AddCommand(scanCmd)
    scanCmd.Flags().StringVarP(&severityFilter, "severity", "s", "HIGH,CRITICAL", "Filtrar por severidade (CRITICAL,HIGH,MEDIUM,LOW)")
}

type TrivyReport struct {
    Results []TrivyResult `json:"Results"`
}

type TrivyResult struct {
    Target          string        `json:"Target"`
    Vulnerabilities []TrivyVuln   `json:"Vulnerabilities"`
}

type TrivyVuln struct {
    VulnerabilityID  string `json:"VulnerabilityID"`
    PkgName          string `json:"PkgName"`
    InstalledVersion string `json:"InstalledVersion"`
    FixedVersion     string `json:"FixedVersion"`
    Severity         string `json:"Severity"`
    Title            string `json:"Title"`
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

    fmt.Printf("=== SCAN: %s ===\n", image)
    fmt.Printf("Severidade: %s\n", severityFilter)
    fmt.Println("Escaneando... (pode demorar alguns segundos)")
    fmt.Println("")

    args := []string{
        "image",
        "--format", "json",
        "--severity", severityFilter,
        "--quiet",
        image,
    }

    out, err := exec.Command("trivy", args...).Output()
    if err != nil {
        if len(out) == 0 {
            fmt.Println("Erro ao executar trivy:", err)
            return
        }
    }

    var report TrivyReport
    err = json.Unmarshal(out, &report)
    if err != nil {
        fmt.Println("Erro ao parsear resultado:", err)
        return
    }

    counts := map[string]int{
        "CRITICAL": 0,
        "HIGH":     0,
        "MEDIUM":   0,
        "LOW":      0,
    }

    var allVulns []TrivyVuln

    for _, result := range report.Results {
        for _, vuln := range result.Vulnerabilities {
            counts[vuln.Severity]++
            allVulns = append(allVulns, vuln)
        }
    }

    total := counts["CRITICAL"] + counts["HIGH"] + counts["MEDIUM"] + counts["LOW"]

    if total == 0 {
        fmt.Println("[OK] Nenhuma vulnerabilidade encontrada!")
        return
    }

    fmt.Println("RESUMO:")
    if counts["CRITICAL"] > 0 {
        fmt.Printf("  [!] CRITICAL: %d\n", counts["CRITICAL"])
    }
    if counts["HIGH"] > 0 {
        fmt.Printf("  [!] HIGH:     %d\n", counts["HIGH"])
    }
    if counts["MEDIUM"] > 0 {
        fmt.Printf("  [-] MEDIUM:   %d\n", counts["MEDIUM"])
    }
    if counts["LOW"] > 0 {
        fmt.Printf("  [-] LOW:      %d\n", counts["LOW"])
    }
    fmt.Printf("      TOTAL:    %d\n", total)
    fmt.Println("")

    fmt.Println("VULNERABILIDADES:")
    fmt.Println(strings.Repeat("-", 95))
    fmt.Printf("%-10s %-18s %-25s %-15s %s\n", "SEVERIDADE", "CVE", "PACOTE", "VERSAO", "FIX")
    fmt.Println(strings.Repeat("-", 95))

    shown := 0
    for _, vuln := range allVulns {
        if shown >= 20 {
            fmt.Printf("\n... e mais %d vulnerabilidades\n", len(allVulns)-20)
            break
        }

        pkgName := vuln.PkgName
        if len(pkgName) > 25 {
            pkgName = pkgName[:22] + "..."
        }

        fixVersion := vuln.FixedVersion
        if fixVersion == "" {
            fixVersion = "sem fix"
        }
        if len(fixVersion) > 15 {
            fixVersion = fixVersion[:12] + "..."
        }

        installedVersion := vuln.InstalledVersion
        if len(installedVersion) > 15 {
            installedVersion = installedVersion[:12] + "..."
        }

        fmt.Printf("%-10s %-18s %-25s %-15s %s\n", 
            vuln.Severity, vuln.VulnerabilityID, pkgName, installedVersion, fixVersion)
        shown++
    }

    fmt.Println(strings.Repeat("-", 95))

    if counts["CRITICAL"] > 0 {
        fmt.Println("")
        fmt.Println("[ATENCAO] Vulnerabilidades CRITICAL encontradas!")
        fmt.Println("          Recomenda-se atualizar os pacotes afetados.")
    }
}
