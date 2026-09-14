package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"context"


	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func ContactHandler(w http.ResponseWriter, r *http.Request) {
	Name := r.FormValue("name")
	Email := r.FormValue("email")
	Message := r.FormValue("message")

	fmt.Println("Name:", Name)
	fmt.Println("Email:", Email)
	fmt.Println("Message:", Message)

}

func AiHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusInternalServerError)
		return
	}

	Question := r.FormValue("question")
	fmt.Println(Question)
}

func main() {

	http.HandleFunc("/ai", AiHandler)
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/contact", ContactHandler)
	fmt.Println("Server running on http://localhost:8082")

	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))

	http.Handle("/style.css", http.FileServer(http.Dir(".")))

	http.Handle("/script.js", http.FileServer(http.Dir(".")))

	err := http.ListenAndServe(":8082", nil)
	if err != nil {
		fmt.Println(err)
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY is missing")
	}

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatal(err)
	}
}
