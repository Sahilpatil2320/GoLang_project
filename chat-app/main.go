package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

// Room-specific chat
var rooms = make(map[string]map[*websocket.Conn]bool)
var broadcasts = make(map[string]chan Message)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", loginPage)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/room/", roomPage)
	http.HandleFunc("/ws/", handleWebSocket)

	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}

// Login Page
func loginPage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/login.html")
	tmpl.Execute(w, nil)
}

// Handle login form and redirect to a room
func handleLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	room := r.FormValue("room")

	if username == "" || room == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "username",
		Value: username,
	})

	http.Redirect(w, r, "/room/"+room, http.StatusSeeOther)
}

// Render the chat room page
func roomPage(w http.ResponseWriter, r *http.Request) {
	room := strings.TrimPrefix(r.URL.Path, "/room/")

	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl, _ := template.ParseFiles("templates/room.html")
	tmpl.Execute(w, struct {
		Room     string
		Username string
	}{
		Room:     room,
		Username: cookie.Value,
	})
}

// Handle WebSocket per room
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	room := strings.TrimPrefix(r.URL.Path, "/ws/")
	usernameCookie, err := r.Cookie("username")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	username := usernameCookie.Value

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}

	// Initialize room if needed
	if _, ok := rooms[room]; !ok {
		rooms[room] = make(map[*websocket.Conn]bool)
		broadcasts[room] = make(chan Message)
		go handleMessages(room)
	}

	rooms[room][conn] = true

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			delete(rooms[room], conn)
			conn.Close()
			break
		}
		msg.Username = username
		broadcasts[room] <- msg
	}
}

// Broadcast to all clients in a room
func handleMessages(room string) {
	for {
		msg := <-broadcasts[room]
		for client := range rooms[room] {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(rooms[room], client)
			}
		}
	}
}
