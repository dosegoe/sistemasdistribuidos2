package main

import(
	"fmt"
	"time"
	"math/rand"
)



func main(){
	rand.Seed(time.Now().UnixNano())
	fmt.Println(rand.Intn(3)+1)
}