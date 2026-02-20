package cmd

import (
    "fmt"
    "os/exec"
    "strings"

    "github.com/spf13/cobra"
)

var ingressCmd = &cobra.Command{
    Use:   "ingress",
    Short: "Lista Ingresses com URLs",
    Example: `  ck ingress
  ck ingress -n default`,
    Run: func(cmd *cobra.Command, args []string) {
        showIngress()
    },
}

func init() {
    RootCmd.AddCommand(ingressCmd)
    ingressCmd.Flags().StringVarP(&Namespace, "namespace", "n", "", "Filtrar por namespace")
}

func showIngress() {
    var args []string
    if Namespace == "" {
        args = []string{"get", "ingress", "-A", "--no-headers"}
    } else {
        args = []string{"get", "ingress", "-n", Namespace, "--no-headers"}
    }

    out, err := exec.Command("kubectl", args...).Output()
    if err != nil {
        fmt.Println("Erro ao buscar Ingresses:", err)
        return
    }

    outStr := strings.TrimSpace(string(out))
    if outStr == "" {
        fmt.Println("Nenhum Ingress encontrado")
        return
    }

    fmt.Println("=== INGRESSES ===")
    fmt.Println(strings.Repeat("-", 100))
    
    if Namespace == "" {
        fmt.Printf("%-15s %-25s %-40s %s\n",
            "NAMESPACE", "NAME", "HOSTS", "ADDRESS")
    } else {
        fmt.Printf("%-30s %-40s %s\n",
            "NAME", "HOSTS", "ADDRESS")
    }
    fmt.Println(strings.Repeat("-", 100))

    lines := strings.Split(outStr, "\n")
    var allHosts []string

    for _, line := range lines {
        fields := strings.Fields(line)
        
        var namespace, name, hosts, address string

        if Namespace == "" && len(fields) >= 4 {
            namespace = fields[0]
            name = fields[1]
            hosts = fields[3]
            if len(fields) >= 5 {
                address = fields[4]
            }
        } else if len(fields) >= 3 {
            name = fields[0]
            hosts = fields[2]
            if len(fields) >= 4 {
                address = fields[3]
            }
        } else {
            continue
        }

        if hosts != "" && hosts != "*" {
            hostList := strings.Split(hosts, ",")
            allHosts = append(allHosts, hostList...)
        }

        if len(name) > 25 {
            name = name[:22] + "..."
        }
        if len(hosts) > 40 {
            hosts = hosts[:37] + "..."
        }

        if Namespace == "" {
            fmt.Printf("%-15s %-25s %-40s %s\n",
                namespace, name, hosts, address)
        } else {
            fmt.Printf("%-30s %-40s %s\n",
                name, hosts, address)
        }
    }

    fmt.Println(strings.Repeat("-", 100))
    fmt.Printf("Total: %d ingresses\n", len(lines))

    if len(allHosts) > 0 {
        fmt.Println("\nURLs:")
        for _, host := range allHosts {
            if host != "" && host != "*" {
                fmt.Printf("  https://%s\n", host)
            }
        }
    }
}
