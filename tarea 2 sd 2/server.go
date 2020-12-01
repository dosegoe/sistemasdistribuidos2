package main

import(
	//"io"
	"fmt"
	//"math/rand"
	//"time"
	"net"
	//"io/ioutil"
	//"strconv"
	//"os"
	//"context"
	"log"
	client_name "./tareasd2/grpc/client_name/client_name" 
	grpc "google.golang.org/grpc"
)
//strconv.Itoa() para convertir entero en strings

func main(){
	port := "9010"
	//ipaddr := "192.168.1.82"
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil{
		log.Fatal(err)
	}else{ 
		fmt.Println("está escuchando en puerto: 9000") 
	}
	
	grpcServer := grpc.NewServer()

	client_name.RegisterClientNameServer(grpcServer, &client_name.Server{})

	if err1:=grpcServer.Serve(lis); err1!=nil{
		log.Printf("No se pudo servir en grpc en el puerto: %s; %v\n",port, err)
		return
	} 
}
