package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"

	"github.com/chai2010/webp"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var conn net.Conn
var currentImage *ebiten.Image
var windowWidth, windowHeight = 640, 480 // Default awal agar tidak error
var once sync.Once

type game struct{}

func (g *game) Update() error {
	// Baca total length frame (uint32)
	var totalLen uint32
	if err := binary.Read(conn, binary.BigEndian, &totalLen); err != nil {
		return err
	}

	// Baca width dan height (masing-masing uint32)
	var width, height uint32
	if err := binary.Read(conn, binary.BigEndian, &width); err != nil {
		return err
	}
	if err := binary.Read(conn, binary.BigEndian, &height); err != nil {
		return err
	}

	// Baca frame data (WebP image)
	imageData := make([]byte, totalLen-8)
	_, err := io.ReadFull(conn, imageData)
	if err != nil {
		return err
	}

	// Decode WebP image
	img, err := webp.Decode(bytes.NewReader(imageData))
	if err != nil {
		log.Println(" Decode WebP failed:", err)
		return err
	}

	// Set ukuran window dari server (hanya sekali)
	once.Do(func() {
		windowWidth = int(width)
		windowHeight = int(height)
		ebiten.SetWindowSize(windowWidth, windowHeight)
	})

	currentImage = ebiten.NewImageFromImage(img)
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	if currentImage != nil {
		screen.DrawImage(currentImage, nil)
	} else {
		ebitenutil.DebugPrint(screen, "Waiting for frame...")
	}
}

func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// Cegah crash jika resolusi belum tersedia
	if windowWidth <= 0 || windowHeight <= 0 {
		return 640, 480
	}
	return windowWidth, windowHeight
}

func main() {
	var err error

	// Ganti IP sesuai IP server (Mac kamu)
	conn, err = net.Dial("tcp", "10.10.10.7:8088")
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer conn.Close()

	ebiten.SetWindowTitle(" Live Screen Viewer")
	ebiten.SetWindowSize(windowWidth, windowHeight) // default awal
	if err := ebiten.RunGame(&game{}); err != nil {
		log.Fatal(err)
	}
}
