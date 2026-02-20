package main

import "fmt"

func main() {
    pods := map[string]string{
        "nginx": "Running",
        "api":   "",
    }

    status, existe := pods["nginx"]
    
    if !existe {
        fmt.Println("Pod nao encontrado")
    } else {
        fmt.Println("Status:", status)
    }



    // Verifica pod que existe mas status vazio
    status2, existe2 := pods["api"]
    if !existe2 {
        fmt.Println("Pod api nao encontrado")
    } else {
        fmt.Println("Status api:", status2)  // Mostra vazio, mas pod existe
    }

}
