package main

import (
    "context"
    "encoding/json"
    "log"
    "net"
    "google.golang.org/grpc"
    pb "path/to/generated/protobuf/package"
)

type server struct {
    pb.UnimplementedStreamServiceServer
}

func (s *server) StreamData(req *pb.RequestMessage, stream pb.StreamService_StreamDataServer) error {
    // Simulate sending JSON data
    for i := 0; i < 5; i++ {
        data := map[string]interface{}{
            "request_id": req.GetRequestId(),
            "event":      i,
        }
        jsonData, err := json.Marshal(data)
        if err != nil {
            return err
        }
        res := &pb.ResponseMessage{Data: string(jsonData)}
        if err := stream.Send(res); err != nil {
            return err
        }
        // Sleep to simulate long-running operation
        time.Sleep(time.Second)
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