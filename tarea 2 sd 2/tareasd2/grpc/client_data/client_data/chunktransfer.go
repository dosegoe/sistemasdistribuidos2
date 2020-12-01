package client_data

import(
	"math/rand"
	"fmt"
	"context"
	"time"
	data_data "../../data_data/data_data"
	data_name "../../data_name/data_name"
	"strconv"
	"log"
)

//esta versión solo se utilizaría en la distribuida (ricart & agrawala)
func (server *Server) SendChunksToOtherDataNodes(chunks []*Chunk, fileName string) error{
	fmt.Println("starting")
	rand.Seed(time.Now().UnixNano())
	fmt.Println("getting the id of other datanode A")
	IDA:=server.FriendIdA
	

	fmt.Println("received node id")
	fmt.Printf("that id is: \n",IDA)

	defineorder:
	fmt.Println("defining order")
	numberOfChunks:=len(chunks)
	fmt.Println("number of chunks: %d",numberOfChunks)

	nodeidsorders:=[]int64{}
	//para asegurar de que cada datanode tenga como minimo un chunk, los primeros 3 chunks iran para cada nodo
	nodeidsorders=append(nodeidsorders,1)
	nodeidsorders=append(nodeidsorders,2)
	nodeidsorders=append(nodeidsorders,3)
	fmt.Print("actual order (first 3): ")
	fmt.Println(nodeidsorders)
	fmt.Println("shuffling...")
	rand.Shuffle(len(nodeidsorders),func(i, j int) { nodeidsorders[i], nodeidsorders[j] = nodeidsorders[j], nodeidsorders[i] })
	fmt.Println("shuffled")
	if numberOfChunks>=3{
		chunksleft:=numberOfChunks-3
		for chunksleft>0{
			nodeidsorders=append(nodeidsorders,int64(rand.Intn(3)+1))
			chunksleft-=1
		}
	} else{
		nodeidsorders=nodeidsorders[numberOfChunks:]
	}

	fmt.Print("final orders: ")
	fmt.Println(nodeidsorders)
	
	
	//aca debería haver un if que pregunte si es distribuido o centralizado. En este caso es distribuido
	var aprobation bool
	var err error
	if server.Mode=="distribuido"{
		fmt.Println("adding the filename: "+fileName+" to the order requests")
		directions:=[]*data_data.OrderReq{&data_data.OrderReq{
			Req: &data_data.OrderReq_FileName{
				FileName: fileName,
			},
		}}
		fmt.Println("adding the orderrequests")
		for i,order:=range nodeidsorders{
			directions=append(directions,&data_data.OrderReq{
				Req: &data_data.OrderReq_OrderData{
					OrderData: &data_data.OrderData{
						NodeId: order,
						ChunkId: int64(i),		
					},
				},
			})
		}
		err,aprobation=server.DistributedRequest(directions)
	} else {//centralizado
		fmt.Println("adding the filename: "+fileName+" to the order requests")
		directions:=[]*data_name.OrderReq{&data_name.OrderReq{
			Req: &data_name.OrderReq_FileName{
				FileName: fileName,
			},
		}}
		fmt.Println("adding the orderrequests")
		for i,order:=range nodeidsorders{
			directions=append(directions,&data_name.OrderReq{
				Req: &data_name.OrderReq_OrderData{
					OrderData: &data_name.OrderData{
						NodeId: order,
						ChunkId: int64(i),		
					},
				},
			})
		}
		err,aprobation=server.CentralizedRequest(directions,server.NameNode) 
	}
	
	fmt.Println("aproved?: ",aprobation)
	if err!=nil{
		return err
	}

	if aprobation==false{//si se rechazó la propuesta, se vuelve a armar otra
		goto defineorder
	}
	
	fmt.Println("saving the filename to transfer")
	//una vez se aprobaron, creo los arreglos en los que se enviarán la info. Parto con los nombres de archivos
	fileNameData:=&data_data.TransferReq{
		Req: &data_data.TransferReq_FileName{
			FileName: fileName,
	}}
	chunks1:=[]*data_data.TransferReq{fileNameData}
	chunks2:=[]*data_data.TransferReq{fileNameData}
	chunks3:=[]*data_data.TransferReq{fileNameData}
	fmt.Println("saving the chunks to transfer")
	//distribuyo que chunks irán para cada datanode
	for i,node:=range nodeidsorders{
		newchunk:=&data_data.TransferReq{
			Req:&data_data.TransferReq_DataChunk{
				DataChunk: &data_data.Chunk{
					Content: chunks[i].Content,
					ChunkId: int64(i),
				},
			},
		}
		switch node{
		case 1:
			chunks1=append(chunks1, newchunk)
		case 2:
			chunks2=append(chunks2, newchunk)
		case 3:
			chunks3=append(chunks3, newchunk)
		}
	}
	fmt.Println("sending chunks:")
	fmt.Println("IDA: ",IDA)
	fmt.Println("my nodeId: ",server.NodeId)
	//funcion definida en client_data.go: SendChunksToDataNode(chunks1)
	switch server.NodeId{
	case 1://soy 1
		if IDA==2{//A es 2 y  B es 3
			err=server.SendChunksToDataNode(chunks2,server.OtherDataNodeA)
			if err!=nil{
				return err
			}
			err=server.SendChunksToDataNode(chunks3,server.OtherDataNodeB)
			if err!=nil{
				return err
			}
		} else {//A es 3 y B es 2
			err=server.SendChunksToDataNode(chunks3,server.OtherDataNodeA)
			if err!=nil{
				return err
			}
			err=server.SendChunksToDataNode(chunks2,server.OtherDataNodeB)
			if err!=nil{
				return err
			}
		}
		server.SaveChunks(chunks1)
		//por ultimo, me "quedo con el resto"
	case 2:
		if IDA==1{//A es 1 y B es 3
			err=server.SendChunksToDataNode(chunks1,server.OtherDataNodeA)
			if err!=nil{
				return err
			}
			err=server.SendChunksToDataNode(chunks3,server.OtherDataNodeB)
			if err!=nil{
				return err
			}
		} else {//A es 3 y B es 2
			err=server.SendChunksToDataNode(chunks3,server.OtherDataNodeA)
			if err!=nil{
				return err
			}
			err=server.SendChunksToDataNode(chunks1,server.OtherDataNodeB)
			if err!=nil{
				return err
			}
		}
		//por ultimo, me "quedo con el resto"
		server.SaveChunks(chunks2)
	case 3:
		if IDA==1{//A es 1 y B es 2
			err=server.SendChunksToDataNode(chunks1,server.OtherDataNodeA)
			if err!=nil{
				return err
			}
			err=server.SendChunksToDataNode(chunks2,server.OtherDataNodeB)
			if err!=nil{
				return err
			}
		} else {//A es 2 y B es 1
			err=server.SendChunksToDataNode(chunks2,server.OtherDataNodeA)
			if err!=nil{
				return err
			}
			err=server.SendChunksToDataNode(chunks1,server.OtherDataNodeB)
			if err!=nil{
				return err
			}
		}
		//por ultimo, me "quedo con el resto"
		server.SaveChunks(chunks3)//funcion definida en client_data.go
	default:
		fmt.Println("node id not recognized")
	}

	fmt.Println("done")

	//TODO: ENVIAR INFO A LOG DEL NAMENODE, LUEGO DE SOLUCIONAR LA WEAA CON EL ALGORITMO
	//



	//stream,err:=server.NameNode.InformOrder(context.Background(),numberOfChunks int)
	//for i,totalChunks:=range nodeidsorders{
	//	stream.Send(data_name.OrderReq{...})
	//}

	//ALGORITMO RICART AGRAWALA (conjunción de este código con EntranceRequest e InformOrder)
	//Primero debemos cambiar el status del Servidor a WANTED
	server.Status = 1

	//primero debo enviar los request de entrada al log a los datanode con EntranceRequest
	resA,err := server.OtherDataNodeA.EntranceRequest(context.Background(),&data_data.EnReq{NodeId:server.NodeId})
	if err!=nil{
		return err
	} else {
		*(server.Messages)=*(server.Messages)+1
	}

	resB,err1 := server.OtherDataNodeB.EntranceRequest(context.Background(),&data_data.EnReq{NodeId:server.NodeId})
	if err1!=nil{
		return err
	} else {
		*(server.Messages)=*(server.Messages)+1
	}
	if resA.ResCode==data_data.OrderResCode_Yes && resB.ResCode==data_data.OrderResCode_Yes{
		//Los otros nodos permitieron el acceso, se cambio el status a HELD
		//se procede a escribir en el log
		server.Status = 2
		stream,err:=server.NameNode.InformOrder(context.Background())
		fmt.Println("comienzo envíos de datos en chunktransfer: \n")
		fmt.Println("Primero se envía el nombre")
		stream.Send(&data_name.OrderReq{ Req: &data_name.OrderReq_FileName{FileName: fileName+" "+strconv.Itoa(numberOfChunks),}})
		*(server.Messages)=*(server.Messages)+1
		
		//En cada iteración de este ciclo se envía los datos de un chunk
		for i,no_chunks:=range nodeidsorders{
			fmt.Printf("iteracion n°: %d, en el nodo: %d \n", i, no_chunks)
			if no_chunks == server.FriendIdA{
				stream.Send(&data_name.OrderReq{ Req: &data_name.OrderReq_OrderData{
					OrderData: &data_name.OrderData{ ChunkId: int64(i), NodeId: int64(no_chunks),},
				},})
			}
		}
			*(server.Messages)=*(server.Messages)+1

		//Se obtiene una respuesta del lado del servidor
		response, err := stream.CloseAndRecv()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("se recibio la respuesta en chunktransfer:")
		log.Println(response)
		if response.ResCode==data_name.OrderResCode_Yes{
			//como ya realizó la tarea se libera la zona critica y el recurso
			server.Status=0
			fmt.Println("Se libera el status del servidor\n")
		}
	}else{
		log.Fatal("No se pudo acceder al registro\n")
	}

	return nil

}
func (server *Server) CentralizedRequest(directions []*data_name.OrderReq, cli data_name.DataNameClient) (error, bool){
	fmt.Println("sending the requests to name node")
	stream,err:=cli.RequestOrder(context.Background())
	if err!=nil{
		return err,false
	}
	for _,data:=range directions{
		if err := stream.Send(data); err != nil {
			return err,false
		}
		*(server.Messages)=*(server.Messages)+1
	}
	reply,err:=stream.CloseAndRecv()
	if reply.ResCode==data_name.OrderResCode_Yes{
		fmt.Println("request accepted")
		return nil, true
	} else {
		fmt.Println("request rejected")
		return nil, false
	}


	return nil,true
}

func (server *Server) DistributedRequest(directions []*data_data.OrderReq) (error, bool){
//en la versión "centralizada", lo unico que cambiaría es que el request se le hace al name node y no a los otros datanodes
	//le pregunto al otro datanode A
	fmt.Println("sending the requests to node A")
	err,req:=server.RequestOrdersToNode(directions,server.OtherDataNodeA)
	if err!=nil{
		return err,false
	}
	if req==false{//se reachazo la propuesta de parte del nodo A
		return nil,false
	}
	fmt.Println("sending the requests to node B")
	err,req=server.RequestOrdersToNode(directions,server.OtherDataNodeB)
	if err!=nil{
		return err,false
	}
	if req==false{//se reachazo una propuesta
		return nil,false
	}
	return nil,true
}

func (server *Server) RequestOrdersToNode(directions []*data_data.OrderReq, cli data_data.DataDataClient) (error,bool) {
	fmt.Println("requesting order to other node...")
	stream,err:=cli.RequestOrder(context.Background())
	if err!=nil{
		return err,false
	}
	for _,data:=range directions{
		if err := stream.Send(data); err != nil {
			return err,false
		}
		*(server.Messages)=*(server.Messages)+1

	}
	reply,err:=stream.CloseAndRecv()
	if reply.ResCode==data_data.OrderResCode_Yes{
		fmt.Println("request accepted")
		return nil, true
	} else {
		fmt.Println("request rejected")
		return nil, false
	}

}

