package cmd

import (
    "fmt"
    "os/exec"
    "strings"

    "github.com/spf13/cobra"
)

var pvcCmd = &cobra.Command{
    Use:   "pvc",
    Short: "Lista PersistentVolumeClaims com tamanho e status",
    Example: `  ck pvc
  ck pvc -n default`,
    Run: func(cmd *cobra.Command, args []string) {
        showPVC()
    },
}

func init() {
    RootCmd.AddCommand(pvcCmd)
    pvcCmd.Flags().StringVarP(&Namespace, "namespace", "n", "", "Filtrar por namespace")
}

func showPVC() {
    var args []string
    if Namespace == "" {
        args = []string{"get", "pvc", "-A", "--no-headers"}
    } else {
        args = []string{"get", "pvc", "-n", Namespace, "--no-headers"}
    }

    out, err := exec.Command("kubectl", args...).Output()
    if err != nil {
        fmt.Println("Erro ao buscar PVCs:", err)
        return
    }

    outStr := strings.TrimSpace(string(out))
    if outStr == "" {
        fmt.Println("Nenhum PVC encontrado")
        return
    }

    fmt.Println("=== PERSISTENT VOLUME CLAIMS ===")
    fmt.Println(strings.Repeat("-", 95))
    
    if Namespace == "" {
        fmt.Printf("%-20s %-30s %-10s %-10s %-15s\n",
            "NAMESPACE", "NAME", "STATUS", "SIZE", "STORAGECLASS")
    } else {
        fmt.Printf("%-35s %-10s %-10s %-15s\n",
            "NAME", "STATUS", "SIZE", "STORAGECLASS")
    }
    fmt.Println(strings.Repeat("-", 95))

    lines := strings.Split(outStr, "\n")
    
    boundCount := 0
    pendingCount := 0

    for _, line := range lines {
        fields := strings.Fields(line)
        
        var namespace, name, status, capacity, storageClass string

        if Namespace == "" && len(fields) >= 6 {
            namespace = fields[0]
            name = fields[1]
            status = fields[2]
            capacity = fields[4]
            storageClass = fields[5]
        } else if len(fields) >= 5 {
            name = fields[0]
            status = fields[1]
            capacity = fields[3]
            storageClass = fields[4]
        } else {
            continue
        }

        if status == "Bound" {
            boundCount++
        } else {
            pendingCount++
        }

        statusMark := status
        if status != "Bound" {
            statusMark = status + " [!]"
        }

        if len(name) > 30 {
            name = name[:27] + "..."
        }

        if Namespace == "" {
            fmt.Printf("%-20s %-30s %-10s %-10s %-15s\n",
                namespace, name, statusMark, capacity, storageClass)
        } else {
            fmt.Printf("%-35s %-10s %-10s %-15s\n",
                name, statusMark, capacity, storageClass)
        }
    }

    fmt.Println(strings.Repeat("-", 95))
    fmt.Printf("Total: %d PVCs | %d Bound | %d Pending\n", len(lines), boundCount, pendingCount)

    if pendingCount > 0 {
        fmt.Println("\n[ATENCAO] Existem PVCs pendentes!")
    }
}
