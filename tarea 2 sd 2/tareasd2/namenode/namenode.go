package main

import(
	"fmt"
	client_name "../grpc/client_name/client_name"
	data_name "../grpc/data_name/data_name"
	"net"
	grpc "google.golang.org/grpc"
	"log"
	"time"
	"bufio"
	"os"
)

func getIPAddr() string{
	addrs, err := net.InterfaceAddrs()
    if err != nil {
        return ""
    }
    for _, address := range addrs {
        // check the address type and if it is not a loopback the display it
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return ""
}

func ListenToDataNodes(ipaddr string){
	//agregar para ips.
	go ListenToDN(ipaddr,"8993","1")
	go ListenToDN(ipaddr,"8992","2")
	go ListenToDN(ipaddr,"8991","3")
}

func ListenToDN(IPAddr string, PortNum string, NodeNum string) error {
	portstring:=":"+PortNum
	lis, err := net.Listen("tcp", portstring)
	if err!=nil{
		fmt.Printf("Error escuchando al datanode %s en el puerto :%s: %v", NodeNum,portstring, err)
		return err	
	} else{
		fmt.Println("Escuchado datanode "+NodeNum +" desde: ",IPAddr+portstring)
	}
	cli_serv:=data_name.Server{
		Probability: 0.9,
		BookNum: bookNum,
		Messages: messages,
	}

	grpcServer:=grpc.NewServer()
	data_name.RegisterDataNameServer(grpcServer, &cli_serv)

	if err:=grpcServer.Serve(lis); err!=nil{
		log.Printf("No se pudo servir en grpc en el puerto: %s; %v\n",portstring, err)
		return err
	} else {
		fmt.Println("Servidor comunicandose con datanode ",NodeNum)
	}
	return nil
}

//namenode escucha a un cliente desde un puerto PortNum
func ListenToClient(IPAddr string, PortNum string) error {
	portstring:=":"+PortNum
	lis, err := net.Listen("tcp", portstring)
	if err!=nil{
		fmt.Printf("Error escuchando al cliente en el puerto :%s: %v", portstring, err)
		return err	
	} else{
		fmt.Println("Escuchado cliente desde: ",IPAddr+portstring)
	}

	cli_serv:=client_name.Server{
		Messages: messages,
	}

	grpcServer:=grpc.NewServer()
	client_name.RegisterClientNameServer(grpcServer, &cli_serv)

	if err:=grpcServer.Serve(lis); err!=nil{
		log.Printf("No se pudo servir en grpc en el puerto: %s; %v\n",portstring, err)
		return err
	} else {
		fmt.Println("Servidor comunicandose con cliente")
	}
	return nil
}

var bookNum *int;
var messages *int

func main(){

	bn:=0
	bookNum=&bn
	msgs:=0
	messages=&msgs

	IPAddr:=getIPAddr()
	ListenToDataNodes(IPAddr)
	time.Sleep(130 * time.Millisecond)
	file, err := os.OpenFile("namenode/Dnodes.txt", os.O_RDWR|os.O_CREATE, 0666)  
	if err != nil {
		log.Fatal(err)
	}
	
	defer file.Close()

	file2, err := os.OpenFile("namenode/log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)  
	if err != nil {
		log.Fatal(err)
	}
	defer file2.Close()


	w := bufio.NewWriter(file)
	var IPs string
	fmt.Println("Ingrese IP DataNode 1")
	fmt.Scanln(&IPs)
	_, err1 := w.WriteString( "1 "+IPs+"\n")
	if err1 != nil {
		log.Fatal(err1)
		}
	fmt.Println("Ingrese IP DataNode 2")
	fmt.Scanln(&IPs)
	_, err2 := w.WriteString( "2 "+IPs+"\n")
	if err2 != nil {
		log.Fatal(err2)
		}
	fmt.Println("Ingrese IP DataNode 3")
	fmt.Scanln(&IPs)
	_, err3 := w.WriteString( "3 "+IPs+"\n")
	if err3 != nil {
		log.Fatal(err3)
		}
	w.Flush()
	file.Close()
	for{
		fmt.Println("comandos:")
		fmt.Println("\"listen\": agregar nuevo puerto para escuchar a cliente")
		fmt.Println("\"quit\": salir")


		var Command string
		fmt.Scanln(&Command)

		switch Command{
		case "listen":
			fmt.Println("ingrese puerto para escuchar a cliente")
			var PortNum string
			fmt.Scanln(&PortNum)
			go ListenToClient(IPAddr,PortNum)
			time.Sleep(130 * time.Millisecond)

		case "quit":
			fmt.Println("número de mensajes enviados por el NameNode: ",*messages)
			fmt.Println("adios")
			return
			
		default:
			fmt.Println("comando inválido, ingresar de nuevo")
		}
	}
}