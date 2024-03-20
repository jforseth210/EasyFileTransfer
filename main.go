package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	app := app.New()
	window := app.NewWindow("File Transfer")

	// Send Files Column
	sendFilesLabel := widget.NewLabel("Drop files anywhere in this window to send them")

	sendFilesColumn := container.NewVBox(
		widget.NewLabel("Send Files"),
		sendFilesLabel,
	)

	// Receive Files Column
	receiveFilesText := widget.NewLabel("This device's name is \"" + EncodeIPToWords(GetLocalIP()) + "\".\n To send files to this device, leave this window open\n and enter this name on the sending device.")

	receiveFilesColumn := container.NewVBox(
		widget.NewLabel("Receive Files"),
		receiveFilesText,
	)

	// Combine both columns
	columns := container.NewGridWithColumns(2, sendFilesColumn, receiveFilesColumn)

	window.SetOnDropped(func(position fyne.Position, uris []fyne.URI) {
		uriStrings := []string{}
		for _, uri := range uris {
			uriStrings = append(uriStrings, uri.Path())
		}
		addressEntry := widget.NewEntry()
		form := widget.NewForm()
		form.Append("Device Name", addressEntry)
		dialog.ShowForm("Send Files", "Send", "Cancel", form.Items, func(confirmed bool) {
			if !confirmed {
				return
			}
			err := UploadMultipleFiles("http://"+DecodeIPFromWords(addressEntry.Text).String()+":8080", uriStrings)
			if err != nil {
				log.Fatal(err)
			}
		}, window)

	})

	window.SetContent(columns)
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
			handlers := []*multipart.FileHeader{}
			requestFiles := request.MultipartForm.File["file"]
			for _, handler := range requestFiles {

				file, err := handler.Open()
				if err != nil {
					http.Error(responseWriter, err.Error(), http.StatusBadRequest)
					return
				}
				handlers = append(handlers, handler)
				defer file.Close()
			}

			dialog.ShowConfirm("Accept Files",
				"\""+EncodeIPToWords(net.ParseIP(strings.Split(request.RemoteAddr, ":")[0]))+"\" wants to send you these files:"+"\n"+strings.Join(getFileNames(handlers), "\n")+"\n Accept?", func(ok bool) {
					if !ok {
						return
					}
					dialog.ShowFolderOpen(func(uris fyne.ListableURI, err error) {
						if err != nil {
							http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
							return
						}

						if uris == nil {
							return
						}
						for i, handler := range handlers {
							file, err := handler.Open()
							if err != nil {
								http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
								return
							}
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
				}, window)
		})

		http.ListenAndServe(":8080", nil)
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

func getFileNames(handlers []*multipart.FileHeader) []string {
	var fileNames []string
	for _, handler := range handlers {
		fileNames = append(fileNames, handler.Filename)
	}
	return fileNames
}

func UploadMultipleFiles(url string, filePaths []string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
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
