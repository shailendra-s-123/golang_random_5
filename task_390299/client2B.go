package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	pb "path/to/generated/protobuf/package" // Adjust import path
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewStreamServiceClient(conn)
	req := &pb.Request{ClientId: "client123"}

	stream, err := c.StreamData(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling StreamData: %v", err)
	}

	for {
		res, err := stream.Recv()
		if err != nil {
			if err == context.Canceled || err == io.EOF {
				break
			}
			log.Fatalf("error while receiving: %v", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(res.Data), &data); err != nil {
			log.Error(err)
			continue
		}
		fmt.Println("Received Data:", data)
	}
}