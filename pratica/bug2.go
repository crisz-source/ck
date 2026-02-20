package main

import "fmt"

func dobrar(x *int) {
    *x = *x * 2
}

func main() {
    numero := 10
    dobrar(&numero)
    fmt.Println("Dobro:", numero)
}
