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
)

type Server struct {
	Probability float64
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
				return stream.SendAndClose(&OrderRes{
					ResCode: OrderResCode_Yes,
				})
			} else {
				fmt.Println("rejecting")
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
	file, err := os.OpenFile("namenode/log.txt", os.O_RDWR|os.O_CREATE/*|os.O_APPEND|*/, 0666)  
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	rand.Seed(time.Now().UnixNano())
	fmt.Println("receiving order request")
	var fileName string
	parte := 0
	for {
		ordReq, err := stream.Recv()
		if err == io.EOF {
			w.Flush()
			fmt.Print("sending response (inform order): \n")
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
			fmt.Print("received request: ")
			fmt.Print(ordReq.Req.(*OrderReq_OrderData).OrderData.ChunkId)
			fmt.Print(" in node: ")
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
			//lee primero una vez para obtener el título y el número de partes
			for i:=0; i<3; i++{
				scan.Scan()
				fmt.Println(scan.Text())
				muchotexto := scan.Text()
				fmt.Println(reflect.TypeOf(muchotexto))
				separados := strings.Split(muchotexto, " ")
				//separados[0]-> id; separados[1]-> ip
				fmt.Println(separados)
				id, errata := strconv.ParseInt(separados[0], 10, 64)
				if errata!=nil{
					log.Fatal(errata)
				}
				if id == ordReq.Req.(*OrderReq_OrderData).OrderData.NodeId{
					_, err := w.WriteString("parte_1_"+s+" "+separados[1] + "\n")
					if err != nil {
						log.Fatal(err)
					}
					f.Close()
					break
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
			fileName = ordReq.Req.(*OrderReq_FileName).FileName
			fmt.Println("receiving file of name: " + fileName)
			_, err1 := w.WriteString(fileName + "\n")
  			if err1 != nil {
    			log.Fatal(err1)
			}
		}
		parte = parte + 1
	}
}
