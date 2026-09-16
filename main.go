package main

import (
	"context"
	"fmt"
	"html/template"

	// "internal/runtime/gc/scan"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/genai"
)

var db *sql.DB

var geminiClient *genai.Client
var ctx = context.Background()

// const churchSystemInstruction = `You are the official AI assistant for Ebenezer Christ Church of the Lord (Aladura).
// Help visitors learn about the church, its history, founder, services, ministries, branches, and general information.
// Be respectful, welcoming, clear, and concise. Only provide information that is supplied as official church information.
// Never invent church facts. If you do not know an answer, say that the information is not currently available and suggest contacting the church.

// Official Church Data:
// - Name: Ebenezer Christ Church of the Lord (Aladura)
// - Motto: “One House One Unit”
// - Location: Lagos State, Nigeria
// - Headquarters: No. 21 Olowo Street
// - Branches: About 7 branches (use placeholder branch names if asked)
// - Service days: Monday, Tuesday, Wednesday, Friday, and Sunday
// - Ministries/Departments: Youth, Children, Choir, Prayer, Men's Fellowship, Women's Fellowship, Bible Study, and Evangelism
// - Founder and First Primate: His Eminence Primate (Dr.) Erastus Adeshina Oduneye (J.P.)`

const churchSystemInstruction = `You are the official AI assistant for Ebenezer Christ Church of the Lord (Aladura).

Your ONLY purpose is to help visitors with questions related to the church.

You may answer questions about:
- The church
- The church's history
- The founder and primates
- Church services
- Service days
- Ministries and departments
- Branches and locations
- Church events
- Church activities
- The church motto
- General information about the church

IMPORTANT RULES:

1. If the visitor asks a church-related question, answer helpfully, respectfully, and clearly.

2. Only use the official church information provided below. Never invent church facts, dates, locations, people, events, or other information.

3. If the requested church information is not available in the official information below, say:
"I don't currently have that information. Please contact the church directly for more details."

4. If the visitor asks a question that is NOT related to the church, do NOT answer the unrelated question.

Instead, politely say:
"I'm here to help with information about Ebenezer Christ Church of the Lord (Aladura), including our history, services, ministries, branches and church activities. Please ask me a church-related question."

5. Do not act as a general-purpose AI assistant.

6. Be welcoming and respectful to every visitor.

Language Rule:

- Detect the language used by the visitor.
- Respond in the same language whenever possible.
- The assistant may respond in English, Yoruba, Igbo, Hausa, or other languages it understands.
- Do not translate the official church facts into a different language unless the visitor asks in that language.
- Keep names, official church names, addresses, and other proper names accurate.

Official Church Data:

- Name: Ebenezer Christ Church of the Lord (Aladura)
- Motto: "One House One Unit"
- Location: Lagos State, Nigeria
- Headquarters: No. 21 Olowo Street
- Branches: About 7 branches

Official Church Data:

- Name: Ebenezer Christ Church of the Lord (Aladura)
- Motto: "One House One Unit"
- Location: Lagos State, Nigeria
- Headquarters: No. 21 Olowo Street

Branch Locations:

- Headquarters:
  No. 21 Olowo Street, Mushin, Lagos State

- Ketu Branch:
  15 Ogunseinde Street, Ketu, Lagos State

- Ikorodu Branch:
  6 Chris Aghanenu Close, Eyita, Ikorodu, Lagos State

- Ikala – Ijebu Branch:
  Ikala – Ijebu, Ijebu Imusin, Ijebu East LGA, Ogun State

- Baale Ajuwon Branch:
  54/56 Fashina Street, Baale Ajuwon, Ogun State

- Sango Ota Branch:
  1 Queens Avenue, Sango Ota, Ogun State

- Service days:
  Monday, Tuesday, Wednesday, Friday, and Sunday

- Ministries/Departments:
  Youth, Children, Choir, Prayer, Men's Fellowship,
  Women's Fellowship, Bible Study, and Evangelism

- Founder and First Primate:
  His Eminence Primate (Dr.) Erastus Adeshina Oduneye (J.P.)`

type DashboardData struct {
	TotalMessages  int
	UnreadMessages int
	RecentMessages []Message
}

type Message struct {
	ID        int
	Name      string
	Email     string
	Phone     string
	Message   string
	Status    string
	CreatedAt string
}

type MessagesData struct {
	Messages []Message
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func ContactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		Name := r.FormValue("name")
		Email := r.FormValue("email")
		Message := r.FormValue("message")
		Phone := r.FormValue("phone")

		query := "INSERT INTO messages(name , email ,phone, message,status, create_at) VALUES(?,?,?,?,?,?)"
		_, err := db.Exec(query, Name, Email, Phone, Message, "unread", time.Now())
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func AiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	question := r.FormValue("question")
	if question == "" {
		http.Error(w, "Question is required", http.StatusBadRequest)
		return
	}

	// Generate content using Gemini API (SDK v1.71.0)
	result, err := geminiClient.Models.GenerateContent(
		ctx,
		"gemini-3.6-flash",
		genai.Text(question),
		&genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(
				churchSystemInstruction,
				genai.RoleUser,
			),
		},
	)

	if err != nil {
		log.Printf("Gemini error: %v", err)

		errorMessage := "Sorry, the church AI assistant is temporarily unavailable. Please try again later."

		if strings.Contains(err.Error(), "429") ||
			strings.Contains(err.Error(), "quota") ||
			strings.Contains(err.Error(), "RESOURCE_EXHAUSTED") {

			errorMessage = "Our church AI assistant has temporarily reached its usage limit. Please try again later or contact the church directly."
		}

		http.Error(w, errorMessage, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result.Text()))
}

func createAdmin() {
	username := "admin"
	password := "oluwatosin21@777"

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		log.Fatal(err)
	}

	query := "INSERT INTO admins(username, password, created_at) VALUES(?,?,?)"

	_, err = db.Exec(query, username, string(hashedPassword), time.Now())

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Admin created successfully!")
}

func AdminLoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "admin-login.html")
		return
	}

	if r.Method == http.MethodPost {

		username := r.FormValue("username")
		password := r.FormValue("password")

		var hashedPassword string

		query := "SELECT password FROM admins WHERE username = ?"

		err := db.QueryRow(query, username).Scan(&hashedPassword)

		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword(
			[]byte(hashedPassword),
			[]byte(password),
		)

		if err != nil {
			fmt.Println("BCRYPT ERROR:", err)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

}

func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	var totalMessages int
	var unreadMessages int

	err := db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&totalMessages)
	if err != nil {
		http.Error(w, "Failed to get message count", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM messages WHERE STATUS = 'unread'").Scan(&unreadMessages)
	if err != nil {
		http.Error(w, "Failed to get unread message count", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT id, name, email, phone, message, status, create_at FROM messages ORDER BY id DESC LIMIT 5")
	if err != nil {
		http.Error(w, "Failed to get recent messages", http.StatusInternalServerError)
		return
	}

	defer rows.Close() // Crucial: Prevents database connection leaks

	var recentMessages []Message
	for rows.Next() {
		var msg Message
		// Scan the columns into the Message struct fields
		if err := rows.Scan(&msg.ID, &msg.Name, &msg.Email, &msg.Phone, &msg.Message, &msg.Status, &msg.CreatedAt); err != nil {
			http.Error(w, "Failed to parse recent messages", http.StatusInternalServerError)
			return
		}
		recentMessages = append(recentMessages, msg)
	}

	// Check for errors from iterating rows
	if err = rows.Err(); err != nil {
		http.Error(w, "Error reading message rows", http.StatusInternalServerError)
		return
	}

	data := DashboardData{
		TotalMessages:  totalMessages,
		UnreadMessages: unreadMessages,
		RecentMessages: recentMessages,
	}

	templ, err := template.ParseFiles("admin-dashboard.html")
	if err != nil {
		http.Error(w, "Failed to load dashboard template", http.StatusInternalServerError)
		return
	}
	templ.Execute(w, data)
}

func AdminMessageHandler(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query("SELECT * FROM messages")
	if err != nil {
		log.Println("Database query error:", err)
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []Message

	for rows.Next() {
		var message Message

		err := rows.Scan(&message.ID, &message.Name, &message.Email, &message.Phone, &message.Message, &message.Status, &message.CreatedAt)

		if err != nil {
			http.Error(w, "Fail To Get", http.StatusInternalServerError)
			return
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Error reading messages", http.StatusInternalServerError)
		return
	}

	// log.Println("Messages loaded:", len(messages))

	// for _, message := range messages {
	// 	log.Println("Message:", message.ID, message.Name, message.Message)
	// }
	data := MessagesData{
		Messages: messages,
	}

	templ, err := template.ParseFiles("admin-messages.html")
	if err != nil {
		log.Println("Template error:", err)
		http.Error(w, "Failed to load messages template", http.StatusInternalServerError)
		return
	}
	// log.Println("Rendering admin-messages.html")
	templ.Execute(w, data)
}

func MarkMessageReadHandler(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue("id")

	_, err := db.Exec("UPDATE messages SET Status = ? WHERE id = ?", "read", id)
	if err != nil {
		http.Error(w, "Failed to mark message as read", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/messages", http.StatusSeeOther)
}

func DeleteMessageHandler(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue("id")
	_, err := db.Exec("DELETE FROM messages WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Failed to Delete", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/messages", http.StatusSeeOther)
}

func main() {
	// 1. Load environment variables first
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY is missing")
	}

	// 2. Initialize Gemini client prior to launching the server
	geminiClient, err = genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	// 3. Connect to MySQL database

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := dbUser + ":" + dbPassword + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Database connected successfully!")
	// createAdmin()

	// 3. Register routes and static file handlers
	http.HandleFunc("/admin", AdminDashboardHandler)
	http.HandleFunc("/ai", AiHandler)
	http.HandleFunc("/admin/login", AdminLoginHandler)
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/contact", ContactHandler)
	http.HandleFunc("/admin/messages", AdminMessageHandler)

	http.HandleFunc("/admin/messages/delete", DeleteMessageHandler)
	http.HandleFunc("/admin/messages/read", MarkMessageReadHandler)

	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	http.Handle("/style.css", http.FileServer(http.Dir(".")))
	http.Handle("/script.js", http.FileServer(http.Dir(".")))

	fmt.Println("Server running on http://localhost:8082")

	// 4. Start HTTP listener (blocking call)
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatal(err)
	}

}
