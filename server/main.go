package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/kbinani/screenshot"
)

var (
	selectedMonitor = 0
	selectedWidth   = 1280
	selectedHeight  = 720
)

func main() {
	if runtime.GOOS != "windows" {
		log.Println("⚠️ Warning: This version is optimized for Windows only")
	}

	myApp := app.New()
	window := myApp.NewWindow("Extendx (Server - Windows)")

	window.Resize(fyne.NewSize(600, 400))
	window.SetFixedSize(true)

	// Get IP
	localIP, err := GetIPLocal()
	if err != nil {
		log.Fatal(err)
	}

	labelTitle := widget.NewLabel("EXTENDX (SERVER - WINDOWS)")
	labelIP := widget.NewLabel("IP LOCAL: " + localIP)
	labelMonitorInfo := widget.NewLabel("")
	labelResolution := widget.NewLabel("")

	// Monitor list
	monitorCount := screenshot.NumActiveDisplays()
	monitorOptions := []string{}
	for i := 0; i < monitorCount; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		monitorOptions = append(monitorOptions, fmt.Sprintf("Monitor %d - %dx%d", i, bounds.Dx(), bounds.Dy()))
	}

	comboSelectMonitor := widget.NewSelect(monitorOptions, func(selected string) {
		selectedMonitor = GetMonitorIndexFromLabel(selected)
		bounds := screenshot.GetDisplayBounds(selectedMonitor)
		labelMonitorInfo.SetText(fmt.Sprintf("🖥️ Monitor %d: %dx%d", selectedMonitor, bounds.Dx(), bounds.Dy()))
	})

	comboSelectResolution := widget.NewSelect([]string{"480p", "720p", "1080p"}, func(selected string) {
		switch selected {
		case "480p":
			selectedWidth, selectedHeight = 640, 480
		case "720p":
			selectedWidth, selectedHeight = 1280, 720
		case "1080p":
			selectedWidth, selectedHeight = 1920, 1080
		}
		labelResolution.SetText(fmt.Sprintf("Selected resolution: %dx%d", selectedWidth, selectedHeight))
	})

	submitButton := widget.NewButton("Start Capture", func() {
		go StartCapture(localIP)
	})

	window.SetContent(container.NewVBox(
		labelTitle,
		labelIP,
		widget.NewLabel("Select Monitor:"),
		comboSelectMonitor,
		labelMonitorInfo,
		widget.NewLabel("Choose Resolution:"),
		comboSelectResolution,
		labelResolution,
		submitButton,
	))

	window.ShowAndRun()
}

func GetIPLocal() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.To4() != nil && !ip.IsLoopback() {
				return ip.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no suitable IP found")
}

func GetMonitorIndexFromLabel(label string) int {
	parts := strings.Split(label, " ")
	if len(parts) < 2 {
		return 0
	}
	index, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	return index
}

func StartCapture(ip string) {
	ln, err := net.Listen("tcp", ip+":9090")
	if err != nil {
		log.Fatal("❌ Failed to bind to port:", err)
	}
	log.Println("📡 Waiting for client to connect at " + ip + ":9090")

	conn, err := ln.Accept()
	if err != nil {
		log.Fatal("❌ Failed to accept connection:", err)
	}
	log.Println("✅ Client connected.")

	for {
		img, err := screenshot.CaptureDisplay(selectedMonitor)
		if err != nil {
			log.Println("❌ Capture error:", err)
			continue
		}

		resized := resizeImage(img, selectedWidth, selectedHeight)

		var buf bytes.Buffer
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 50})
		if err != nil {
			log.Println("❌ JPEG encode failed:", err)
			continue
		}

		length := int32(len(buf.Bytes()))
		lengthBytes := []byte{
			byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length),
		}
		conn.Write(lengthBytes)
		conn.Write(buf.Bytes())

		time.Sleep(100 * time.Millisecond)
	}
}

func resizeImage(img image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	srcBounds := img.Bounds()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := srcBounds.Min.X + x*srcBounds.Dx()/width
			srcY := srcBounds.Min.Y + y*srcBounds.Dy()/height
			dst.Set(x, y, img.At(srcX, srcY))
		}
	}
	return dst
}
