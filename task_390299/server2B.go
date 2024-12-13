package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	pb "path/to/generated/protobuf/package" // Adjust import path
)

type server struct {
	pb.UnimplementedStreamServiceServer
}

func (s *server) StreamData(ctx context.Context, req *pb.Request) (*pb.ResponseStream, error) {
	stream, err := pb.NewStreamServiceServer(ctx, s)
	if err != nil {
		return nil, err
	}

	go func() {
		for i := 0; ; i++ {
			data := map[string]interface{}{
				"client_id": req.ClientId,
				"timestamp": time.Now().Format(time.RFC3339),
				"value":     i,
			}
			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Error(err)
				return
			}
			if err := stream.Send(&pb.Response{Data: string(jsonData)}); err != nil {
				log.Error(err)
				return
			}
			time.Sleep(2 * time.Second)
		}
	}()

	return stream, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterStreamServiceServer(s, &server{})
	log.Println("Server is running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}