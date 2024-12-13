package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"time"

	pb "example"
	"google.golang.org/grpc"
)

type dataStreamServer struct {
	pb.UnimplementedDataStreamServiceServer
}

func (s *dataStreamServer) StreamData(ctx context.Context, req *pb.StreamRequest) (*pb.StreamResponse, error) {
	stream, err := pb.NewDataStreamServiceServer(ctx, s)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var data []byte
		switch req.DataType {
		case "json":
			data, err = json.Marshal(map[string]string{"key": "value", "time": time.Now().Format(time.RFC3339)})
		case "xml":
			var xmlData struct {
				Key   string `xml:"key"`
				Time  string `xml:"time"`
			}
			xmlData.Key = "value"
			xmlData.Time = time.Now().Format(time.RFC3339)
			data, err = xml.Marshal(&xmlData)
		default:
			err = fmt.Errorf("unsupported data type: %s", req.DataType)
		}

		if err != nil {
			return nil, err
		}

		if err := stream.Send(&pb.StreamResponse{Data: data}); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterDataStreamServiceServer(s, &dataStreamServer{})

	log.Println("server starting on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}