package database

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"
)

type Pokemon struct {
	ID     int
	Name   string
	Height int
	Weight int
	Types  []string
}

type Database struct {
	db *sql.DB
}

func NewDatabase(sqliteDB string) (*Database, error) {
	db, err := sql.Open("sqlite3", sqliteDB)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := createTable(db); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func createTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS pokemon (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		height INTEGER NOT NULL,
		weight INTEGER NOT NULL,
		types TEXT NOT NULL
	);`

	_, err := db.Exec(query)
	return err
}

func (d *Database) AddPokemon(pokemon Pokemon) (*Pokemon, error) {
	types, err := json.Marshal(pokemon.Types)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO pokemon (id, name, height, weight, types) VALUES (?, ?, ?, ?, ?)`
	_, err = d.db.Exec(query, pokemon.ID, pokemon.Name, pokemon.Height, pokemon.Weight, string(types))
	if err != nil {
		return nil, err
	}

	return &pokemon, nil
}

func (d *Database) GetPokemonByName(name string) (*Pokemon, error) {
	query := `SELECT id, name, height, weight, types FROM pokemon WHERE name = ?`
	row := d.db.QueryRow(query, name)

	var pokemon Pokemon
	var typesJSON string
	if err := row.Scan(&pokemon.ID, &pokemon.Name, &pokemon.Height, &pokemon.Weight, &typesJSON); err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(typesJSON), &pokemon.Types); err != nil {
		return nil, err
	}
	return &pokemon, nil
}

func (d *Database) GetPokemonByType(pokemonType string) ([]Pokemon, error) {
	query := `SELECT id, name, height, weight, types FROM pokemon`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pokemons []Pokemon
	for rows.Next() {
		var p Pokemon
		var typesJSON string
		if err := rows.Scan(&p.ID, &p.Name, &p.Height, &p.Weight, &typesJSON); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(typesJSON), &p.Types); err != nil {
			return nil, err
		}

		// filter by type
		for _, t := range p.Types {
			if t == pokemonType {
				pokemons = append(pokemons, p)
				break
			}
		}
	}
	return pokemons, nil
}

func (d *Database) GetAllPokemon() ([]Pokemon, error) {
	query := `SELECT id, name, height, weight, types FROM pokemon`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pokemons []Pokemon
	for rows.Next() {
		var p Pokemon
		var typesJSON string
		if err := rows.Scan(&p.ID, &p.Name, &p.Height, &p.Weight, &typesJSON); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(typesJSON), &p.Types); err != nil {
			return nil, err
		}
		pokemons = append(pokemons, p)
	}
	return pokemons, nil
}

func (d *Database) TotalPokemon() (int, error) {
	query := `SELECT COUNT(*) FROM pokemon`
	var count int
	err := d.db.QueryRow(query).Scan(&count)
	return count, err
}

func (d *Database) DeletePokemonById(id int) error {
	query := `DELETE FROM pokemon WHERE id = ?`
	result, err := d.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (d *Database) PokemonExists(name string) (bool, error) {
	query := `SELECT COUNT(*) FROM pokemon WHERE name = ?`
	var count int
	err := d.db.QueryRow(query, name).Scan(&count)
	return count > 0, err
}
