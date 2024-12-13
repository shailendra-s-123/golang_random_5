package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    pb "path/to/your/protobuf/generated/package" // Update this path
)

func main() {
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("did not connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewStreamServiceClient(conn)

    req := &pb.RequestMessage{RequestId: "request-123"}

    stream, err := client.StreamData(context.Background(), req)
    if err != nil {
        log.Fatalf("error while calling StreamData: %v", err)
    }

    for {
        res, err := stream.Recv()
        if err != nil {
            log.Fatalf("error while receiving: %v", err)
        }
        log.Printf("Received JSON data: %s", res.Data)
    }
}