package client_data

import(
	"io"
	"os"
	"io/ioutil"
	"strconv"
	//"math/rand"
	"fmt"
	data_data "../../data_data/data_data"
	data_name "../../data_name/data_name"
	//"time"
	"context"
)

type Server struct{
	FileChunksPath string
	OtherDataNodeA data_data.DataDataClient
	OtherDataNodeB data_data.DataDataClient
	FriendIdA int64
	FriendIdB int64
	NodeId int64
	Status int64
	NameNode data_name.DataNameClient
	Mode string //excluido o centralizado
	//NameNodeCli data_name.DataNameClient
	//Mode string //(excluido o centralizado)
	Messages *int
}

func (server *Server) UploadFile(stream ClientData_UploadFileServer) error{
	var fileName string

	totalChunks:=[]*Chunk{}
	
	for{
		upReq,err:=stream.Recv()
		if err == io.EOF{
			fmt.Println("sending chunks to friends")

			err=server.SendChunksToOtherDataNodes(totalChunks, fileName)//ver archivo "chunktransfer"".go, en este mismo paquete

			if err!=nil{
				fmt.Println("error")
				*(server.Messages)=*(server.Messages)+1
				return stream.SendAndClose(&UploadRes{
					ResCode:UploadResCode_Failed,
					Message:fmt.Sprintf("Failed: %v\n",err),
				})
			}
			
			//server.printTotalChunks()
			*(server.Messages)=*(server.Messages)+1
			return stream.SendAndClose(&UploadRes{
				ResCode:UploadResCode_Ok,
				Message:"Ok",
			})	
		}
		if err!=nil{
			*(server.Messages)=*(server.Messages)+1
			return stream.SendAndClose(&UploadRes{
				ResCode:UploadResCode_Failed,
				Message:fmt.Sprintf("Failed: %v\n",err),
			})
		}
		//fmt.Printf("type: %T\n",upreq.Data)
		switch upReq.Req.(type){
		case *UploadReq_DataChunk:
			totalChunks=append(totalChunks, upReq.Req.(*UploadReq_DataChunk).DataChunk)
			/*
			ioutil.WriteFile(
				server.FileChunksPath+"/"+fileName+"_chunk_"+strconv.Itoa(int(upReq.Req.(*UploadReq_DataChunk).DataChunk.ChunkId)),
				upReq.Req.(*UploadReq_DataChunk).DataChunk.Content,
				os.ModeAppend,
			)*/

		case *UploadReq_FileName:
			fileName=upReq.Req.(*UploadReq_FileName).FileName
			fmt.Println("receiving file of name: "+fileName)
		}
	}
	
	return nil
}



func (server *Server) DownloadFile(stream ClientData_DownloadFileServer) error{
	//lo primero que el datanode recive es el nombre del archivo
	fmt.Println("downloading file")
	in,err:=stream.Recv()
	if err != nil {
		return err
	}
	var fileName string
	fileName=in.GetFileName()
	fmt.Println("received file name: "+fileName)
	if fileName==""{
		return err
	}
	
	//chunkIds:=[]int64{}
	for{
		in,err:=stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		newChunkId:=in.GetChunkId()
		//chunkIds=append(chunkIds,newChunkId)
		fmt.Printf("received chunk id: %d \n ",newChunkId)
		chunkBytes,err:=server.getFileChunk(fileName,newChunkId)
		if err!=nil{
			return err
		}

		*(server.Messages)=*(server.Messages)+1
		stream.Send(&Chunk{
			Content: chunkBytes,
			ChunkId: newChunkId,
		})
	}
	
	
	return nil
}

func (server *Server) getFileChunk(fileName string,chunkId int64) ([]byte, error){
	filePath:=server.FileChunksPath+"/"+fileName+"_chunk_"+strconv.Itoa(int(chunkId)-1)
	fmt.Println("opening chunk path: "+filePath)

	err:=os.Chmod(filePath,0777)
	if err!=nil{
		return nil,err
	}
	chunkFile,err:=os.Open(filePath)
	if err!=nil{
		return nil,err
	}
	chunkBytes,err:=ioutil.ReadAll(chunkFile)
	if err != nil {
		//fmt.Println(err)
		return nil,err
	}
	return chunkBytes,nil
}

func(server *Server) SendChunksToDataNode(chunks []*data_data.TransferReq, cli data_data.DataDataClient)(error){
	fmt.Println("sending chunks to other data node")

	stream,err:=cli.ChunksTransfer(context.Background())
	if err!=nil{
		return err
	}
	for _,chunk:=range chunks{
		if err:=stream.Send(chunk);err!=nil{
			*(server.Messages)=*(server.Messages)+1
			return err
		}
	}

	res,err:=stream.CloseAndRecv()
	if err!=nil{
		return err
	} else {
		fmt.Println("response received: "+res.Message)
	}

	return nil
}

//almacena localmente un conjunto de chunks
func (server *Server) SaveChunks(chunks []*data_data.TransferReq){
	fmt.Println("saving left chunks locally")
	var fileName string
	for _,chunk:=range chunks{
		switch chunk.Req.(type){
		case *data_data.TransferReq_DataChunk:
			fmt.Println("receiving ",chunk.Req.(*data_data.TransferReq_DataChunk).DataChunk.ChunkId)
			ioutil.WriteFile(
				server.FileChunksPath+"/"+fileName+"_chunk_"+strconv.Itoa(int(chunk.Req.(*data_data.TransferReq_DataChunk).DataChunk.ChunkId)),
				chunk.Req.(*data_data.TransferReq_DataChunk).DataChunk.Content,
				os.ModeAppend,
			)

		case *data_data.TransferReq_FileName:
			fileName=chunk.Req.(*data_data.TransferReq_FileName).FileName
			fmt.Println("receiving file of name: "+fileName)
		}
	}
}