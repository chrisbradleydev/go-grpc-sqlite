package main

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	pb "github.com/chrisbradleydev/go-grpc-sqlite/protos"
	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	BaseURL         = "https://pokeapi.co/api/v2"
	FirstGenPokemon = 151
	RequestDelay    = 1 * time.Second
)

type PokemonResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Height int    `json:"height"`
	Weight int    `json:"weight"`
	Types  []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

func seedDatabase(grpcHost, grpcPort string) error {
	// connect to gRPC server
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%s", grpcHost, grpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewPokemonServiceClient(conn)
	restyClient := resty.New()
	restyClient.SetBaseURL(BaseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pokemonList, err := client.GetAllPokemon(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not list pokemon: %v", err)
		return err
	}

	existingPokemonIds := make([]int, FirstGenPokemon)
	for i, p := range pokemonList.Pokemon {
		// check if the pokemon exists
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		pbResp, err := client.PokemonExists(ctx, &pb.PokemonNameRequest{Name: p.Name})
		defer cancel()

		if err != nil {
			log.Printf("Error checking if Pokemon %s exists: %v\n", p.Name, err)
			continue
		}

		if pbResp.Exists {
			log.Printf("Pokemon %s (ID: %d) already exists\n", p.Name, p.Id)
			existingPokemonIds[i] = int(p.Id)
		}
	}

	for id := 1; id <= FirstGenPokemon; id++ {
		// skip if the pokemon exists
		if slices.Contains(existingPokemonIds, id) {
			continue
		}

		// add delay between requests
		time.Sleep(RequestDelay)

		var pokemonResp PokemonResponse
		resp, err := restyClient.R().
			SetResult(&pokemonResp).
			Get(fmt.Sprintf("/pokemon/%d", id))
		if err != nil {
			log.Printf("Error fetching Pokemon %d: %v\n", id, err)
			continue
		}

		if resp.IsError() {
			log.Printf("Error response for Pokemon %d: %v\n", id, resp.Status())
			continue
		}

		// extract types from the response
		types := make([]string, len(pokemonResp.Types))
		for i, t := range pokemonResp.Types {
			types[i] = strings.ToLower(t.Type.Name)
		}

		// create Pokemon protobuf message
		pokemon := &pb.Pokemon{
			Id:     int32(pokemonResp.ID),
			Name:   strings.ToLower(pokemonResp.Name),
			Height: int32(pokemonResp.Height),
			Weight: int32(pokemonResp.Weight),
			Types:  types,
		}

		// add Pokemon using gRPC client
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = client.AddPokemon(ctx, &pb.AddPokemonRequest{Pokemon: pokemon})
		defer cancel()

		if err != nil {
			log.Printf("Error adding Pokemon %s (ID: %d): %v\n", pokemon.Name, pokemon.Id, err)
			continue
		}

		log.Printf("Successfully added Pokemon: %s (ID: %d, Types: %v)\n",
			pokemon.Name,
			pokemon.Id,
			strings.Join(pokemon.Types, ", "),
		)
	}

	return nil
}
