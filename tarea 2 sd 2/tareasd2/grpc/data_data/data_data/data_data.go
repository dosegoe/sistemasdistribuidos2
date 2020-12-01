package data_data

import(
	"io"
	"fmt"
	"math/rand"
	"time"
	"io/ioutil"
	"strconv"
	"os"
	"context"
)

type Server struct{
	NodeId int64
	Probability float64
	FileChunksPath string
	Status int64 //0-> RELEASED; 1->WANTED; 2-> HELD
}

//ojo, hay que tener en mente que estas funciones se ejecutan del lado del datanode "servidor"!!1 (por mas confuso que sea)

func(server *Server) ChunksTransfer(stream DataData_ChunksTransferServer) error{

	var fileName string
	fmt.Println("receiving chunks")
	for{
		transReq,err:=stream.Recv()
		if err == io.EOF{
			fmt.Println("chunks for the file: "+fileName+" received")
			//server.printTotalChunks()
			return stream.SendAndClose(&TransferRes{
				ResCode:TransferResCode_Ok,
				Message:"Ok",
			})	
		}
		if err!=nil{
			return nil
		}
		//fmt.Printf("type: %T\n",upreq.Data)
		switch transReq.Req.(type){
		case *TransferReq_DataChunk:
			fmt.Println("receiving chunk ",transReq.Req.(*TransferReq_DataChunk).DataChunk.ChunkId)
			ioutil.WriteFile(
				server.FileChunksPath+"/"+fileName+"_chunk_"+strconv.Itoa(int(transReq.Req.(*TransferReq_DataChunk).DataChunk.ChunkId)),
				transReq.Req.(*TransferReq_DataChunk).DataChunk.Content,
				os.ModeAppend,
			)

		case *TransferReq_FileName:
			fileName=transReq.Req.(*TransferReq_FileName).FileName
			fmt.Println("receiving file of name: "+fileName)
		}
	}



	return nil
}

//este se usa solo en caso de que el algoritmo sea de exclusion distribuida:
func(server *Server) RequestOrder(stream DataData_RequestOrderServer) error{

	rand.Seed(time.Now().UnixNano())
	fmt.Println("receiving order request")
	var fileName string
	for{
		ordReq,err:=stream.Recv()
		if err == io.EOF{
			fmt.Print("sending response: ")
			//acá se simula la aceptación o rechazo del orden, según una "probabilidad de exito"
			r:=rand.Float64()
			if r<server.Probability{
				fmt.Println("sipo approbing ")
				return stream.SendAndClose(&OrderRes{
					ResCode:OrderResCode_Yes,
				})
			} else {
				fmt.Println("rejecting")
				return stream.SendAndClose(&OrderRes{
					ResCode:OrderResCode_No,
				})
			}
			//server.printTotalChunks()
				
		}
		if err!=nil{
			return nil
		}
		//fmt.Printf("type: %T\n",upreq.Data)
		switch ordReq.Req.(type){
		case *OrderReq_OrderData:
			fmt.Print("received request: ")
			fmt.Print(ordReq.Req.(*OrderReq_OrderData).OrderData.ChunkId)
			fmt.Print(" in node: ")
			fmt.Println(ordReq.Req.(*OrderReq_OrderData).OrderData.NodeId)
		case *OrderReq_FileName:
			fileName=ordReq.Req.(*OrderReq_FileName).FileName
			fmt.Println("receiving file of name: "+fileName)
		}
	}


	return nil
}

func(server *Server)GetId(ctx context.Context, idreq *IdReq) (*IdRes, error){
	fmt.Println("getting id request")
	fmt.Printf("my node id is: %d\n",server.NodeId)
	return &IdRes{
		NodeId: server.NodeId,
	},nil
}

//esta funcion envía a los demas datanodes una solicitud y recibe una respuesta.
func(server *Server) EntranceRequest(ctx context.Context, in *EnReq) (*EnRes, error){
	//2 -> statud=HELD
	//1 -> status=WANTED
	sendRequest:
	if server.Status == 2 || (server.Status==1 && (server.NodeId<in.NodeId)){
		time.Sleep(130 * time.Millisecond)
		goto sendRequest
	}else{
		return &EnRes{
			NodeId: server.NodeId,
			ResCode: OrderResCode_Yes,
		},nil
	}
}