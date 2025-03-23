// ===============================
// 🖥️ CLIENT (WINDOWS) - udp_screen_client.go
// ===============================
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/chai2010/webp"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	currentImage              *ebiten.Image
	windowWidth, windowHeight = 640, 480
)

type game struct{}

func (g *game) Update() error { return nil }
func (g *game) Draw(screen *ebiten.Image) {
	if currentImage != nil {
		screen.DrawImage(currentImage, nil)
	} else {
		ebitenutil.DebugPrint(screen, "Waiting for frame...")
	}
}
func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return windowWidth, windowHeight
}

func startUDPListener(port string) {
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		log.Fatal("UDP Resolve error:", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal("UDP Listen error:", err)
	}
	defer conn.Close()

	log.Println("📥 Listening for UDP screen packets on port", port)
	buffer := make([]byte, 65535)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Read error:", err)
			continue
		}

		go func(data []byte) {
			if len(data) < 8 {
				return
			}
			w := binary.BigEndian.Uint32(data[:4])
			h := binary.BigEndian.Uint32(data[4:8])

			img, err := webp.Decode(bytes.NewReader(data[8:n]))
			if err != nil {
				log.Println("❌ Decode error:", err)
				return
			}
			windowWidth = int(w)
			windowHeight = int(h)
			currentImage = ebiten.NewImageFromImage(img)
		}(append([]byte{}, buffer[:n]...))
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Masukkan IP Server (Mac): ")
	serverIP, _ := reader.ReadString('\n')
	serverIP = strings.TrimSpace(serverIP)
	fmt.Println("✅ Listening for screen stream from:", serverIP)

	go startUDPListener("8088")
	ebiten.SetWindowTitle("Live Screen via UDP")
	ebiten.SetWindowSize(windowWidth, windowHeight)
	if err := ebiten.RunGame(&game{}); err != nil {
		log.Fatal(err)
	}
}
