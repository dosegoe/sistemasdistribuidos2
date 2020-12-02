package main

import(
	grpc "google.golang.org/grpc"
	client_data "../grpc/client_data/client_data"
	data_data "../grpc/data_data/data_data"
	data_name "../grpc/data_name/data_name"
	"net"
	"fmt"
	//"strconv"
	"log"
	"time"
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

func ConnectToNameNode(NodeId int64)(data_name.DataNameClient, *grpc.ClientConn){
	entry:
		fmt.Println("ingrese dirección IP name node (en el formato: 255.255.255.255)")
		var IPaddr string
		fmt.Scanln(&IPaddr)
		var PortNum string
		switch NodeId{
		case int64(1):
			PortNum="8993"
		case int64(2):
			PortNum="8992"
		case int64(3):
			PortNum="8991"
		default:
			log.Fatalf("error imposible conectarse con name node, id del nodo es inválido")
		}

		CompleteAddr:=IPaddr+":"+PortNum
		fmt.Println(CompleteAddr)
		conn, err:=grpc.Dial(CompleteAddr,grpc.WithInsecure(),grpc.WithBlock())
		//defer conn.Close()
	if err!=nil{
		goto entry
	}
	dnc:=data_name.NewDataNameClient(conn)
	fmt.Println("conexión a name node creada")
	return dnc, conn
}

func ListenToClient(IPAddr string, PortNum string) error {
	portstring:=":"+PortNum
	lis, err := net.Listen("tcp", portstring)
	if err!=nil{
		fmt.Printf("Error escuchando al cliente en el puerto :%s: %v", portstring, err)
		return err	
	} else{
		fmt.Println("Escuchado cliente desde: ",IPAddr+portstring)
	}

	//filechunks:=[](fc.FileChunk){}

	cli_serv:=client_data.Server{
		//TotalChunks: &(filechunks),
		FileChunksPath:"datanode/files",
		OtherDataNodeA: OtherDNodeA,
		OtherDataNodeB: OtherDNodeB,
		FriendIdA:IdNodeA,
		FriendIdB:IdNodeB,
		NodeId: NodeId,
		NameNode: NameNode,
		Mode: Mode,
		Messages: Messages,
	}

	transServer:=grpc.NewServer()
	client_data.RegisterClientDataServer(transServer, &cli_serv)

	if err:=transServer.Serve(lis); err!=nil{
		log.Printf("No se pudo servir en grpc en el puerto: %s; %v\n",portstring, err)
		return err
	} else {
		fmt.Println("Servidor comunicandose con cliente")
	}
	return nil

}



func ListenToDataNodes(IPAddr string, NodeId int64, Probability float64)error{
	switch NodeId{
	case 1:
		go ListenToDN(IPAddr,"8997",NodeId,Probability)
		go ListenToDN(IPAddr,"8995",NodeId,Probability)
	case 2:
		go ListenToDN(IPAddr,"8999",NodeId,Probability)
		go ListenToDN(IPAddr,"8994",NodeId,Probability)
	case 3:
		go ListenToDN(IPAddr,"8998",NodeId,Probability)
		go ListenToDN(IPAddr,"8996",NodeId,Probability)
	}
	
	return nil

}

func ListenToDN(IPAddr string, PortNum string, NodeId int64, Probability float64) error {
	portstring:=IPAddr+":"+PortNum
	fmt.Println("esperando a namenode")
	lis, err := net.Listen("tcp", portstring)
	if err!=nil{
		fmt.Printf("Error escuchando a otro data node en el puerto :%s: %v", portstring, err)
		return err	
	} else{
		fmt.Printf("node id :%d\n",NodeId)
		fmt.Println("Escuchado data node desde: ",IPAddr+portstring)
	}

	//filechunks:=[](fc.FileChunk){}

	cli_serv:=data_data.Server{
		//TotalChunks: &(filechunks),
		FileChunksPath:"datanode/files",
		NodeId: NodeId,
		Probability: Probability,
		Messages: Messages,
	}

	transServer:=grpc.NewServer()
	data_data.RegisterDataDataServer(transServer, &cli_serv)

	if err:=transServer.Serve(lis); err!=nil{
		log.Printf("No se pudo servir en grpc en el puerto: %s; %v\n",portstring, err)
		return err
	} else {
		fmt.Println("Servidor comunicandose con datanode")
	}
	return nil

}

func ConnectToDataNodes(NodeId int64){
	//NodeId int64
	entry:
	switch NodeId{
	case 1:
		fmt.Println("este es el datanode 1")
		fmt.Println("ingrese dirección IP del Data Node 2, en el formato: '255.255.255.255'")
		var IPaddrA string
		fmt.Scanln(&IPaddrA)
		fmt.Println("ingrese dirección IP del Data Node 3, en el formato: '255.255.255.255'")
		var IPaddrB string
		fmt.Scanln(&IPaddrB)
		err:=ConnectToDN("8999",IPaddrA,"A",2)
		if err!=nil{
			goto entry
		}
		err=ConnectToDN("8998",IPaddrB,"B",3)
		if err!=nil{
			goto entry
		}
		fmt.Printf("La dirección ip de A es: %s", IPaddrA)
		fmt.Printf("La dirección ip de B es: %s", IPaddrB)
	case 2:
		fmt.Println("este es el datanode 2")
		fmt.Println("ingrese dirección IP del Data Node 1, en el formato: '255.255.255.255'")
		var IPaddrA string
		fmt.Scanln(&IPaddrA)
		fmt.Println("ingrese dirección IP del Data Node 3, en el formato: '255.255.255.255'")
		var IPaddrB string
		fmt.Scanln(&IPaddrB)
		err:=ConnectToDN("8997",IPaddrA,"A",1)
		if err!=nil{
			goto entry
		}
		err=ConnectToDN("8996",IPaddrB,"B",3)
		if err!=nil{
			goto entry
		}
		fmt.Printf("La dirección ip de A es: %s", IPaddrA)
		fmt.Printf("La dirección ip de B es: %s", IPaddrB)
	case 3:
		fmt.Println("este es el datanode 3")
		fmt.Println("ingrese dirección IP del Data Node 1, en el formato: '255.255.255.255'")
		var IPaddrA string
		fmt.Scanln(&IPaddrA)
		fmt.Println("ingrese dirección IP del Data Node 2, en el formato: '255.255.255.255'")
		var IPaddrB string
		fmt.Scanln(&IPaddrB)
		err:=ConnectToDN("8995",IPaddrA,"A",1)
		if err!=nil{
			goto entry
		}
		err=ConnectToDN("8994",IPaddrB,"B",2)
		if err!=nil{
			goto entry
		}
		fmt.Printf("La dirección ip de A es: %s", IPaddrA)
		fmt.Printf("La dirección ip de B es: %s", IPaddrB)
	}
}

func ConnectToDN(PortNum string, IPAddr string, letter string, friendId int64)(error){
	CompleteAddr:=IPAddr+":"+PortNum
	fmt.Println("esperando a datanode ",friendId)
	conn, err:=grpc.Dial(CompleteAddr,grpc.WithInsecure(),grpc.WithBlock())
		//defer conn.Close()
	if err!=nil{
		return err
	}
	fmt.Println("conectado a datanode ",friendId)
	if letter=="A"{
		OtherDNodeA=data_data.NewDataDataClient(conn)
		IdNodeA=friendId
	} else{//=="B"
		OtherDNodeB=data_data.NewDataDataClient(conn)
		IdNodeB=friendId
	}
	return nil
}

var NodeId int64
var OtherDNodeA data_data.DataDataClient
var IdNodeA int64
var OtherDNodeB data_data.DataDataClient
var IdNodeB int64
var NameNode data_name.DataNameClient
var Mode string
var Messages *int

func main(){
	msgs:=0
	Messages=&msgs
	IPAddr:=getIPAddr()
	//log.Println(IpAddr)
	fmt.Println("dirección IP de dataNode: ",IPAddr)

	datanodeid:
		fmt.Println("ingrese el id del datanode: '1', '2' o '3'")
		fmt.Scanln(&NodeId)

	if NodeId!=1&&NodeId!=2&&NodeId!=3 {
		fmt.Println("ID incorrecto, ingrese '1', '2' o '3'")
		goto datanodeid
	}
	mode:
		fmt.Println("seleccione el modo: ")
		fmt.Println("ingrese '1' para modo excluyente distribuido")
		fmt.Println("ingrese '2' para modo excluyente centralizado")
		var mod string 
		fmt.Scanln(&mod)
	switch mod{
	case "1": 
		fmt.Println("elegido modo distribuido")
		Mode="distribuido"
	case "2":
		fmt.Println("elegido modo centralizado")
		Mode="centralizado"
	default:
		fmt.Println("opción inválida")
		goto mode
	}


	fmt.Println("conectandose a name node")
	dnc,conn:=ConnectToNameNode(NodeId)
	defer conn.Close()
	NameNode=dnc


	fmt.Println("escuchando a otros nodos")
	if err:=ListenToDataNodes(IPAddr,NodeId,0.99); err!=nil{
		log.Fatalf("error escuchando: %v",err)
	}
	time.Sleep(120 * time.Millisecond)
	fmt.Println("conectandose con otros nodos")
	ConnectToDataNodes(NodeId)
		//todo: ingresar conexiones del resto de weas
	
	for{
		fmt.Println("comandos:")
		fmt.Println("\"listen\": agregar nuevo puerto para escuchar a cliente")
		fmt.Println("\"quit\": salir")


		var Command string
		fmt.Scanln(&Command)

		switch Command{
		case "listen":
			var PortNum string
			fmt.Println("ingrese puerto para escuchar a cliente")
			fmt.Scanln(&PortNum)
			go ListenToClient(IPAddr,PortNum)
			time.Sleep(200 * time.Millisecond)

		case "quit":
			fmt.Println("número de mensajes enviados por este DataNode: ",*Messages)
			fmt.Println("adios")
			return
			
		default:
			fmt.Println("comando inválido, ingresar de nuevo")
		}
	}

		
}