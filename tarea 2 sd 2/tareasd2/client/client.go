package main

import(
	transform "./transform"
	grpc "google.golang.org/grpc"
	client_data "../grpc/client_data/client_data"
	client_name "../grpc/client_name/client_name"
	"fmt"
	"context"
	"log"
	"sort"
	"math/rand"
	"time"
	"io"
	"os"
	"bufio"
	"strings"
	"strconv"
)

func SetNameNodeConnection()(client_name.ClientNameClient, *grpc.ClientConn){
	entry:
		fmt.Println("ingrese dirección IP del servidor del name node (en el formato: 255.255.255.255)")
		var IPaddr string
		fmt.Scanln(&IPaddr)
		fmt.Println("ingrese el numero de puerto en el que el name node está escuchando a un cliente")
		var PortNum string
		fmt.Scanln(&PortNum)

		CompleteAddr:=IPaddr+":"+PortNum
		fmt.Println(CompleteAddr)
		conn, err:=grpc.Dial(CompleteAddr,grpc.WithInsecure(),grpc.WithBlock())
		//defer conn.Close()
	if err!=nil{
		goto entry
	}
	cdn:=client_name.NewClientNameClient(conn)
	fmt.Println("conexión a namenode creada")
	return cdn, conn
}

func SetDataNodeConnection(id string)(client_data.ClientDataClient, *grpc.ClientConn){
	entry:
		fmt.Println("ingrese dirección IP del servidor del data node "+id+" (en el formato: 255.255.255.255)")
		var IPaddr string
		fmt.Scanln(&IPaddr)
		fmt.Println("ingrese el numero de puerto en el que el data node "+id+" está escuchando a un cliente")
		var PortNum string
		fmt.Scanln(&PortNum)

		CompleteAddr:=IPaddr+":"+PortNum
		fmt.Println(CompleteAddr)
		conn, err:=grpc.Dial(CompleteAddr,grpc.WithInsecure(),grpc.WithBlock())
		//defer conn.Close()
	if err!=nil{
		goto entry
	}
	cnn:=client_data.NewClientDataClient(conn)
	fmt.Println("conexión a datanode "+id+" creada")
	return cnn, conn
}


func UploadFileToDataNodes(
	cdn1 client_data.ClientDataClient,
	cdn2 client_data.ClientDataClient,
	cdn3 client_data.ClientDataClient,fileName string) error{
	fmt.Println("attempting to download chunks")
	chunks,err:=transform.FileToChunks(uploadPath,fileName)
	if err!=nil{
		return err
	}

	var cdn client_data.ClientDataClient
	// se elige uno de los 3 aleatoriamente
	rand.Seed(time.Now().UnixNano())
	
	nid:=int64(rand.Intn(3)+1)
	fmt.Println("datanode elegido para transferir: ",nid)
	switch nid{
	case int64(1):
		cdn=cdn1
	case int64(2):
		cdn=cdn2
	case int64(3):
		cdn=cdn3
	}
	//cdn=cdn1



	fmt.Println("invoking stream")
	stream, err:=cdn.UploadFile(context.Background())
	if err!=nil {
		return err
	}
	fmt.Println("stream invoked")
	data:=&client_data.UploadReq{
		Req: &client_data.UploadReq_FileName{
			FileName: fileName,
		},
	}
	fmt.Println("sending filename")
	if err := stream.Send(data); err != nil {
		return err
	}
	fmt.Println("sending chunks")
	for i,chunk:=range chunks{
		data:=&client_data.UploadReq{
			Req: &client_data.UploadReq_DataChunk{
				DataChunk: &client_data.Chunk{
					Content: chunk,
					ChunkId: int64(i),
				},
			},
		}
		if err := stream.Send(data); err != nil {
			return err
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	log.Printf("Route summary: %v", reply)

	return nil
}

func RequestChunksOrders(cnn client_name.ClientNameClient, fileName string) ([]client_name.OrderRes,error){
	stream,err:=cnn.ChunksOrder(context.Background(),&client_name.OrderReq{Filename:fileName})
	if err!=nil{
		return nil,err
	}
	orderchunks:=[]client_name.OrderRes{}

	for{
		in,err:=stream.Recv()
		if err == io.EOF {
			return orderchunks,nil
		}
		if err != nil {
			return nil,err
		}
		order:=client_name.OrderRes{
			ChunkId: in.GetChunkId(),
			NodeId: in.GetNodeId(),
		}
		orderchunks=append(orderchunks,order)
	}

	return orderchunks,nil
}



func errCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}


func RequestChunksContents(reqs []*client_data.DownloadReq, cdn client_data.ClientDataClient, fileName string) ([]*client_data.Chunk ,error){
	chunks:=[]*client_data.Chunk{}
	//fmt.Println("invoking stream")
	fmt.Println("requesting orders")
	stream, err:=cdn.DownloadFile(context.Background())
	
	//fmt.Println("creating filename")
	DonwnloadFileName:=&client_data.DownloadReq{
		Req: &client_data.DownloadReq_FileName{
			FileName: fileName,
		},
	}
	if err!=nil{
		return nil,err
	}
	//fmt.Println("sending filename")
	if err:=stream.Send(DonwnloadFileName); err!=nil{
		return nil,err
	}

	for _,req:=range reqs{
		fmt.Println("requesting chunk of id: ",req.Req)
		if err:=stream.Send(req); err!=nil{
			return nil,err
		}
		fmt.Println("sended, receiving response")
		downloadRes,err:=stream.Recv()
		if err!=nil{
			return nil,err
		}
		fmt.Print("received chunk of id: ",downloadRes.ChunkId)
		chunks=append(chunks,downloadRes)
	}

	return chunks,nil
}

//descarga archivo
func DownloadFileFromDataNodes(
	cdn1 client_data.ClientDataClient,
	cdn2 client_data.ClientDataClient,
	cdn3 client_data.ClientDataClient,
	cnn client_name.ClientNameClient, fileName string) error {

		fmt.Println("attempting to download chunks")

		orderchunks, err:=RequestChunksOrders(cnn,fileName)
		if err!=nil{
			return nil
		}
	
		ordersDN1:=[]*client_data.DownloadReq{}
		ordersDN2:=[]*client_data.DownloadReq{}
		ordersDN3:=[]*client_data.DownloadReq{}
	
		allChunks:=[]*client_data.Chunk{}
		
		//separar los requests que haremos por datanode
		for _,order:=range orderchunks{
			switch(order.NodeId){
			case int64(1):
				ordersDN1=append(ordersDN1,&client_data.DownloadReq{
					Req: &client_data.DownloadReq_ChunkId{
						ChunkId: order.ChunkId,
					},
				})
			case int64(2):
				ordersDN2=append(ordersDN2,&client_data.DownloadReq{
					Req: &client_data.DownloadReq_ChunkId{
						ChunkId: order.ChunkId,
					},
				})
			case int64(3):
				ordersDN3=append(ordersDN3,&client_data.DownloadReq{
					Req: &client_data.DownloadReq_ChunkId{
						ChunkId: order.ChunkId,
					},
				})
			}
				//bla bla bla...
		}
		//pedir chunks correspondientes a DN1
		chunks1,err:=RequestChunksContents(ordersDN1,cdn1,fileName)
		if err!=nil{
			return err
		}
		allChunks=append(allChunks,chunks1...)
		//pedir chunks correspondientes a DN2
		chunks2,err:=RequestChunksContents(ordersDN2,cdn2,fileName)
		if err!=nil{
			return err
		}
		allChunks=append(allChunks,chunks2...)
		//pedir chunks correspondientes a DN3
		chunks3,err:=RequestChunksContents(ordersDN3,cdn3,fileName)
		if err!=nil{
			return err
		}
		allChunks=append(allChunks,chunks3...)
		
		//ordenar los chunks por id
		fmt.Println("all chunks received, ordering by id")
		sort.Slice(allChunks, func(i, j int) bool {
			return allChunks[i].ChunkId < allChunks[j].ChunkId
		})
	
		chunksData:=[][]byte{}
		fmt.Println("chunks:")
		for _,chunk:=range allChunks{
			fmt.Println(chunk.ChunkId)
			chunksData=append(chunksData,chunk.Content)
		}
	
		fmt.Println("merging bytes to file")
		err=transform.ChunksToFile(chunksData,fileName,downloadPath)
		if err!=nil{
			return err
		}
		
		fmt.Println("done! :)")
		return nil

}

var(
	downloadPath string 
	uploadPath string
	//cdn1 client_data.ClientDataClient
	//cdn2 client_data.ClientDataClient
	//cdn3 client_data.ClientDataClient
	//cnn client_name.ClientNameClient
	
)


func main(){

	downloadPath="client/filesdown"
	uploadPath="client/filesup"

	cdn1, connd1:=SetDataNodeConnection("1")
	defer connd1.Close()
	cdn2, connd2:=SetDataNodeConnection("2")
	defer connd2.Close()
	cdn3, connd3:=SetDataNodeConnection("3")
	defer connd3.Close()
	cnn, conncn:=SetNameNodeConnection()
	defer conncn.Close() 


	for{
		fmt.Println("Selecciona una acción: ")
		fmt.Println("\"1\": Subir Archivo\n")
		fmt.Println("\"2\": Descargar Archivo\n")
		fmt.Println("\"3\": Salir\n")

		var Accion string
		fmt.Scanln(&Accion)

		switch Accion{
		case "1":
			//subir archivo
			var fileUpload string
			fmt.Println("Ingrese el título del archivo (incluir extensión \".pdf\"): \n")
			fmt.Scanln(&fileUpload)
			err:=UploadFileToDataNodes(cdn1,cdn2,cdn3,fileUpload)
			if err!=nil{
				log.Fatalf("error subiendo archivos: %v",err)
			}

		case "2":
			//descargar archivo

			f, err5 := os.Open("namenode/log.txt")
			errCheck(err5)
			s := bufio.NewScanner(f)
			//lee primero una vez para obtener el título y el número de partes
			for s.Scan(){
				if s.Text() == ""{
					break
				}
				errCheck(s.Err())
				muchotexto := s.Text()
				separados := strings.Split(muchotexto, " ")
				nombre := separados[0]
				partes, err6 := strconv.Atoi(separados[1])
				errCheck(err6)
				fmt.Println(nombre)
				for i := 0; i < partes; i++ {
					s.Scan()
				}
			}
			f.Close()
			var fileDownload string
			fmt.Println("Ingrese el título del archivo a descargar (incluir extensión \".pdf\"): \n")
			fmt.Scanln(&fileDownload)
			err:=DownloadFileFromDataNodes(cdn1,cdn2,cdn3,cnn,fileDownload)
			if err!=nil{
				log.Fatalf("error descargando archivo: %v",err)
			}	
		
		case "3":
			fmt.Println("Adios")
			return 

		default:
			fmt.Println("comando inválido, ingresar de nuevo")
		}
	}
	
	//err=transform.ChunksToFile(chunks,fileName,downloadPath)
	
}