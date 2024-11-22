package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/chrisbradleydev/go-grpc-sqlite/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type config struct {
	env      string
	grpcHost string
	grpcPort string
}

func displayPokemon(pokemon *pb.Pokemon) {
	fmt.Printf("ID: %d\n", pokemon.Id)
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Printf("Types: %s\n", strings.Join(pokemon.Types, ", "))
	fmt.Println(strings.Repeat("-", 20))
}

func main() {
	var cfg config
	flag.StringVar(&cfg.env, "env", os.Getenv("APP_ENV"), "environment (development|production)")
	flag.StringVar(&cfg.grpcHost, "grpcHost", os.Getenv("GRPC_HOST"), "gRPC host")
	flag.StringVar(&cfg.grpcPort, "grpcPort", os.Getenv("GRPC_PORT"), "gRPC port")
	flag.Parse()

	// connect to gRPC server
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%s", cfg.grpcHost, cfg.grpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewPokemonServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// output all Pokemon in database
	resp, err := client.GetAllPokemon(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not list pokemon: %v", err)
		return
	}

	if len(resp.Pokemon) == 0 {
		log.Fatalf("no pokemon found")
		return
	}

	fmt.Println(strings.Repeat("-", 20))
	for _, pokemon := range resp.Pokemon {
		displayPokemon(pokemon)
	}
}
