// ===============================
// 📡 SERVER (MAC) - udp_screen_server.go
// ===============================
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chai2010/webp"
	"github.com/kbinani/screenshot"
	"github.com/nfnt/resize"
)

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no suitable local IP address found")
}

func chooseMonitor() int {
	localIP, _ := getLocalIP()
	fmt.Println("IP LOCAL YOU:", localIP)
	displayCount := screenshot.NumActiveDisplays()
	fmt.Printf("🖥️ Available %d monitor:\n", displayCount)
	for i := 0; i < displayCount; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		fmt.Printf("[%d] %dx%d @ (%d,%d)\n", i, bounds.Dx(), bounds.Dy(), bounds.Min.X, bounds.Min.Y)
	}

	fmt.Print("Choose monitor (number): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	index, err := strconv.Atoi(input)
	if err != nil || index < 0 || index >= displayCount {
		log.Fatalf("❌ Invalid monitor selection")
	}
	return index
}

func main() {
	quality := flag.Int("q", 80, "WebP quality (0-100)")
	width := flag.Int("w", 960, "Resize width")
	height := flag.Int("h", 540, "Resize height")
	port := flag.String("port", "8088", "UDP port")
	flag.Parse()

	// Input IP tujuan (Windows client)
	fmt.Print("Masukkan IP client (Windows): ")
	reader := bufio.NewReader(os.Stdin)
	targetIP, _ := reader.ReadString('\n')
	targetIP = strings.TrimSpace(targetIP)

	dstAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", targetIP, *port))
	if err != nil {
		log.Fatal("ResolveUDPAddr error:", err)
	}

	conn, err := net.DialUDP("udp", nil, dstAddr)
	if err != nil {
		log.Fatal("DialUDP error:", err)
	}
	defer conn.Close()
	display := chooseMonitor()

	log.Println("📡 UDP Screen Stream started to", dstAddr)
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		img, err := screenshot.CaptureDisplay(display)
		if err != nil {
			log.Println("❌ Screenshot error:", err)
			continue
		}
		resized := resize.Resize(uint(*width), uint(*height), img, resize.Lanczos3)
		bounds := resized.Bounds()
		w := uint32(bounds.Dx())
		h := uint32(bounds.Dy())

		var buf bytes.Buffer
		err = webp.Encode(&buf, resized, &webp.Options{Quality: float32(*quality)})
		if err != nil {
			log.Println("❌ WebP encode error:", err)
			continue
		}

		frame := buf.Bytes()
		if len(frame)+8 > 65507 {
			log.Println("⚠️ Frame too large for UDP, skipped")
			continue
		}

		packet := new(bytes.Buffer)
		binary.Write(packet, binary.BigEndian, w)
		binary.Write(packet, binary.BigEndian, h)
		packet.Write(frame)
		conn.Write(packet.Bytes())
	}
}
