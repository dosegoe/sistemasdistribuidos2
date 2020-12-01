package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"strconv"
	"reflect"
)
func errCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	/*
	file, err := os.OpenFile("./file.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)  
	errCheck(err)
	w := bufio.NewWriter(file)

	_, err1 := w.WriteString("old falcon\n")
	errCheck(err1)
	w.Flush()
	file.Close()
	
	/*
	f, err := os.Create("log.txt")

	errCheck(err)


	_, err2 := f.WriteString("old falcon\n")

	errCheck(err2)
	f.Close()

	f1, err1 := os.Create("log.txt")

	errCheck(err1)
	defer f.Close()

	
	fmt.Println("done")
	flag := 1
	for flag == 1 {
		fmt.Print("shupalo")
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your city: ")
		city, _ := reader.ReadString('\n')
		fmt.Print("You live in " + city)
		/*
		_, err2 := f1.WriteString(city)
		errCheck(err2)
		w := bufio.NewWriter(file)

		file, err := os.OpenFile("./file.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)  
		errCheck(err)
		w := bufio.NewWriter(file)

		_, err1 := w.WriteString(city)
		errCheck(err1)
		w.Flush()
		if strings.Compare(city, "0\n") == 0 {
			fmt.Print("entraste aca")
			flag = 0
		}
	}
	*/
	libro  := "Desobediencia_civil-Henry_David_Thoreau"
	//libro2 := "Dracula-Stoker_Bram"
	f, err5 := os.Open("file.txt")
    errCheck(err5)
    defer func() {
    	f.Close()
    }()
	s := bufio.NewScanner(f)
	//lee primero una vez para obtener el título y el número de partes
	for s.Scan(){
		
		errCheck(s.Err())
		muchotexto := s.Text()
		fmt.Println(reflect.TypeOf(muchotexto))
		separados := strings.Split(muchotexto, " ")
		fmt.Print(separados[0]+"\n")
		nombre := separados[0]
		fmt.Println("el nombre del libro es: "+nombre+"\n")
		partes, err6 := strconv.Atoi(separados[1])
		errCheck(err6)
		fmt.Println("la cantidad de partes son: ")
		fmt.Print(partes)
		fmt.Println("\n")
		if strings.Compare(nombre, libro) == 0{
			for i := 0; i < partes; i++ {
				s.Scan()
				fmt.Println(s.Text())
				muchotexto := s.Text()
				fmt.Println(reflect.TypeOf(muchotexto))
				separados := strings.Split(muchotexto, " ")
				fmt.Println(separados)
			}
			break
		}else{
			for i := 0; i < partes; i++ {
				s.Scan()
			}
		}
	}	
}