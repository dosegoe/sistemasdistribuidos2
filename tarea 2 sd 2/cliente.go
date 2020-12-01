package main 

import(
	"io"
	//"fmt"
	//"math/rand"
	//"time"
	//"io/ioutil"
	//"strconv"
	//"os"
	"context"
	"log"
	client_name "./tareasd2/grpc/client_name/client_name" 
	grpc "google.golang.org/grpc"
)

func main(){
	ipaddr := "192.168.1.82"
	port := "9010"

	conn, err := grpc.Dial(ipaddr+":"+port, grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    client := client_name.NewClientNameClient(conn)

	req := client_name.OrderReq{ Filename: "El_maravilloso_Mago_de_Oz-L._Frank_Baum.pdf"}  // initialize a pb.Rectangle
	stream, err := client.ChunksOrder(context.Background(), &req)
	if err != nil {
		panic(err)
	}
	for {
		feature, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
		}
		log.Println(feature)
	}

}