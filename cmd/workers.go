package cmd

import (
    "fmt"
    "os/exec"
    "strings"

    "ck/types"

    "github.com/spf13/cobra"
    "gopkg.in/yaml.v3"
)

var workersCmd = &cobra.Command{
    Use:   "workers",
    Short: "Mostra status dos workers do Supervisor",
    Example: `  ck workers -n php-worker           # Todos os pods do namespace
  ck workers php-worker-light-xyz -n php-worker  # Pod especifico`,
    Run: func(cmd *cobra.Command, args []string) {
        if len(args) > 0 {
            showWorkers(args[0])
        } else {
            showAllWorkers()
        }
    },
}

func init() {
    RootCmd.AddCommand(workersCmd)
    workersCmd.Flags().StringVarP(&Namespace, "namespace", "n", "", "Namespace dos pods")
}

type WorkerStatus struct {
    Name   string
    Status string
    Info   string
}

func showAllWorkers() {
    if Namespace == "" {
        fmt.Println("Erro: namespace e obrigatorio. Use -n <namespace>")
        return
    }

    args := []string{"get", "pods", "-n", Namespace, "-o", "yaml"}
    out, err := exec.Command("kubectl", args...).Output()
    if err != nil {
        fmt.Println("Erro ao buscar pods:", err)
        return
    }

    var podList types.PodList
    err = yaml.Unmarshal(out, &podList)
    if err != nil {
        fmt.Println("Erro ao parsear YAML:", err)
        return
    }

    if len(podList.Items) == 0 {
        fmt.Println("Nenhum pod encontrado no namespace:", Namespace)
        return
    }

    fmt.Printf("=== WORKERS STATUS - NAMESPACE: %s ===\n\n", Namespace)

    totalRunning := 0
    totalFatal := 0
    totalStopped := 0

    for _, pod := range podList.Items {
        if pod.Status.Phase != "Running" {
            fmt.Printf("POD: %s [%s - SKIP]\n\n", pod.Metadata.Name, pod.Status.Phase)
            continue
        }

        running, fatal, stopped := showWorkersForPod(pod.Metadata.Name)
        totalRunning += running
        totalFatal += fatal
        totalStopped += stopped
    }

    fmt.Println(strings.Repeat("=", 50))
    fmt.Printf("TOTAL GERAL: %d OK | %d FATAL | %d STOPPED\n", 
        totalRunning, totalFatal, totalStopped)

    if totalFatal > 0 {
        fmt.Println("\n ATENCAO: Workers em FATAL precisam de investigacao!")
    }
}

func showWorkersForPod(podName string) (int, int, int) {
    var args []string
    if Namespace == "" {
        args = []string{"exec", podName, "--", "supervisorctl", "status"}
    } else {
        args = []string{"exec", podName, "-n", Namespace, "--", "supervisorctl", "status"}
    }

    out, _ := exec.Command("kubectl", args...).CombinedOutput()
    outStr := string(out)
    
    // Verifica se tem output valido - procura por palavras-chave do supervisorctl
    if outStr == "" || 
       strings.HasPrefix(outStr, "error:") || 
       strings.Contains(outStr, "unable to upgrade") ||
       strings.Contains(outStr, "OCI runtime exec failed") ||
       strings.Contains(outStr, "command not found") {
        fmt.Printf("POD: %s [Erro: sem Supervisor]\n\n", podName)
        return 0, 0, 0
    }

    // Verifica se parece com output do supervisorctl (tem RUNNING, FATAL ou STOPPED)
    if !strings.Contains(outStr, "RUNNING") && 
       !strings.Contains(outStr, "FATAL") && 
       !strings.Contains(outStr, "STOPPED") {
        fmt.Printf("POD: %s [Erro: sem Supervisor]\n\n", podName)
        return 0, 0, 0
    }

    lines := strings.Split(strings.TrimSpace(outStr), "\n")
    
    var running []WorkerStatus
    var fatal []WorkerStatus
    var stopped []WorkerStatus

    for _, line := range lines {
        if line == "" {
            continue
        }

        fields := strings.Fields(line)
        if len(fields) < 2 {
            continue
        }

        worker := WorkerStatus{
            Name:   fields[0],
            Status: fields[1],
        }

        switch worker.Status {
        case "RUNNING":
            running = append(running, worker)
        case "FATAL":
            fatal = append(fatal, worker)
        case "STOPPED":
            stopped = append(stopped, worker)
        }
    }

    fmt.Printf("POD: %s\n", podName)
    
    if len(fatal) > 0 {
        fmt.Printf("  ✗ FATAL (%d): ", len(fatal))
        names := []string{}
        for _, w := range fatal {
            parts := strings.Split(w.Name, ":")
            names = append(names, parts[0])
        }
        if len(names) > 5 {
            fmt.Printf("%s ... (+%d mais)\n", strings.Join(names[:5], ", "), len(names)-5)
        } else {
            fmt.Printf("%s\n", strings.Join(names, ", "))
        }
    }

    if len(stopped) > 0 {
        fmt.Printf(" STOPPED (%d)\n", len(stopped))
    }

    fmt.Printf("  ✓ RUNNING (%d)\n", len(running))
    fmt.Printf("  Status: %d OK | %d FATAL | %d STOPPED\n\n", 
        len(running), len(fatal), len(stopped))

    return len(running), len(fatal), len(stopped)
}

func showWorkers(podName string) {
    var args []string
    if Namespace == "" {
        args = []string{"exec", podName, "--", "supervisorctl", "status"}
    } else {
        args = []string{"exec", podName, "-n", Namespace, "--", "supervisorctl", "status"}
    }

    out, _ := exec.Command("kubectl", args...).CombinedOutput()
    outStr := string(out)
    
    if outStr == "" || 
       strings.HasPrefix(outStr, "error:") ||
       strings.Contains(outStr, "command not found") {
        fmt.Println("Erro ao executar supervisorctl")
        fmt.Println("Dica: Verifique se o pod tem Supervisor instalado")
        return
    }

    lines := strings.Split(strings.TrimSpace(outStr), "\n")
    
    var running []WorkerStatus
    var fatal []WorkerStatus
    var stopped []WorkerStatus
    var other []WorkerStatus

    for _, line := range lines {
        if line == "" {
            continue
        }

        fields := strings.Fields(line)
        if len(fields) < 2 {
            continue
        }

        worker := WorkerStatus{
            Name:   fields[0],
            Status: fields[1],
        }
        if len(fields) > 2 {
            worker.Info = strings.Join(fields[2:], " ")
        }

        switch worker.Status {
        case "RUNNING":
            running = append(running, worker)
        case "FATAL":
            fatal = append(fatal, worker)
        case "STOPPED":
            stopped = append(stopped, worker)
        default:
            other = append(other, worker)
        }
    }

    fmt.Printf("=== WORKERS STATUS ===\n")
    fmt.Printf("POD: %s\n", podName)
    if Namespace != "" {
        fmt.Printf("NAMESPACE: %s\n", Namespace)
    }
    fmt.Println()

    if len(fatal) > 0 {
        fmt.Printf("FATAL (%d):\n", len(fatal))
        for _, w := range fatal {
            parts := strings.Split(w.Name, ":")
            fmt.Printf("  ✗ %s\n", parts[0])
        }
        fmt.Println()
    }

    if len(stopped) > 0 {
        fmt.Printf("STOPPED (%d):\n", len(stopped))
        for _, w := range stopped {
            parts := strings.Split(w.Name, ":")
            fmt.Printf("  ○ %s\n", parts[0])
        }
        fmt.Println()
    }

    if len(running) > 0 {
        fmt.Printf("RUNNING (%d):\n", len(running))
        for _, w := range running {
            parts := strings.Split(w.Name, ":")
            fmt.Printf("  ✓ %s\n", parts[0])
        }
        fmt.Println()
    }

    if len(other) > 0 {
        fmt.Printf("OTHER (%d):\n", len(other))
        for _, w := range other {
            fmt.Printf("  ? %s (%s)\n", w.Name, w.Status)
        }
        fmt.Println()
    }

    fmt.Println(strings.Repeat("-", 40))
    total := len(running) + len(fatal) + len(stopped) + len(other)
    fmt.Printf("RESUMO: %d OK | %d FATAL | %d STOPPED | %d TOTAL\n", 
        len(running), len(fatal), len(stopped), total)

    if len(fatal) > 0 {
        fmt.Println()
        fmt.Println(" ATENCAO: Workers em FATAL precisam de investigacao!")
        fmt.Println("   Use: ck exec <pod> -n <ns> -- cat /var/log/supervisor/<worker>-stderr.log")
    }
}
