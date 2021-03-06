package data_name

import (
	"fmt"
	"io"
	"math/rand"
	"time"
	"os"
	"log"
	"bufio"
	"strconv"
	"strings"
	"reflect"
	"sync"
)

type Server struct {
	Probability float64
	BookNum *int
	Messages *int
}
//copié el mismo código de requestorder en data_data
func (server *Server) RequestOrder(stream DataName_RequestOrderServer) error {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("receiving order request")
	var fileName string
	for {
		ordReq, err := stream.Recv()
		if err == io.EOF {
			fmt.Print("sending response (requiest order): ")
			//acá se simula la aceptación o rechazo del orden, según una "probabilidad de exito"
			r := rand.Float64()
			if r < server.Probability {
				fmt.Println("sipo approbing ")
				*(server.Messages)=*(server.Messages)+1
				return stream.SendAndClose(&OrderRes{
					ResCode: OrderResCode_Yes,
				})
			} else {
				fmt.Println("rejecting")
				*(server.Messages)=*(server.Messages)+1
				return stream.SendAndClose(&OrderRes{
					ResCode: OrderResCode_No,
				})
			}
			//server.printTotalChunks()

		}
		if err != nil {
			return nil
		}
		//fmt.Printf("type: %T\n",upreq.Data)
		switch ordReq.Req.(type) {
		case *OrderReq_OrderData:
			fmt.Print("received request: ")
			fmt.Print(ordReq.Req.(*OrderReq_OrderData).OrderData.ChunkId)
			fmt.Print(" in node: ")
			fmt.Println(ordReq.Req.(*OrderReq_OrderData).OrderData.NodeId)
		case *OrderReq_FileName:
			fileName = ordReq.Req.(*OrderReq_FileName).FileName
			fmt.Println("receiving file of name: " + fileName)
		}
	}

	return nil
}

//similar a lo RequestOrder, pero en lugar

func (Server *Server) InformOrder(stream DataName_InformOrderServer) error {
	fmt.Println("Inform Order")
	file, err := os.OpenFile("namenode/log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)  
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Receiving Order Request (Inform Order)")
	var fileName string
	parte := 0
	*(Server.BookNum)=*(Server.BookNum)+1
	var modo string
	var candado sync.Mutex
	for {
		//Recibe archivo desde chunk transfer filename o chunk_id+node_id
		ordReq, err := stream.Recv()

		//Se acabo el stream
		if err == io.EOF {
			w.Flush()
			fmt.Print("Sending Response (inform order): \n")
			*(Server.Messages)=*(Server.Messages)+1
			return stream.SendAndClose(&OrderRes{
				ResCode: OrderResCode_Yes,
			})
		}
		if err != nil {
			return nil
		}
		//fmt.Printf("type: %T\n",upreq.Data)
		switch ordReq.Req.(type) {
		case *OrderReq_OrderData:
			if modo == "dist"{
				fmt.Print("Received Request (ChunkID): ")
				fmt.Print(ordReq.Req.(*OrderReq_OrderData).OrderData.ChunkId)
				fmt.Print(" In Node: ")
				fmt.Println(ordReq.Req.(*OrderReq_OrderData).OrderData.NodeId)
				/*
					TODO: ESCRIBIR EN EL LOG



				*/
				s := strconv.Itoa(parte)
				f, err5 := os.Open("namenode/Dnodes.txt")
				if err5 != nil{
					log.Fatal(err5)
				}
				scan := bufio.NewScanner(f)
				
				bookN:=strconv.Itoa(*(Server.BookNum))
				fmt.Println("book id: "+bookN)
				//lee primero una vez para obtener el título y el número de partes
				for i:=0; i<3; i++{
					scan.Scan()
					fmt.Println(scan.Text())
					muchotexto := scan.Text()
					fmt.Println(reflect.TypeOf(muchotexto))
					separados := strings.Split(muchotexto, " ")
					//separados[0]-> id; separados[1]-> ip
					fmt.Println(separados)
					fmt.Printf("separados[0] es %s", separados[0])
					id, errata := strconv.ParseInt(separados[0], 10, 64)
					if errata!=nil{
						log.Fatal(errata)
					}
					if id == ordReq.Req.(*OrderReq_OrderData).OrderData.NodeId{
						_, err := w.WriteString("parte_"+bookN+"_"+s+" "+separados[1] + "\n")
						if err != nil {
							log.Fatal(err)
						}
						f.Close()
						break
					}
				}
			}else{
				fmt.Print("ORDER DATA CENTRALIZADO: ")
				fmt.Print("Received Request (ChunkID): ")
				fmt.Print(ordReq.Req.(*OrderReq_OrderData).OrderData.ChunkId)
				fmt.Print(" In Node: ")
				fmt.Println(ordReq.Req.(*OrderReq_OrderData).OrderData.NodeId)
				/*
					TODO: ESCRIBIR EN EL LOG
				*/
				s := strconv.Itoa(parte)
				candado.Lock()
				f, err5 := os.Open("namenode/Dnodes.txt")
				if err5 != nil{
					log.Fatal(err5)
				}
				scan := bufio.NewScanner(f)
				
				bookN:=strconv.Itoa(*(Server.BookNum))
				fmt.Println("book id: "+bookN)
				//lee primero una vez para obtener el título y el número de partes
				for i:=0; i<3; i++{
					scan.Scan()
					fmt.Println(scan.Text())
					muchotexto := scan.Text()
					fmt.Println(reflect.TypeOf(muchotexto))
					separados := strings.Split(muchotexto, " ")
					//separados[0]-> id; separados[1]-> ip
					fmt.Println(separados)
					fmt.Printf("separados[0] es %s", separados[0])
					id, errata := strconv.ParseInt(separados[0], 10, 64)
					if errata!=nil{
						log.Fatal(errata)
					}
					if id == ordReq.Req.(*OrderReq_OrderData).OrderData.NodeId{
						_, err := w.WriteString("parte_"+bookN+"_"+s+" "+separados[1] + "\n")
						if err != nil {
							log.Fatal(err)
						}
						f.Close()
						candado.Unlock()
						fmt.Println("Finalizó escritura en Inform Order Centralizado")
						break
					}
				}
			}
			/*
			fmt.Println("parte_1_"+s+" "+strconv.FormatInt(ordReq.Req.(*OrderReq_OrderData).OrderData.NodeId,10) + "\n")
			_, err := w.WriteString("parte_1_"+s+" "+strconv.FormatInt(ordReq.Req.(*OrderReq_OrderData).OrderData.NodeId,10) + "\n")
  			if err != nil {
    			log.Fatal(err)
			  }
			  */
		case *OrderReq_FileName:
			fmt.Println(ordReq.Req.(*OrderReq_FileName).FileName)
			fn := ordReq.Req.(*OrderReq_FileName).FileName
			modo_fn := strings.Split(fn, "+")
			modo = modo_fn[0]
			fileName = modo_fn[1]
			fmt.Printf("El modo es %s", modo)
			fmt.Printf("El fileName es %s", fileName)
			fmt.Println("Receiving file of name in inform order: " + fileName)
			_, err1 := w.WriteString(fileName + "\n")
  			if err1 != nil {
    			log.Fatal(err1)
			}
		}
		parte = parte + 1
	}
}
