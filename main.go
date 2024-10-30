package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

type Recipe struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Ingredients  []string `json:"ingredients"`
	Instructions string   `json:"instructions"`
}

func main() {
	// Connects to database, makes sure no error opening db
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS recipes (
			id SERIAL PRIMARY KEY,
			name TEXT,
			ingredients TEXT[],
			instructions TEXT
		);
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create router
	router := mux.NewRouter()
	// Routes
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		http.ServeFile(w, r, "./public/index.html")
	})
	router.HandleFunc("/recipes", getRecipes(db)).Methods("GET")
	router.HandleFunc("/recipes/{id}", getRecipe(db)).Methods("GET")
	router.HandleFunc("/recipes", createRecipe(db)).Methods("POST")
	router.HandleFunc("/recipes/{id}", updateRecipe(db)).Methods("PUT")
	router.HandleFunc("/recipes/{id}", deleteRecipe(db)).Methods("DELETE")
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./public/"))))
	log.Fatal(http.ListenAndServe(":8000", jsonContentTypeMiddleware(router)))
}

// Middleware function which sets header content type to json
func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func getRecipes(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM recipes")
		if err != nil {
			log.Println("Error querying recipes:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		recipes := []Recipe{}
		for rows.Next() {
			var r Recipe
			if err := rows.Scan(&r.ID, &r.Name, pq.Array(&r.Ingredients), &r.Instructions); err != nil {
				log.Println("Error scanning row:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			recipes = append(recipes, r)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(recipes); err != nil {
			log.Println("Error encoding response:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func getRecipe(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		var rec Recipe
		err := db.QueryRow("SELECT * FROM recipes WHERE id = $1", id).Scan(&rec.ID, &rec.Name, pq.Array(&rec.Ingredients), &rec.Instructions)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(rec)
	}
}

func createRecipe(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rec Recipe
		err := json.NewDecoder(r.Body).Decode(&rec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Println("Decoded recipe:", rec)
		err = db.QueryRow("INSERT INTO recipes (name, ingredients, instructions) VALUES ($1, $2, $3) RETURNING id", rec.Name, pq.Array(rec.Ingredients), rec.Instructions).Scan(&rec.ID)
		if err != nil {
			log.Println("Error inserting recipe:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(rec); err != nil {
			log.Println("Error encoding response:", err)
		}
	}
}

// Test/fix later

func updateRecipe(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rec Recipe
		json.NewDecoder(r.Body).Decode(&r)
		vars := mux.Vars(r)
		id := vars["id"]
		_, err := db.Exec("UPDATE recipes SET name = $1, ingredients = $2, instructions = $3 WHERE id = $4", rec.Name, rec.Ingredients, rec.Instructions, id)
		if err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode(rec)
	}
}

func deleteRecipe(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		_, err := db.Exec("DELETE FROM recipes WHERE id = $1", id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode("Recipe deleted")
	}
}
