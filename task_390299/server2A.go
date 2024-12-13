package main

import (
    "context"
    "encoding/json"
    "log"
    "net"
    "time"

    "google.golang.org/grpc"
    pb "path/to/your/protobuf/generated/package" // Update this path
)

// Server struct implements the gRPC interface
type server struct {
    pb.UnimplementedStreamServiceServer
}

// StreamData method sends continuous data to the client
func (s *server) StreamData(req *pb.RequestMessage, stream pb.StreamService_StreamDataServer) error {
    for i := 0; i < 5; i++ { // Simulate 5 messages
        jsonData, err := json.Marshal(map[string]interface{}{
            "request_id": req.GetRequestId(),
            "event":      i,
        })
        if err != nil {
            return err
        }

        res := &pb.ResponseMessage{Data: string(jsonData)}
        if err := stream.Send(res); err != nil {
            return err
        }
        time.Sleep(1 * time.Second) // Simulate some processing delay
    }
    return nil
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    pb.RegisterStreamServiceServer(grpcServer, &server{})
    log.Println("Server is running on port :50051")
    
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}