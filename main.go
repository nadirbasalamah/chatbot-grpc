package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/nadirbasalamah/chatbot-grpc/chat/chatpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// membuat server untuk mengimplementasikan fungsi rpc
type server struct {
}

// membuat fungsi untuk membuat pesan untuk server
// yang akan dikirimkan
func sendMessage(msg string) string {
	if strings.Contains("hello", msg) {
		return "Hi, nice to meet you :)"
	} else if strings.Contains("who are you", msg) {
		return " I am simple chatbot"
	} else if strings.Contains("good bye", msg) {
		return "Good bye!"
	} else {
		return "Sorry, i dont know what do you mean"
	}
}

// implementasi fungsi rpc Chat
func (*server) Chat(stream chatpb.ChatBot_ChatServer) error {
	// membuat variabel
	// untuk mengirim pesan ke client
	var message string
	for {
		// menerima request
		req, err := stream.Recv()

		// jika tidak ada request
		// hentikan eksekusi
		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Fatalf("Error while reading stream: %v", err)
			return err
		}

		// menerima pesan client dari request
		clientMessage := req.GetMessage()

		// mengirim pesan ke client
		message = clientMessage
		sendErr := stream.Send(&chatpb.ChatResponse{
			Message: sendMessage(message),
		})
		// jika terdapat error saat mengirim pesan
		// tampilkan error
		if sendErr != nil {
			log.Fatalf("Error while sending result: %v", sendErr)
			return sendErr
		}
	}
}

func main() {
	// jika kode mengalami crash, nomor line akan ditampilkan
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Chatbot service started")

	// membuat gRPC server
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v\n", err)
	}

	s := grpc.NewServer()
	// melakukan register ChatBotServer
	chatpb.RegisterChatBotServer(s, &server{})
	// mengaktifkan reflection
	// agar bisa digunakan untuk pengujian dengan evans
	reflection.Register(s)

	go func() {
		fmt.Println("Starting server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Menunggu hingga dihentikan dengan Ctrl + C
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Lakukan block hingga sinyal sudah didapatkan
	<-ch
	fmt.Println("Stopping the server..")
	s.Stop()
	fmt.Println("Stopping listener...")
	lis.Close()
	fmt.Println("End of Program")
}
