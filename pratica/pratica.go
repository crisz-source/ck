package main

import "fmt"

func main() {
    // Variaveis basicas
    nome := "Cristhian"
    idade := 25
    
    fmt.Println("Nome:", nome)
    fmt.Println("Idade:", idade)
    
    // Ponteiros
    ptr := &idade
    fmt.Println("Endereco de idade:", ptr)
    fmt.Println("Valor no endereco:", *ptr)
    
    // Modificando via ponteiro
    *ptr = 30
    fmt.Println("Nova idade:", idade)
}
