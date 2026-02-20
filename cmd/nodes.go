package cmd

import (
    "fmt"
    "os/exec"
    "strings"

    "github.com/spf13/cobra"
)

var nodesCmd = &cobra.Command{
    Use:   "nodes",
    Short: "Mostra status dos nodes com CPU e memoria",
    Example: `  ck nodes`,
    Run: func(cmd *cobra.Command, args []string) {
        showNodes()
    },
}

func init() {
    RootCmd.AddCommand(nodesCmd)
}

func showNodes() {
    out, err := exec.Command("kubectl", "get", "nodes", "-o", "wide", "--no-headers").Output()
    if err != nil {
        fmt.Println("Erro ao buscar nodes:", err)
        return
    }

    metricsOut, metricsErr := exec.Command("kubectl", "top", "nodes", "--no-headers").Output()
    
    metricsMap := make(map[string][]string)
    if metricsErr == nil {
        metricsLines := strings.Split(strings.TrimSpace(string(metricsOut)), "\n")
        for _, line := range metricsLines {
            fields := strings.Fields(line)
            if len(fields) >= 5 {
                metricsMap[fields[0]] = fields[1:]
            }
        }
    }

    fmt.Println("=== NODES ===")
    fmt.Println(strings.Repeat("-", 100))
    fmt.Printf("%-40s %-10s %-12s %-10s %-10s %-10s\n", 
        "NAME", "STATUS", "ROLES", "CPU", "CPU%", "MEMORY")
    fmt.Println(strings.Repeat("-", 100))

    lines := strings.Split(strings.TrimSpace(string(out)), "\n")
    readyCount := 0
    notReadyCount := 0

    for _, line := range lines {
        fields := strings.Fields(line)
        if len(fields) < 5 {
            continue
        }

        name := fields[0]
        status := fields[1]
        roles := fields[2]

        cpu := "-"
        cpuPercent := "-"
        mem := "-"

        if metrics, ok := metricsMap[name]; ok && len(metrics) >= 4 {
            cpu = metrics[0]
            cpuPercent = metrics[1]
            mem = metrics[2]
        }

        statusMark := status
        if status == "Ready" {
            readyCount++
        } else {
            statusMark = status + " [!]"
            notReadyCount++
        }

        fmt.Printf("%-40s %-10s %-12s %-10s %-10s %-10s\n",
            name, statusMark, roles, cpu, cpuPercent, mem)
    }

    fmt.Println(strings.Repeat("-", 100))
    fmt.Printf("Total: %d nodes | %d Ready | %d NotReady\n", len(lines), readyCount, notReadyCount)

    if notReadyCount > 0 {
        fmt.Println("\n[ATENCAO] Existem nodes com problema!")
    }
}
