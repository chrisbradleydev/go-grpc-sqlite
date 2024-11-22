package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/chrisbradleydev/go-grpc-sqlite/internal/database"
	pb "github.com/chrisbradleydev/go-grpc-sqlite/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedPokemonServiceServer
	db *database.Database
}

func (s *server) AddPokemon(ctx context.Context, req *pb.AddPokemonRequest) (*pb.Pokemon, error) {
	var types []string
	for _, t := range req.Pokemon.Types {
		types = append(types, strings.ToLower(t))
	}
	pokemon, err := s.db.AddPokemon(database.Pokemon{
		ID:     int(req.Pokemon.Id),
		Name:   strings.ToLower(req.Pokemon.Name),
		Height: int(req.Pokemon.Height),
		Weight: int(req.Pokemon.Weight),
		Types:  types,
	})
	if err != nil {
		return nil, err
	}
	return &pb.Pokemon{
		Id:     int32(pokemon.ID),
		Name:   pokemon.Name,
		Height: int32(pokemon.Height),
		Weight: int32(pokemon.Weight),
		Types:  pokemon.Types,
	}, nil
}

func (s *server) GetPokemonByName(ctx context.Context, req *pb.PokemonNameRequest) (*pb.Pokemon, error) {
	p, err := s.db.GetPokemonByName(strings.ToLower(req.Name))
	if err != nil {
		return nil, err
	}
	return &pb.Pokemon{
		Id:     int32(p.ID),
		Name:   p.Name,
		Height: int32(p.Height),
		Weight: int32(p.Weight),
		Types:  p.Types,
	}, nil
}

func (s *server) GetPokemonByType(ctx context.Context, req *pb.PokemonTypeRequest) (*pb.PokemonList, error) {
	allPokemonByType, err := s.db.GetPokemonByType(strings.ToLower(req.Type))
	if err != nil {
		return nil, err
	}

	var pbPokemons []*pb.Pokemon
	for _, p := range allPokemonByType {
		pbPokemons = append(pbPokemons, &pb.Pokemon{
			Id:     int32(p.ID),
			Name:   p.Name,
			Height: int32(p.Height),
			Weight: int32(p.Weight),
			Types:  p.Types,
		})
	}

	return &pb.PokemonList{Pokemon: pbPokemons}, nil
}

func (s *server) GetAllPokemon(ctx context.Context, _ *pb.Empty) (*pb.PokemonList, error) {
	allPokemon, err := s.db.GetAllPokemon()
	if err != nil {
		return nil, err
	}

	var pbPokemons []*pb.Pokemon
	for _, p := range allPokemon {
		pbPokemons = append(pbPokemons, &pb.Pokemon{
			Id:     int32(p.ID),
			Name:   p.Name,
			Height: int32(p.Height),
			Weight: int32(p.Weight),
			Types:  p.Types,
		})
	}

	return &pb.PokemonList{Pokemon: pbPokemons}, nil
}

func (s *server) DeletePokemonById(ctx context.Context, req *pb.DeletePokemonRequest) (*pb.Empty, error) {
	err := s.db.DeletePokemonById(int(req.Id))
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (s *server) PokemonExists(ctx context.Context, req *pb.PokemonNameRequest) (*pb.PokemonExistsResponse, error) {
	_, err := s.db.GetPokemonByName(strings.ToLower(req.Name))
	exists := err == nil
	return &pb.PokemonExistsResponse{Exists: exists}, nil
}

type config struct {
	env      string
	grpcHost string
	grpcPort string
	sqliteDB string
}

func main() {
	var cfg config
	flag.StringVar(&cfg.env, "env", os.Getenv("APP_ENV"), "environment (development|production)")
	flag.StringVar(&cfg.grpcHost, "grpcHost", os.Getenv("GRPC_HOST"), "gRPC host")
	flag.StringVar(&cfg.grpcPort, "grpcPort", os.Getenv("GRPC_PORT"), "gRPC port")
	flag.StringVar(&cfg.sqliteDB, "sqliteDB", os.Getenv("SQLITE_DB"), "path to SQLite database")
	flag.Parse()

	// initialize database
	db, err := database.NewDatabase(cfg.sqliteDB)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	lis, err := net.Listen("tcp", ":"+cfg.grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPokemonServiceServer(s, &server{db: db})

	// enable reflection
	reflection.Register(s)

	// seed database in a goroutine
	go func() {
		// wait for server to start
		time.Sleep(2 * time.Second)

		// before seeding, check to see if database has all first generation Pokemon
		count, err := db.TotalPokemon()
		if err != nil {
			log.Printf("Error counting Pokemon: %v\n", err)
		}
		if count >= FirstGenPokemon {
			log.Printf("Database contains %d Pokemon, skipping seeding\n", count)
			return
		}

		if err := seedDatabase(cfg.grpcHost, cfg.grpcPort); err != nil {
			log.Printf("Error seeding database: %v\n", err)
		}
	}()

	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
