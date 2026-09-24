package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"path/filepath"

	// "internal/runtime/gc/scan"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
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
6. If a visitor wants to contact the church, speak with a church representative, send a prayer request, or ask for help from a person, guide them to the Contact Us section of the website.
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
	TotalMessages    int
	UnreadMessages   int
	TotalBranches    int
	TotalEvents      int
	RecentMessages   []Message
	RecentActivities []AdminActivity
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

type Branch struct {
	ID          int
	Name        string
	Address     string
	Pastor      string
	Phone       string
	CreatedAt   string
	Image       string
	PastorImage string
}

type Leader struct {
	ID        int
	Name      string
	Position  string
	Branch    string
	Image     string
	CreatedAt string
}

type Event struct {
	ID          int
	Title       string
	Description string
	EventDate   string
	EventTime   string
	Location    string
	CreatedAt   string
}

type AdminSettingsData struct {
	Username string
}

type AdminActivity struct {
	ID        int
	Action    string
	CreatedAt string
}

func GetLeaders() ([]Leader, error) {

	rows, err := db.Query(`
        SELECT id, name, position, branch, image, created_at
        FROM leaders
        ORDER BY id ASC
    `)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var leaders []Leader

	for rows.Next() {

		var leader Leader

		err := rows.Scan(
			&leader.ID,
			&leader.Name,
			&leader.Position,
			&leader.Branch,
			&leader.Image,
			&leader.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		leaders = append(leaders, leader)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return leaders, nil
}

func AdminLeadershipHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	leaders, err := GetLeaders()

	if err != nil {
		http.Error(w, "Failed to load leaders", http.StatusInternalServerError)
		return
	}

	data := struct {
		Leaders []Leader
	}{
		Leaders: leaders,
	}

	templ, err := template.ParseFiles("admin-leadership.html")

	if err != nil {
		fmt.Println("PARSE ERROR:", err)
		http.Error(w, "Failed to load add leader page", http.StatusInternalServerError)
		return
	}

	err = templ.Execute(w, data)

	if err != nil {
		fmt.Println("TEMPLATE EXECUTE ERROR:", err)
		return
	}
}

func AddLeaderHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	// Show the Add Leader form
	if r.Method == http.MethodGet {

		branches, err := GetBranches()

		if err != nil {
			http.Error(w, "Failed to load branches", http.StatusInternalServerError)
			return
		}

		templ, err := template.ParseFiles("admin-leadership-add.html")

		if err != nil {
			http.Error(w, "Failed to load add leader page", http.StatusInternalServerError)
			return
		}

		data := struct {
			Branches []Branch
		}{
			Branches: branches,
		}

		err = templ.Execute(w, data)

		if err != nil {
			fmt.Println("TEMPLATE ERROR:", err)
			http.Error(w, "Failed to load add leader page", http.StatusInternalServerError)
			return
		}

		return
	}

	// Process the form
	if r.Method == http.MethodPost {

		err := r.ParseMultipartForm(10 << 20)

		if err != nil {
			http.Error(w, "Failed to process form", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		position := r.FormValue("position")
		branch := r.FormValue("branch")

		var imageName string

		file, header, err := r.FormFile("image")

		if err == nil {

			defer file.Close()

			imageName = filepath.Base(header.Filename)

			imagePath := filepath.Join("images", imageName)

			destination, err := os.Create(imagePath)

			if err != nil {
				http.Error(w, "Failed to save image", http.StatusInternalServerError)
				return
			}

			defer destination.Close()

			_, err = io.Copy(destination, file)

			if err != nil {
				http.Error(w, "Failed to save image", http.StatusInternalServerError)
				return
			}
		}

		_, err = db.Exec(`
            INSERT INTO leaders
            (name, position, branch, image)
            VALUES (?, ?, ?, ?)
        `, name, position, branch, imageName)

		if err != nil {
			http.Error(w, "Failed to add leader", http.StatusInternalServerError)
			return
		}

		LogAdminActivity("Added a new church leader")

		http.Redirect(w, r, "/admin/leadership", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func EditLeaderHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/admin/leadership/edit/")

	if id == "" {
		http.Error(w, "Leader ID is missing", http.StatusBadRequest)
		return
	}

	// GET — show edit form
	if r.Method == http.MethodGet {

		var leader Leader

		err := db.QueryRow(`
            SELECT id, name, position, branch, image, created_at
            FROM leaders
            WHERE id = ?
        `, id).Scan(
			&leader.ID,
			&leader.Name,
			&leader.Position,
			&leader.Branch,
			&leader.Image,
			&leader.CreatedAt,
		)

		if err != nil {
			http.Error(w, "Leader not found", http.StatusNotFound)
			return
		}

		templ, err := template.ParseFiles("admin-leadership-edit.html")

		if err != nil {
			http.Error(w, "Failed to load edit page", http.StatusInternalServerError)
			return
		}

		err = templ.Execute(w, leader)

		if err != nil {
			http.Error(w, "Failed to display edit page", http.StatusInternalServerError)
			return
		}

		return
	}

	// POST — save changes
	if r.Method == http.MethodPost {

		err := r.ParseMultipartForm(10 << 20)

		if err != nil {
			http.Error(w, "Failed to process form", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		position := r.FormValue("position")
		branch := r.FormValue("branch")

		var oldImage string

		err = db.QueryRow(
			"SELECT image FROM leaders WHERE id = ?",
			id,
		).Scan(&oldImage)

		if err != nil {
			http.Error(w, "Leader not found", http.StatusNotFound)
			return
		}

		imageName := oldImage

		file, header, err := r.FormFile("image")

		if err == nil {

			defer file.Close()

			imageName = filepath.Base(header.Filename)

			imagePath := filepath.Join("images", imageName)

			destination, err := os.Create(imagePath)

			if err != nil {
				http.Error(w, "Failed to save image", http.StatusInternalServerError)
				return
			}

			defer destination.Close()

			_, err = io.Copy(destination, file)

			if err != nil {
				http.Error(w, "Failed to save image", http.StatusInternalServerError)
				return
			}
		}

		_, err = db.Exec(`
            UPDATE leaders
            SET name = ?, position = ?, branch = ?, image = ?
            WHERE id = ?
        `, name, position, branch, imageName, id)

		if err != nil {
			http.Error(w, "Failed to update leader", http.StatusInternalServerError)
			return
		}

		LogAdminActivity("Updated a church leader")

		http.Redirect(w, r, "/admin/leadership", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func DeleteLeaderHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/admin/leadership/delete/")

	if id == "" {
		http.Error(w, "Leader ID is missing", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(
		"DELETE FROM leaders WHERE id = ?",
		id,
	)

	if err != nil {
		http.Error(w, "Failed to delete leader", http.StatusInternalServerError)
		return
	}

	LogAdminActivity("Deleted a church leader")

	http.Redirect(w, r, "/admin/leadership", http.StatusSeeOther)
}

func AdminActivityHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	rows, err := db.Query(
		"SELECT id, action, created_at FROM admin_activity_logs ORDER BY id DESC",
	)

	if err != nil {
		http.Error(w, "Failed to load activity logs", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var activities []AdminActivity

	for rows.Next() {

		var activity AdminActivity

		err := rows.Scan(
			&activity.ID,
			&activity.Action,
			&activity.CreatedAt,
		)

		if err != nil {
			http.Error(w, "Failed to read activity logs", http.StatusInternalServerError)
			return
		}

		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error reading activity logs", http.StatusInternalServerError)
		return
	}

	data := struct {
		Activities []AdminActivity
	}{
		Activities: activities,
	}

	templ, err := template.ParseFiles("admin-activity.html")

	if err != nil {
		http.Error(w, "Failed to load activity page", http.StatusInternalServerError)
		return
	}

	err = templ.Execute(w, data)

	if err != nil {
		http.Error(w, "Failed to display activity page", http.StatusInternalServerError)
		return
	}
}

var adminSessions = make(map[string]bool)
var sessionMutex sync.RWMutex

func IsAdminLoggedIn(r *http.Request) bool {

	cookie, err := r.Cookie("admin_session")

	if err != nil {
		return false
	}

	sessionMutex.RLock()
	defer sessionMutex.RUnlock()

	return adminSessions[cookie.Value]
}

func GetBranches() ([]Branch, error) {
	rows, err := db.Query("SELECT id, name, address, pastor, phone, image, pastor_image, created_at FROM branches")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []Branch

	for rows.Next() {
		var branch Branch

		err := rows.Scan(
			&branch.ID,
			&branch.Name,
			&branch.Address,
			&branch.Pastor,
			&branch.Phone,
			&branch.Image,
			&branch.PastorImage,
			&branch.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		branches = append(branches, branch)
	}

	return branches, nil
}

type BranchDisplay struct {
	Number      int
	Name        string
	Address     string
	Pastor      string
	Phone       string
	Image       string
	PastorImage string
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	events, err := GetEvents()
	if err != nil {
		http.Error(w, "Failed to load events", http.StatusInternalServerError)
		return
	}

	for i := range events {

		eventDate, err := time.Parse("2006-01-02", events[i].EventDate)
		if err != nil {
			continue
		}

		events[i].EventDate = strings.ToUpper(eventDate.Format("02 Jan"))

		eventTime, err := time.Parse("15:04:05", events[i].EventTime)
		if err != nil {
			continue
		}

		events[i].EventTime = eventTime.Format("03:04 PM")
	}

	branches, err := GetBranches()
	if err != nil {
		http.Error(w, "Failed to load branches", http.StatusInternalServerError)
		return
	}

	Leaders, err := GetLeaders()
	if err != nil {
		http.Error(w, "Failed to load leaders", http.StatusInternalServerError)
		return
	}

	var branchDisplays []BranchDisplay

	for i, branch := range branches {
		branchDisplays = append(branchDisplays, BranchDisplay{
			Number:      i + 1,
			Name:        branch.Name,
			Address:     branch.Address,
			Pastor:      branch.Pastor,
			Phone:       branch.Phone,
			Image:       branch.Image,
			PastorImage: branch.PastorImage,
		})
	}

	data := struct {
		Events   []Event
		Branches []BranchDisplay
		Leaders  []Leader
	}{
		Events:   events,
		Branches: branchDisplays,
		Leaders:  Leaders,
	}

	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		fmt.Println("TEMPLATE PARSE ERROR:", err)
		http.Error(w, "Failed to load homepage", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Println("TEMPLATE EXECUTE ERROR:", err)
		http.Error(w, "Failed to load homepage", http.StatusInternalServerError)
		return
	}
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

		sessionID := fmt.Sprintf("%d", time.Now().UnixNano())
		sessionMutex.Lock()
		adminSessions[sessionID] = true
		sessionMutex.Unlock()
		LogAdminActivity("Admin logged in")

		http.SetCookie(w, &http.Cookie{
			Name:     "admin_session",
			Value:    sessionID,
			HttpOnly: true,
			Path:     "/",
		})

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

}

func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	var totalMessages int
	var unreadMessages int
	var totalBranches int
	var totalEvents int
	var err error

	err = db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&totalMessages)
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

	err = db.QueryRow("SELECT COUNT(*) FROM branches").Scan(&totalBranches)
	if err != nil {
		http.Error(w, "Failed to get branch count", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM events").Scan(&totalEvents)
	if err != nil {
		http.Error(w, "Failed to get event count", http.StatusInternalServerError)
		return
	}
	// fmt.Println("TOTAL EVENTS:", totalEvents)

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

	activityRows, err := db.Query(
		"SELECT id, action, created_at FROM admin_activity_logs ORDER BY id DESC LIMIT 5",
	)

	if err != nil {
		http.Error(w, "Failed to get recent activities", http.StatusInternalServerError)
		return
	}

	defer activityRows.Close()

	var recentActivities []AdminActivity

	for activityRows.Next() {

		var activity AdminActivity

		if err := activityRows.Scan(
			&activity.ID,
			&activity.Action,
			&activity.CreatedAt,
		); err != nil {
			http.Error(w, "Failed to read recent activities", http.StatusInternalServerError)
			return
		}

		recentActivities = append(recentActivities, activity)
	}

	if err := activityRows.Err(); err != nil {
		http.Error(w, "Error reading recent activities", http.StatusInternalServerError)
		return
	}

	data := DashboardData{
		TotalMessages:    totalMessages,
		UnreadMessages:   unreadMessages,
		TotalBranches:    totalBranches,
		RecentMessages:   recentMessages,
		TotalEvents:      totalEvents,
		RecentActivities: recentActivities,
	}

	templ, err := template.ParseFiles("admin-dashboard.html")
	if err != nil {
		http.Error(w, "Failed to load dashboard template", http.StatusInternalServerError)
		return
	}
	templ.Execute(w, data)
}

func AdminMessageHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

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

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	id := r.FormValue("id")

	_, err := db.Exec("UPDATE messages SET Status = ? WHERE id = ?", "read", id)
	if err != nil {
		http.Error(w, "Failed to mark message as read", http.StatusInternalServerError)
		return
	}

	LogAdminActivity("Marked a message as read")

	http.Redirect(w, r, "/admin/messages", http.StatusSeeOther)
}

func DeleteMessageHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	id := r.FormValue("id")
	_, err := db.Exec("DELETE FROM messages WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Failed to Delete", http.StatusInternalServerError)
		return
	}

	LogAdminActivity("Deleted a message")

	http.Redirect(w, r, "/admin/messages", http.StatusSeeOther)
}

func AdminBranchesHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	branches, err := GetBranches()
	if err != nil {
		http.Error(w, "Failed to load branches", http.StatusInternalServerError)
		return
	}

	data := struct {
		Branches []Branch
	}{
		Branches: branches,
	}

	err = template.Must(template.ParseFiles("admin-branches.html")).Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to load page", http.StatusInternalServerError)
		return
	}
}

func AddBranchHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	fmt.Println("METHOD:", r.Method)
	fmt.Println("PATH:", r.URL.Path)

	if r.Method == http.MethodGet {
		template.Must(template.ParseFiles("admin-branches-add.html")).Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {

		name := r.FormValue("name")
		address := r.FormValue("address")
		pastor := r.FormValue("pastor")
		phone := r.FormValue("phone")

		// 1. Handle Building Image
		imageFile, imageHeader, err := r.FormFile("image")
		if err != nil {
			http.Error(w, "Failed to upload building image", http.StatusBadRequest)
			return
		}
		defer imageFile.Close()

		filename := filepath.Base(imageHeader.Filename)
		imagePath := filepath.Join("images", filename)

		dst, err := os.Create(imagePath)
		if err != nil {
			http.Error(w, "Failed to save building image", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, imageFile)
		if err != nil {
			http.Error(w, "Failed to save building image", http.StatusInternalServerError)
			return
		}
		image := filepath.Base(imageHeader.Filename)

		// 2. Handle Pastor Image (Fixed error checking order)
		pastorFile, pastorHeader, err := r.FormFile("pastor_image")
		if err != nil {
			http.Error(w, "Failed to upload pastor image", http.StatusBadRequest)
			return
		}
		defer pastorFile.Close()

		pastorFilename := filepath.Base(pastorHeader.Filename)
		pastorImagePath := filepath.Join("images", pastorFilename)

		pastorDst, err := os.Create(pastorImagePath)
		if err != nil {
			http.Error(w, "Failed to save pastor image", http.StatusInternalServerError)
			return
		}
		defer pastorDst.Close()

		_, err = io.Copy(pastorDst, pastorFile)
		if err != nil {
			http.Error(w, "Failed to save pastor image", http.StatusInternalServerError)
			return
		}
		pastorImage := filepath.Base(pastorHeader.Filename)

		fmt.Println("Branch Name:", name)
		fmt.Println("Address:", address)
		fmt.Println("Pastor:", pastor)
		fmt.Println("Phone:", phone)
		fmt.Println("Image:", image)
		fmt.Println("Pastor Image:", pastorImage)

		// 3. Insert into Database
		_, err = db.Exec(
			"INSERT INTO branches (name, address, pastor, phone, image, pastor_image) VALUES (?, ?, ?, ?, ?, ?)",
			name,
			address,
			pastor,
			phone,
			image,
			pastorImage,
		)

		if err != nil {
			fmt.Println("DATABASE ERROR:", err)
			http.Error(w, "Failed to add branch", http.StatusInternalServerError)
			return
		}
		LogAdminActivity("Added a new branch")

		http.Redirect(w, r, "/admin/branches", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func EditBranchHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/admin/branches/edit/")

	// 1. Fetch existing branch data first (needed to keep old images if none are uploaded)
	var branch Branch
	err := db.QueryRow("SELECT id, name, address, pastor, phone, image, pastor_image, created_at FROM branches WHERE id = ?", id).
		Scan(&branch.ID, &branch.Name, &branch.Address, &branch.Pastor, &branch.Phone, &branch.Image, &branch.PastorImage, &branch.CreatedAt)

	if err != nil {
		http.Error(w, "Branch not found", http.StatusNotFound)
		return
	}

	LogAdminActivity("Edited a branch")

	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		address := r.FormValue("address")
		pastor := r.FormValue("pastor")
		phone := r.FormValue("phone")

		// Keep existing images by default
		imageName := branch.Image
		pastorImageName := branch.PastorImage

		// Handle Building Image (Optional during edit)
		imageFile, imageHeader, err := r.FormFile("image")
		if err == nil {
			defer imageFile.Close()
			filename := filepath.Base(imageHeader.Filename)
			imagePath := filepath.Join("images", filename)

			dst, err := os.Create(imagePath)
			if err != nil {
				http.Error(w, "Failed to save building image", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			_, err = io.Copy(dst, imageFile)
			if err != nil {
				http.Error(w, "Failed to save building image", http.StatusInternalServerError)
				return
			}
			imageName = filename
		}

		// Handle Pastor Image (Optional during edit)
		pastorFile, pastorHeader, err := r.FormFile("pastor_image")
		if err == nil {
			defer pastorFile.Close()
			pastorFilename := filepath.Base(pastorHeader.Filename)
			pastorImagePath := filepath.Join("images", pastorFilename)
			pastorImageName = pastorFilename
			pastorDst, err := os.Create(pastorImagePath)
			if err != nil {
				http.Error(w, "Failed to save pastor image", http.StatusInternalServerError)
				return
			}
			defer pastorDst.Close()

			_, err = io.Copy(pastorDst, pastorFile)
			if err != nil {
				http.Error(w, "Failed to save pastor image", http.StatusInternalServerError)
				return
			}
			pastorImageName = pastorFilename
		}

		// Fixed compilation bug: changed := to = because 'err' already exists
		_, err = db.Exec(
			"UPDATE branches SET name = ?, address = ?, pastor = ?, phone = ?, image = ?, pastor_image = ? WHERE id = ?",
			name,
			address,
			pastor,
			phone,
			imageName,
			pastorImageName,
			id,
		)

		if err != nil {
			fmt.Println("DATABASE ERROR:", err)
			http.Error(w, "Failed to update branch", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/branches", http.StatusSeeOther)
		return
	}

	// Render edit page for GET request
	err = template.Must(
		template.ParseFiles("admin-branches-edit.html"),
	).Execute(w, branch)

	if err != nil {
		http.Error(w, "Failed to load edit page", http.StatusInternalServerError)
		return
	}
}

func DeleteBranchHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/admin/branches/delete/")

	fmt.Println("Delete Branch ID:", id)

	_, err := db.Exec(
		"DELETE FROM branches WHERE id = ?",
		id,
	)

	if err != nil {
		fmt.Println("DATABASE ERROR:", err)
		http.Error(w, "Failed to delete branch", http.StatusInternalServerError)
		return
	}

	LogAdminActivity("Deleted a branch")

	http.Redirect(w, r, "/admin/branches", http.StatusSeeOther)
	return
}

func GetEvents() ([]Event, error) {
	rows, err := db.Query("SELECT id, title, description, event_date, event_time, location, created_at FROM events")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event

	for rows.Next() {
		var event Event

		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.EventDate,
			&event.EventTime,
			&event.Location,
			&event.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

func AdminEventsHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	events, err := GetEvents()
	if err != nil {
		http.Error(w, "Failed to load events", http.StatusInternalServerError)
		return
	}

	data := struct {
		Events []Event
	}{
		Events: events,
	}

	err = template.Must(
		template.ParseFiles("admin-events.html"),
	).Execute(w, data)

	if err != nil {
		http.Error(w, "Failed to load events page", http.StatusInternalServerError)
		return
	}
}

func AddEventHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	fmt.Println("METHOD:", r.Method)
	fmt.Println("PATH:", r.URL.Path)

	if r.Method == http.MethodGet {
		template.Must(
			template.ParseFiles("admin-events-add.html"),
		).Execute(w, nil)

		return
	}

	if r.Method == http.MethodPost {

		title := r.FormValue("title")
		description := r.FormValue("description")
		eventDate := r.FormValue("event_date")
		eventTime := r.FormValue("event_time")
		location := r.FormValue("location")

		fmt.Println("Event Title:", title)
		fmt.Println("Description:", description)
		fmt.Println("Event Date:", eventDate)
		fmt.Println("Event Time:", eventTime)
		fmt.Println("Location:", location)

		_, err := db.Exec(
			"INSERT INTO events (title, description, event_date, event_time, location) VALUES (?, ?, ?, ?, ?)",
			title,
			description,
			eventDate,
			eventTime,
			location,
		)

		if err != nil {
			fmt.Println("DATABASE ERROR:", err)
			http.Error(w, "Failed to add event", http.StatusInternalServerError)
			return
		}

		LogAdminActivity("Added a new event")

		http.Redirect(w, r, "/admin/events", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func AdminSettingsHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	var username string

	err := db.QueryRow(
		"SELECT username FROM admins LIMIT 1",
	).Scan(&username)

	if err != nil {
		http.Error(w, "Failed to load admin settings", http.StatusInternalServerError)
		return
	}

	data := AdminSettingsData{
		Username: username,
	}

	templ, err := template.ParseFiles("admin-settings.html")
	if err != nil {
		http.Error(w, "Failed to load settings page", http.StatusInternalServerError)
		return
	}

	err = templ.Execute(w, data)
	if err != nil {
		fmt.Println("TEMPLATE EXECUTE ERROR:", err)
		http.Error(w, "Failed to display settings page", http.StatusInternalServerError)
		return
	}
}

func EditEventHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/admin/events/edit/")

	fmt.Println("Event ID:", id)

	if r.Method == http.MethodPost {

		title := r.FormValue("title")
		description := r.FormValue("description")
		eventDate := r.FormValue("event_date")
		eventTime := r.FormValue("event_time")
		location := r.FormValue("location")

		fmt.Println("Updated Title:", title)
		fmt.Println("Updated Description:", description)
		fmt.Println("Updated Date:", eventDate)
		fmt.Println("Updated Time:", eventTime)
		fmt.Println("Updated Location:", location)

		_, err := db.Exec(
			"UPDATE events SET title = ?, description = ?, event_date = ?, event_time = ?, location = ? WHERE id = ?",
			title,
			description,
			eventDate,
			eventTime,
			location,
			id,
		)

		if err != nil {
			fmt.Println("DATABASE ERROR:", err)
			http.Error(w, "Failed to update event", http.StatusInternalServerError)
			return
		}

		LogAdminActivity("Edited an event")

		http.Redirect(w, r, "/admin/events", http.StatusSeeOther)
		return
	}

	var event Event

	err := db.QueryRow(
		"SELECT id, title, description, event_date, event_time, location, created_at FROM events WHERE id = ?",
		id,
	).Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&event.EventDate,
		&event.EventTime,
		&event.Location,
		&event.CreatedAt,
	)

	if err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	err = template.Must(
		template.ParseFiles("admin-events-edit.html"),
	).Execute(w, event)

	if err != nil {
		http.Error(w, "Failed to load edit page", http.StatusInternalServerError)
		return
	}

}

func DeleteEventHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/admin/events/delete/")

	fmt.Println("Delete Event ID:", id)

	_, err := db.Exec(
		"DELETE FROM events WHERE id = ?",
		id,
	)

	if err != nil {
		fmt.Println("DATABASE ERROR:", err)
		http.Error(w, "Failed to delete event", http.StatusInternalServerError)
		return
	}

	LogAdminActivity("Deleted an event")

	http.Redirect(w, r, "/admin/events", http.StatusSeeOther)
}

func AdminSettingsProfileHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		_, err := db.Exec(
			"UPDATE admins SET username = ? WHERE id = 1",
			username,
		)
		if err != nil {
			http.Error(w, "Failed to update UserName", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
	}

}

func AdminLogoutHandler(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("admin_session")

	if err == nil {
		sessionMutex.Lock()
		delete(adminSessions, cookie.Value)
		sessionMutex.Unlock()
		LogAdminActivity("Admin logged out")
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func AdminSettingsPasswordHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		currentPassword := r.FormValue("current_password")
		newPassword := r.FormValue("new_password")
		confirmPassword := r.FormValue("confirm_password")

		if newPassword != confirmPassword {
			http.Error(w, "New passwords do not match", http.StatusBadRequest)
			return
		}
		var hashedPassword string

		err := db.QueryRow(
			"SELECT password FROM admins WHERE id = 1",
		).Scan(&hashedPassword)

		if err != nil {
			http.Error(w, "Failed to get current password", http.StatusInternalServerError)
			return
		}

		err = bcrypt.CompareHashAndPassword(
			[]byte(hashedPassword),
			[]byte(currentPassword),
		)
		if err != nil {
			http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
			return
		}

		newHashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(newPassword),
			bcrypt.DefaultCost,
		)
		if err != nil {
			http.Error(w, "Failed to secure new password", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(
			"UPDATE admins SET password = ? WHERE id = 1",
			newHashedPassword,
		)
		if err != nil {
			http.Error(w, "Failed to update password", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
		return
	}

}

func LogAdminActivity(action string) {

	_, err := db.Exec(
		"INSERT INTO admin_activity_logs (action) VALUES (?)",
		action,
	)

	if err != nil {
		fmt.Println("ACTIVITY LOG ERROR:", err)
	}
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

	branches, err := GetBranches()
	if err != nil {
		log.Println("Failed to load branches:", err)
	} else {
		fmt.Println("Branches loaded:", len(branches))
	}

	// createAdmin()

	// 3. Register routes and static file handlers
	http.HandleFunc("/admin", AdminDashboardHandler)
	http.HandleFunc("/ai", AiHandler)
	http.HandleFunc("/admin/login", AdminLoginHandler)
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/contact", ContactHandler)
	http.HandleFunc("/admin/messages", AdminMessageHandler)
	http.HandleFunc("/admin/branches/add", AddBranchHandler)
	http.HandleFunc("/admin/branches/edit/", EditBranchHandler)
	http.HandleFunc("/admin/branches/delete/", DeleteBranchHandler)
	http.HandleFunc("/admin/events", AdminEventsHandler)
	http.HandleFunc("/admin/events/add", AddEventHandler)
	http.HandleFunc("/admin/events/edit/", EditEventHandler)
	http.HandleFunc("/admin/events/delete/", DeleteEventHandler)
	http.HandleFunc("/admin/settings", AdminSettingsHandler)
	http.HandleFunc("/admin/settings/password", AdminSettingsPasswordHandler)
	http.HandleFunc("/admin/logout", AdminLogoutHandler)
	http.HandleFunc("/admin/activity", AdminActivityHandler)
	http.HandleFunc("/admin/leadership", AdminLeadershipHandler)
	http.HandleFunc("/admin/leadership/add", AddLeaderHandler)
	http.HandleFunc("/admin/leadership/edit/", EditLeaderHandler)
	http.HandleFunc("/admin/leadership/delete/", DeleteLeaderHandler)

	http.HandleFunc("/admin/messages/delete", DeleteMessageHandler)
	http.HandleFunc("/admin/messages/read", MarkMessageReadHandler)
	http.HandleFunc("/admin/branches", AdminBranchesHandler)
	http.HandleFunc("/admin/settings/profile", AdminSettingsProfileHandler)

	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	http.Handle("/style.css", http.FileServer(http.Dir(".")))
	http.Handle("/script.js", http.FileServer(http.Dir(".")))

	fmt.Println("Server running on http://localhost:8082")

	// 4. Start HTTP listener (blocking call)
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatal(err)
	}

}
