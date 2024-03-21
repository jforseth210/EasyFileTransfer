package main

import (
	"fmt"

	"log"
	"net"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// Create the UI
	app := app.New()
	window := app.NewWindow("File Transfer")
	nameLabelLabel := widget.NewLabel("Name")
	nameLabelLabel.Alignment = fyne.TextAlignCenter

	nameLabel := widget.NewLabel(EncodeIPToWords(GetLocalIP()))
	nameLabel.TextStyle = widget.RichTextStyleHeading.TextStyle
	nameLabel.Alignment = fyne.TextAlignCenter

	// Send Files Column
	sendFilesLabel := widget.NewLabel("Drop files anywhere in this window to send them")
	sendFilesColumn := container.NewVBox(
		nameLabelLabel,
		nameLabel,
		sendFilesLabel,
	)

	// Get dropped files
	window.SetOnDropped(func(position fyne.Position, uris []fyne.URI) {
		uriStrings := []string{}
		for _, uri := range uris {
			uriStrings = append(uriStrings, uri.Path())
		}
		// Show addess popup form to enter recipient
		ShowAddressForm(window, uriStrings)
	})

	window.SetContent(sendFilesColumn)

	window.Resize(fyne.NewSize(600, 400))
	// Listen for incoming files
	go RecieveFiles(window)
	window.ShowAndRun()

}

// https://gosamples.dev/local-ip-address/
func GetLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP
}

// Display address form, and because of nested callbacks,
// handle submission too.
func ShowAddressForm(window fyne.Window, uriStrings []string) {
	// Create the textbox
	addressEntry := widget.NewEntry()
	form := widget.NewForm()
	form.Append("", addressEntry)
	dialog.ShowForm("Enter Recipient Name", "Send", "Cancel", form.Items, func(confirmed bool) {
		// They cancelled, bail
		if !confirmed {
			return
		}

		// Try to parse the address
		ip := DecodeIPFromWords(addressEntry.Text)
		if ip == nil {
			// Not a valid address, complain
			dialog.ShowError(fmt.Errorf("Unable to connect to "+addressEntry.Text), window)
			return
		}
		// Try to send the files
		err := UploadMultipleFiles("http://"+ip.String()+":2364", uriStrings)
		if err != nil {
			// Something went wrong, complain
			dialog.ShowError(err, window)
		}
	}, window)
}
