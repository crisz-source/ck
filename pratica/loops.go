package main

import (
    "fmt"
    "strings"
)

func main() {
    // Slice de pods com problemas
    pods := []string{"api-crash", "worker-ok", "cache-error", "db-ok"}
    
    fmt.Println("=== Todos os pods ===")
    for i, pod := range pods {
        fmt.Printf("%d: %s\n", i, pod)
    }
    
    fmt.Println("\n=== Pods com problema ===")
    for _, pod := range pods {
        if strings.Contains(pod, "crash") || strings.Contains(pod, "error") {
            fmt.Println("PROBLEMA:", pod)
        }
    }
}

// Funcao auxiliar para verificar se string contem substring
func contains(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}
