package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image/png"
	"io"
	"net"
	"os"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func run() {
	conn, err := net.Dial("tcp", "SERVER_IP:9090") // Ganti SERVER_IP
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Pertama, decode ukuran frame agar bisa set window
	// (misal server selalu kirim frame ukuran sama)
	sizeBuf := make([]byte, 4)
	_, err = io.ReadFull(reader, sizeBuf)
	if err != nil {
		fmt.Println("Read size error:", err)
		return
	}
	imgLen := int(sizeBuf[0])<<24 | int(sizeBuf[1])<<16 | int(sizeBuf[2])<<8 | int(sizeBuf[3])
	imgBuf := make([]byte, imgLen)
	_, err = io.ReadFull(reader, imgBuf)
	if err != nil {
		fmt.Println("Read image error:", err)
		return
	}
	img, err := png.Decode(bytes.NewReader(imgBuf))
	if err != nil {
		fmt.Println("Decode error:", err)
		return
	}

	bounds := img.Bounds()
	win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Title:  "Go Screen Stream Client",
		Bounds: pixel.R(0, 0, float64(bounds.Dx()), float64(bounds.Dy())),
		VSync:  true,
	})
	if err != nil {
		fmt.Println("Window error:", err)
		return
	}

	// Convert first image to sprite
	pic := pixel.PictureDataFromImage(img)
	sprite := pixel.NewSprite(pic, pic.Bounds())

	go func() {
		for {
			// Read size
			sizeBuf := make([]byte, 4)
			_, err := io.ReadFull(reader, sizeBuf)
			if err != nil {
				fmt.Println("Read size error:", err)
				os.Exit(0)
			}
			imgLen := int(sizeBuf[0])<<24 | int(sizeBuf[1])<<16 | int(sizeBuf[2])<<8 | int(sizeBuf[3])
			imgBuf := make([]byte, imgLen)
			_, err = io.ReadFull(reader, imgBuf)
			if err != nil {
				fmt.Println("Read image error:", err)
				os.Exit(0)
			}
			img, err := png.Decode(bytes.NewReader(imgBuf))
			if err != nil {
				fmt.Println("Decode error:", err)
				continue
			}
			pic := pixel.PictureDataFromImage(img)
			sprite.Set(pic, pic.Bounds())
		}
	}()

	for !win.Closed() {
		win.Clear(pixel.RGB(0, 0, 0))
		sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
