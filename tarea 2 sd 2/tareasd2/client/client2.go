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

func DownloadFileFromDataNodes(
	cdn1 client_data.ClientDataClient,
	cdn2 client_data.ClientDataClient,
	cdn3 client_data.ClientDataClient,
	cnn client_name.ClientNameClient, fileName string) error {

	//actualmente solo manda a un solo nodo
	//TODO: pedirle al namenode el verdadero orden!!!
	/* algo así como:
		orderchunks, err:=RequestChunksOrders(cnn,fileName)
		if err!=nil{
			return nil
		}

		for i,order:=range orderchunks {
			bla bla bla...
		}
	*/


	fmt.Println("invoking stream")
	stream, err:=cdn1.DownloadFile(context.Background())//esto funciona solo si los 3 datanodes comparten los chunks
	fmt.Println("creating filename")
	DonwnloadFileName:=&client_data.DownloadReq{
		Req: &client_data.DownloadReq_FileName{
			FileName: fileName,
		},
	}
	if err!=nil{
		return err
	}
	fmt.Println("sending filename")
	if err:=stream.Send(DonwnloadFileName); err!=nil{
		return nil
	}

	chunksData:=[][]byte{}
	chunks:=[]client_data.Chunk{}
	//TODO: pedirle el verdadero orden al namenode
	for i:=0;i<=7;i++{
		fmt.Printf("requesting chunk of id: %d \n",i)
		chunkReq:=&client_data.DownloadReq{
			Req: &client_data.DownloadReq_ChunkId{
				ChunkId: int64(i),
			},
		}
		fmt.Println("sending request")
		if err:=stream.Send(chunkReq); err!=nil{
			return nil
		}
		fmt.Println("sended, receiving response")
		downloadRes,err:=stream.Recv()
		if err!=nil{
			return err
		}
		fmt.Println("response receiving, adding to chunks slice")
		chunks=append(chunks,*downloadRes)	
	}
	//ordenar los chunks por id
	fmt.Println("all chunks received, ordering by id")
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].ChunkId < chunks[j].ChunkId
	})
	fmt.Println("chunks:")
	for _,chunk:=range chunks{
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



	fileName:="El_maravilloso_Mago_de_Oz-L._Frank_Baum.pdf"
	err:=UploadFileToDataNodes(cdn1,cdn2,cdn3,fileName)
	if err!=nil{
		log.Fatalf("error subiendo archivos: %v",err)
	}

	err=DownloadFileFromDataNodes(cdn1,cdn2,cdn3,cnn,fileName)
	if err!=nil{
		log.Fatalf("error descargando archivo: %v",err)
	}
	
	//err=transform.ChunksToFile(chunks,fileName,downloadPath)
	
}