package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/NikoMalik/blockchain/proto"
	"github.com/NikoMalik/blockchain/server"
	"google.golang.org/grpc"
)

func main() {
	addr := [6]byte{127, 0, 0, 1, 0x1F, 0x90}

	opts := [2]grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024 << 1),
		grpc.MaxSendMsgSize(1024 << 1),
	}

	s := server.NewServer(&addr, opts[:]...)
	var wg = &sync.WaitGroup{}
	go func() {
		for i := 0; i < 10; i++ {
			wg.Add(1)

			time.Sleep(2 * time.Second)

			makeTestTransaction()
			wg.Done()
		}
	}()
	wg.Wait()

	// Start the server in a separate goroutine so the main thread can wait for the interrupt signal
	func() {
		s.LoggerReturn().Info("started")
		if err := s.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Block here and wait for a termination signal to stop the server gracefully
	func() {
		if err := s.Stop(); err != nil {
			panic(err)
		}

	}()
}

func makeTestTransaction() {
	conn, err := grpc.Dial("127.0.0.1:8080", grpc.WithInsecure())
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err != nil {
		panic(err)
	}
	version := &proto.Version{
		Version: 0x01,
		Height:  100,
	}

	tx := &proto.Transaction{
		Version: 0x01,
	}
	client := proto.NewServerClient(conn)
	_, err = client.HandShake(ctx, version)
	if err != nil {
		panic(err)
	}
	client.HandleTransaction(ctx, tx)

}
