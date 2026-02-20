package main

import (
    "errors"
    "fmt"
    "strconv"
)

// Funcao que converte string para int
func parsePort(s string) (int, error) {
    port, err := strconv.Atoi(s)
    if err != nil {
        return 0, errors.New("porta invalida: " + s)
    }
    
    if port < 1 || port > 65535 {
        return 0, errors.New("porta fora do range: " + s)
    }
    
    return port, nil
}

func main() {
    // Testes
    portas := []string{"8080", "443", "abc", "99999", "22"}
    
    for _, p := range portas {
        porta, err := parsePort(p)
        
        if err != nil {
            fmt.Printf("ERRO: %s\n", err)
            continue
        }
        
        fmt.Printf("Porta valida: %d\n", porta)
    }
}
