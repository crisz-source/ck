package main

import "fmt"

func main() {
    a := 10
    b := 10
    
    fmt.Println("=== Enderecos diferentes, mesmo valor ===")
    // aqui vai mostrar apenas os enedereço das variavesi a e b
    fmt.Println("a =", a, "endereco:", &a) 
    fmt.Println("b =", b, "endereco:", &b)
    
    // aqui vai ser um ponteiro de &a, que vai também vai mostrar o endereço no primeiro print, e o segundo print vai mostrar o valor que está nessa gaveta que é 10
    fmt.Println("\n=== Ponteiro ===")
    ptr := &a
    fmt.Println("ptr aponta para:", ptr)
    fmt.Println("valor em ptr:", *ptr)
    
    // aqui vai modificar o ponteiro para 100, vai mudar o valor e o valor de b vai continuar
    fmt.Println("\n=== Modificando via ponteiro ===")
    *ptr = 100
    fmt.Println("a agora vale:", a)
    fmt.Println("b continua:", b)
}
