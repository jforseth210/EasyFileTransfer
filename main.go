package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	application := app.New()
	window := application.NewWindow("File Share")
	label := widget.NewLabel(getIPAsWords(GetLocalIP()))
	window.SetContent(label)
	window.SetOnDropped(func(position fyne.Position, uris []fyne.URI) {
		uriStrings := []string{}
		for _, uri := range uris {
			uriStrings = append(uriStrings, uri.Path())
		}
		err := UploadMultipleFiles("http://10.21.20.202:8080", "file", uriStrings)
		if err != nil {
			log.Fatal(err)
		}
	})
	// Channel to communicate between HTTP server and GUI
	updateLabel := make(chan string)

	go func() {
		http.HandleFunc("/", func(responseWriter http.ResponseWriter, request *http.Request) {
			if request.Method != "POST" {
				return
			}
			// Parse multipart form
			err := request.ParseMultipartForm(10 << 20) // 10 MB max
			if err != nil {
				http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
				return
			}
			files := []multipart.File{}
			handlers := []*multipart.FileHeader{}
			for i := 0; i < len(request.MultipartForm.File); i++ {
				// Get the uploaded file
				file, handler, err := request.FormFile("file" + strconv.Itoa(i))
				if err != nil {
					http.Error(responseWriter, err.Error(), http.StatusBadRequest)
					return
				}
				files = append(files, file)
				handlers = append(handlers, handler)
				defer file.Close()
			}
			dialog.ShowFolderOpen(func(uris fyne.ListableURI, err error) {
				if err != nil {
					http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
					return
				}

				if uris == nil {
					return
				}
				for i, file := range files {
					// Create a new file in the selected directory with the same name as the uploaded file
					emptyFile, err := os.Create(filepath.Join(uris.Path(), handlers[i].Filename))
					if err != nil {
						http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
						return
					}
					defer file.Close()

					// Copy the contents of the uploaded file to the newly created file
					_, err = io.Copy(emptyFile, file)
					if err != nil {
						http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
						return
					}
					updateLabel <- fmt.Sprintf("File uploaded: %s", handlers[i].Filename)
				}
			}, window)
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

func UploadMultipleFiles(url string, paramName string, filePaths []string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for i, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		part, err := writer.CreateFormFile("file"+strconv.Itoa(i), filepath.Base(filePath))
		if err != nil {
			return err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return err
		}
	}
	err := writer.Close()
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	// Handle the server response...
	return nil
}
