package main

import "fmt"

func main() {
    nums := []int{1, 2, 3, 4, 5}
    
    var resultado []*int
    
    for _, n := range nums {
        num := n
        resultado = append(resultado, &num)
    }
    
    for _, ptr := range resultado {
        fmt.Println(*ptr)
    }
}
