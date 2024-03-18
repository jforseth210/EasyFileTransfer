package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	application := app.New()
	window := application.NewWindow("Hello World")
	label := widget.NewLabel(getIPAsWords(GetLocalIP()))
	window.SetContent(label)
	window.SetOnDropped(func(position fyne.Position, uris []fyne.URI) {
		for _, uri := range uris {
			println(uri.String())
		}
	})
	// Channel to communicate between HTTP server and GUI
	updateLabel := make(chan string)
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				// Parse multipart form
				err := r.ParseMultipartForm(10 << 20) // 10 MB max
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				// Get the uploaded file
				file, handler, err := r.FormFile("file")
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				defer file.Close()

				saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
					if err != nil {
						fmt.Println("Error saving file:", err)
						return
					}
					if writer == nil {
						fmt.Println("Save file dialog canceled")
						return
					}
					io.Copy(writer, file)
				}, window)
				saveDialog.SetFileName(handler.Filename)
				saveDialog.Show()

				updateLabel <- fmt.Sprintf("File uploaded: %s", handler.Filename)
				return
			}
			// If not a POST request, display the form
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, "<form enctype=\"multipart/form-data\" method=\"post\">"+
				"<input type=\"file\" name=\"file\"/><br>"+
				"<input type=\"submit\" value=\"Upload\"/></form>")
		})

		http.ListenAndServe(":8080", nil)
	}()
	// Update label text based on messages from HTTP server
	go func() {
		for {
			msg := <-updateLabel
			label.SetText(msg)
			window.SetContent(label)
		}
	}()

	window.ShowAndRun()
}

func GetLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP
}

// wordfrequency.info 256 most common nouns
var wordList = []string{
	"time", "people", "year", "way", "thing", "man", "day", "life", "woman", "world", "child", "school", "state", "family", "president", "house", "student", "part", "place", "problem", "country", "week", "point", "hand", "group", "guy", "case", "question", "work", "night", "game", "number", "money", "lot", "book", "system", "government", "city", "company", "story", "job", "friend", "word", "fact", "right", "month", "program", "business", "home", "kind", "study", "issue", "name", "idea", "room", "percent", "law", "power", "kid", "war", "head", "mother", "team", "eye", "side", "water", "service", "area", "person", "end", "hour", "line", "girl", "father", "information", "car", "minute", "party", "back", "health", "reason", "member", "community", "news", "body", "level", "boy", "university", "change", "center", "face", "food", "history", "result", "morning", "parent", "office", "research", "door", "court", "moment", "street", "policy", "table", "care", "process", "teacher", "data", "death", "experience", "plan", "education", "age", "sense", "show", "college", "music", "mind", "class", "police", "use", "effect", "season", "tax", "heart", "son", "art", "market", "air", "force", "foot", "baby", "love", "republican", "interest", "security", "control", "rate", "report", "nation", "action", "wife", "decision", "value", "phone", "thanks", "event", "site", "church", "model", "relationship", "movie", "field", "player", "couple", "record", "difference", "light", "development", "role", "view", "price", "effort", "voice", "department", "leader", "photo", "space", "project", "position", "million", "film", "need", "type", "town", "article", "road", "form", "chance", "drug", "situation", "practice", "science", "brother", "matter", "image", "star", "cost", "post", "society", "picture", "piece", "paper", "energy", "building", "doctor", "activity", "american", "media", "evidence", "product", "arm", "technology", "comment", "look", "term", "color", "choice", "source", "mom", "director", "rule", "campaign", "ground", "election", "page", "test", "patient", "video", "support", "rest", "step", "opportunity", "official", "oil", "call", "organization", "character", "county", "future", "dad", "industry", "second", "list", "stuff", "figure", "attention", "risk", "fire", "dog", "hair", "condition", "wall", "daughter", "deal", "author", "truth", "husband", "period", "series", "order", "officer", "land", "computer", "thought", "economy"}

func getIPAsWords(ip net.IP) string {
	digits := ip.To4()
	var digitStrings []string
	for _, digit := range digits {
		digitStrings = append(digitStrings, wordList[digit])
	}
	println(strings.Join(digitStrings, " "))
	return strings.Join(digitStrings, " ")
}
