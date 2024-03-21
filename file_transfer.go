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
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func RecieveFiles(window fyne.Window) {
	// Listen for incoming files
	http.HandleFunc("/", func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			return
		}

		// Parse multipart form
		err := request.ParseMultipartForm(64 << 30) // 64 GB max

		if err != nil {
			// Something went wrong with parsing
			http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
			dialog.ShowError(fmt.Errorf("Error recieving files"), window)
			return
		}

		handlers := []*multipart.FileHeader{}
		requestFiles := request.MultipartForm.File["file"]
		// Get all the file handlers
		for _, handler := range requestFiles {
			file, err := handler.Open()
			if err != nil {
				// Something's wrong with this file...
				http.Error(responseWriter, err.Error(), http.StatusBadRequest)
				dialog.ShowError(fmt.Errorf("Error recieving files"), window)
				return
			}
			handlers = append(handlers, handler)
			defer file.Close()
		}

		// Ask the user if they want to accept the files
		dialog.ShowConfirm("Accept Files",
			"\""+EncodeIPToWords(net.ParseIP(strings.Split(request.RemoteAddr, ":")[0]))+"\" wants to send you these files:"+"\n"+strings.Join(GetFileNames(handlers), "\n")+"\n Accept?", func(ok bool) {
				if !ok {
					// They don't want to accept the files
					return
				}
				// Ask where to save them
				dialog.ShowFolderOpen(func(uris fyne.ListableURI, err error) {
					if err != nil {
						// Something went wrong choosing the directory
						http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
						dialog.ShowError(fmt.Errorf("Error picking folder"), window)
						return
					}

					if uris == nil {
						dialog.ShowError(fmt.Errorf("Error recieving files"), window)
						// We don't have a directory to save to
						return
					}
					// Save the files
					for i, handler := range handlers {
						file, err := handler.Open()
						if err != nil {
							// Something went wrong with this file... =
							http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
							dialog.ShowError(fmt.Errorf("Error reading: "+handler.Filename), window)
							return
						}

						// Create a new file in the selected directory with the same name as the uploaded file
						emptyFile, err := os.Create(filepath.Join(uris.Path(), handlers[i].Filename))
						if err != nil {
							// Something went wrong creating the file
							http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
							dialog.ShowError(fmt.Errorf("Error creating: "+handler.Filename), window)
							return
						}
						defer file.Close()

						// Copy the contents of the uploaded file to the newly created file
						_, err = io.Copy(emptyFile, file)
						if err != nil {
							http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
							dialog.ShowError(fmt.Errorf("Error saving: "+handler.Filename), window)
							return
						}
					}
				}, window)
			}, window)
	})

	// Arbitrary port for now
	http.ListenAndServe(":2364", nil)
}
func UploadMultipleFiles(url string, filePaths []string) error {
 	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, filePath := range filePaths {
		// Open each file
		file, err := os.Open(filePath)
		if err != nil {
			// Handle read errors
			return fmt.Errorf("Error opening file: " + filePath)
		}
		// Check if it's a directory
		fi, err := file.Stat()
		switch {
		case err != nil:
			// Something went wrong with the file
			return fmt.Errorf("Error opening file: " + filePath)
		case fi.IsDir():
			// It's a directory and we don't support that
			return fmt.Errorf("Folders aren't supported. Please select a file instead.")
		default:
			// It's not a directory, we're fine
		}
		defer file.Close()

		// Create an entry in the form
		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			return fmt.Errorf("Error preparing to send file: " + filePath)
		}

		// Copy the file data
		_, err = io.Copy(part, file)
		if err != nil {
			return fmt.Errorf("Error preparing to send file: " + filePath)
		}
	}
	err := writer.Close()
	if err != nil {
		return fmt.Errorf("Error preparing to send file(s)")
	}

	// Create the http request
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("Error sending file(s)")
	}
	// Add content type header
	request.Header.Add("Content-Type", writer.FormDataContentType())
	// Create the client and send the request
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		// Handle errors
		log.Println(err)
		return fmt.Errorf("Error sending file(s)")
	}
	defer response.Body.Close()
	// Handle the server response...
	return nil
}

// Convert FileHeader to friendly human string.
func GetFileNames(handlers []*multipart.FileHeader) []string {
	var fileNames []string
	for _, handler := range handlers {
		fileNames = append(fileNames, handler.Filename)
	}
	return fileNames
}
