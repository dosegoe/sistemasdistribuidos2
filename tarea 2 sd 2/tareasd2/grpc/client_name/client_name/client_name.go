package client_name

import(
	//"io"
	"bufio"
	"strings"
	"fmt"
	//"math/rand"
	//"time"
	//"io/ioutil"
	"strconv"
	"os"
	//"context"
	"log"
	"reflect"
)

type Server struct{
	Messages *int
}

//revisa errores
func errCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//una vez tenga almacenado ordenes en el log, namenode envía al cliente info del orden
func (server *Server) ChunksOrder(req *OrderReq, stream ClientName_ChunksOrderServer) error{
	
	fileName:=req.GetFilename()
	fmt.Println("Buscamos: "+fileName+"\n")
	//tengo que abrir el log y buscar el maldito libro
	//como el título está seguido de la cantidad de partes puedo abrir un for que 
	//itere la cantidad de veces necesarias para poder obtener todas las partes y
	//por cada iteración enviar un OrderRes de respuesta.
	f, err5 := os.Open("namenode/log.txt")
    errCheck(err5)
    defer func() {
    	f.Close()
    }()
	s := bufio.NewScanner(f)
	//lee primero una vez para obtener el título y el número de partes
	for s.Scan(){
		
		errCheck(s.Err())
		muchotexto := s.Text()
		separados := strings.Split(muchotexto, " ")
		nombre := separados[0]
		partes, err6 := strconv.Atoi(separados[1])
		errCheck(err6)
		fmt.Println("el nombre del libro es: "+nombre+"\n")
		fmt.Println("la cantidad de partes son: ")
		fmt.Print(partes)
		fmt.Println("\n")

		
		if strings.Compare(nombre, fileName) == 0{
			for i := 0; i < partes; i++ {
				var n int64
				s.Scan()
				fmt.Println(s.Text())
				muchotexto := s.Text()
				fmt.Println(reflect.TypeOf(muchotexto))
				separados := strings.Split(muchotexto, " ")
				fmt.Println(separados)
				chunk := strings.Split(separados[0], "_")
				var chuId int64
				chuId, err1 := strconv.ParseInt(chunk[2], 10, 64)
				//n, err1 := strconv.ParseInt(separados[1], 10, 64)
				errCheck(err1)

				fileNodos, err := os.Open("namenode/Dnodes.txt")
				errCheck(err)
				scanNodos := bufio.NewScanner(fileNodos)
				scanNodos.Scan()
				linea_idip := scanNodos.Text()
				id_ip := strings.Split(linea_idip, " ")
				if id_ip[1] == separados[1]{
					n , err = strconv.ParseInt(id_ip[0],10,64)
					errCheck(err)
				}
				if err := stream.Send(&OrderRes{ChunkId: chuId, NodeId: n}); err != nil {
					return err
				}
				*(server.Messages)=*(server.Messages)+1

			}
			break
		}else{
			for i := 0; i < partes; i++ {
				s.Scan()
			}
		}
	}	

	return nil
}

