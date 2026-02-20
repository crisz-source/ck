package main

import "fmt"

func main() {
    // Map de pods: nome -> status
    pods := map[string]string{
        "nginx-7d8f9b-x2k4p":  "Running",
        "api-5c8d7b-m3n2":     "CrashLoopBackOff",
        "worker-6f9a8c-j4k5":  "Running",
        "cache-8b7c6d-p5q6":   "ImagePullBackOff",
        "db-9a8b7c-r7s8":      "Running",
    }
    
    fmt.Println("=== Todos os pods ===")
    for nome, status := range pods {
        fmt.Printf("%s: %s\n", nome, status)
    }
    
    fmt.Println("\n=== Pods com problema ===")
    for nome, status := range pods {
        if status != "Running" {
            fmt.Printf("PROBLEMA: %s (%s)\n", nome, status)
        }
    }
    
    fmt.Println("\n=== Contagem por status ===")
    contagem := make(map[string]int)
    
    for _, status := range pods {
        contagem[status]++
    }
    
    for status, quantidade := range contagem {
        fmt.Printf("%s: %d\n", status, quantidade)
    }
}
