package cmd

import (
    "fmt"
    "os/exec"
    "sort"
    "strconv"
    "strings"

    "ck/types"

    "github.com/spf13/cobra"
)

var sortByMemory bool

var topCmd = &cobra.Command{
    Use:   "top",
    Short: "Mostra pods por consumo de CPU/memoria",
    Example: `  ck top              # Ordena por CPU
  ck top -m           # Ordena por memoria
  ck top -n kube-system -m`,
    Run: func(cmd *cobra.Command, args []string) {
        showTop()
    },
}

func init() {
    RootCmd.AddCommand(topCmd)
    topCmd.Flags().StringVarP(&Namespace, "namespace", "n", "", "Filtrar por namespace")
    topCmd.Flags().BoolVarP(&sortByMemory, "memory", "m", false, "Ordenar por memoria (padrao: CPU)")
}

func showTop() {
    var args []string
    if Namespace == "" {
        args = []string{"top", "pods", "-A", "--no-headers"}
    } else {
        args = []string{"top", "pods", "-n", Namespace, "--no-headers"}
    }

    out, err := exec.Command("kubectl", args...).Output()
    if err != nil {
        fmt.Println("Erro ao executar kubectl top:", err)
        fmt.Println("Dica: Metrics Server precisa estar instalado no cluster")
        return
    }

    lines := strings.Split(strings.TrimSpace(string(out)), "\n")
    var metrics []types.PodMetrics

    for _, line := range lines {
        fields := strings.Fields(line)
        if len(fields) < 3 {
            continue
        }

        var m types.PodMetrics
        if Namespace == "" && len(fields) >= 4 {
            m.Namespace = fields[0]
            m.Name = fields[1]
            m.CPU = fields[2]
            m.Memory = fields[3]
        } else if len(fields) >= 3 {
            m.Namespace = Namespace
            m.Name = fields[0]
            m.CPU = fields[1]
            m.Memory = fields[2]
        }

        cpuStr := strings.TrimSuffix(m.CPU, "m")
        m.CPUValue, _ = strconv.ParseInt(cpuStr, 10, 64)

        memStr := strings.TrimSuffix(m.Memory, "Mi")
        m.MemValue, _ = strconv.ParseInt(memStr, 10, 64)

        metrics = append(metrics, m)
    }

    // Ordena por memoria ou CPU
    if sortByMemory {
        sort.Slice(metrics, func(i, j int) bool {
            return metrics[i].MemValue > metrics[j].MemValue
        })
    } else {
        sort.Slice(metrics, func(i, j int) bool {
            return metrics[i].CPUValue > metrics[j].CPUValue
        })
    }

    // Titulo mostra qual ordenacao
    if sortByMemory {
        fmt.Println("=== TOP PODS (por MEMORIA) ===")
    } else {
        fmt.Println("=== TOP PODS (por CPU) ===")
    }
    
    fmt.Printf("%-20s %-40s %-10s %-10s\n", "NAMESPACE", "NAME", "CPU", "MEMORY")
    fmt.Println(strings.Repeat("-", 80))

    for _, m := range metrics {
        fmt.Printf("%-20s %-40s %-10s %-10s\n", m.Namespace, m.Name, m.CPU, m.Memory)
    }

    fmt.Println(strings.Repeat("-", 80))
    fmt.Printf("Total: %d pods\n", len(metrics))
}
