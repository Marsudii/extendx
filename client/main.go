package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	imageCanvas *canvas.Image
)

func main() {
	a := app.New()
	w := a.NewWindow("Extendx Client")

	ipInput := widget.NewEntry()
	ipInput.SetPlaceHolder("Enter server IP (e.g. 192.168.1.10)")

	startBtn := widget.NewButton("Connect & Stream", func() {
		ip := ipInput.Text
		if ip == "" {
			// Show alert if IP is empty
			// CAPTURE ALERT ERROR POPUP JIKA GAGAL CONNECT
			alert := widget.NewLabel("Capture error: Please check your server IP and ensure the server is running.")
			alert.Hide()
			return
		}
		go connectAndStream(ip)
	})

	imageCanvas = canvas.NewImageFromImage(nil)
	imageCanvas.FillMode = canvas.ImageFillContain

	w.SetContent(container.NewVBox(
		ipInput,
		startBtn,
		imageCanvas,
	))
	// CAPTURE ALERT ERROR POPUP JIKA GAGAL CONNECT
	alert := widget.NewLabel("Capture error: Please check your server IP and ensure the server is running.")
	alert.Hide()

	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}

func connectAndStream(ip string) {
	conn, err := net.Dial("tcp", ip+":9090")
	if err != nil {
		log.Println("❌ Error connecting to server:", err)
		return
	}
	defer conn.Close()

	for {
		lengthBytes := make([]byte, 4)
		if _, err := io.ReadFull(conn, lengthBytes); err != nil {
			log.Println("❌ Failed to read length:", err)
			return
		}

		length := int(lengthBytes[0])<<24 | int(lengthBytes[1])<<16 | int(lengthBytes[2])<<8 | int(lengthBytes[3])
		if length <= 0 || length > 5*1024*1024 {
			log.Println("❌ Invalid image length")
			continue
		}

		imgData := make([]byte, length)
		if _, err := io.ReadFull(conn, imgData); err != nil {
			log.Println("❌ Failed to read image:", err)
			return
		}

		img, err := jpeg.Decode(bytes.NewReader(imgData))
		if err != nil {
			log.Println("❌ Failed to decode image:", err)
			continue
		}

		// Update UI safely
		updateImage(img)
		time.Sleep(100 * time.Millisecond)
	}
}

func updateImage(img image.Image) {
	if imageCanvas != nil {
		imageCanvas.Image = img
		imageCanvas.Refresh()
	}
}
