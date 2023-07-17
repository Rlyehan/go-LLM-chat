package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
  "log"
)

var chatData = make(map[string][]string)
var chatMutex sync.Mutex

type Item struct {
	Text string `json:"text"`
}

type Response struct {
	Response string `json:"response"`
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/start", startChatHandler)
	http.HandleFunc("/chat", handleChat)
	http.HandleFunc("/query", handleQuery)

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	fmt.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func startChatHandler(w http.ResponseWriter, r *http.Request) {
	chatMutex.Lock()
	defer chatMutex.Unlock()

	chatId := fmt.Sprintf("%d", len(chatData))
	chatData[chatId] = []string{}

	http.Redirect(w, r, "/chat?chat_id="+chatId, http.StatusSeeOther)
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/chat.html")
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	chatMutex.Lock()
	defer chatMutex.Unlock()

	if err := r.ParseForm(); err != nil {
    log.Println("Error parsing form:", err)
		fmt.Fprint(w, "Something went wrong.")
		return
	}

	chatId := r.Form.Get("chat_id")
	userQuery := r.Form.Get("query")
	if userQuery == "" {
		fmt.Fprint(w, "You didn't ask anything.")
		return
	}

	chatHistory, ok := chatData[chatId]
	if !ok {
		chatHistory = []string{}
	}

	chatHistory = append(chatHistory, "User: "+userQuery)

	item := &Item{
		Text: userQuery,
	}

	jsonData, err := json.Marshal(item)
	if err != nil {
		fmt.Fprint(w, "Something went wrong.")
		return
	}

	resp, err := http.Post("http://localhost:8000/api/query/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
    log.Println("Error parsing form:", err)
		fmt.Fprint(w, "Something went wrong.")
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var result Response
	err = json.Unmarshal(body, &result)
	if err != nil {
    log.Println("Error parsing form:", err)
		fmt.Fprint(w, "Something went wrong.")
		return
	}

	response := result.Response
	chatHistory = append(chatHistory, "Chatbot: "+response)

	chatData[chatId] = chatHistory

	fmt.Fprint(w, "<p class='message user'>"+chatHistory[len(chatHistory)-2]+"</p>")
	fmt.Fprint(w, "<p class='message bot'>"+chatHistory[len(chatHistory)-1]+"</p>")
}

