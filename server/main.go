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

// getLocalIP detects and returns the non-loopback local IP of the host
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		// Check if IP is IPV4 and not loopback
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no suitable local IP address found")
}

func chooseMonitor() int {
	// Get local IP address
	localIP, err := getLocalIP()
	if err != nil {
		log.Println("⚠️ Warning: Unable to auto-detect IP:", err)
		localIP = "127.0.0.1" // Fallback to localhost
	}

	fmt.Print("IP LOCAL YOU: ", localIP, "\n")
	displayCount := screenshot.NumActiveDisplays()
	fmt.Printf("🖥️ Available %d monitor:\n", displayCount)
	for i := 0; i < displayCount; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		fmt.Printf("[%d] %dx%d @ (%d,%d)\n", i, bounds.Dx(), bounds.Dy(), bounds.Min.X, bounds.Min.Y)
	}

	fmt.Print("Pilih monitor (nomor): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	index, err := strconv.Atoi(input)
	if err != nil || index < 0 || index >= displayCount {
		log.Fatalf("❌ Input monitor tidak valid.")
	}
	return index
}

func main() {
	quality := flag.Int("q", 40, "WebP quality (0-100)")
	width := flag.Int("w", 960, "Resize width")
	height := flag.Int("h", 540, "Resize height")
	port := flag.String("port", "8088", "Server port")
	flag.Parse()

	// Get local IP address
	localIP, err := getLocalIP()
	if err != nil {
		log.Println("⚠️ Warning: Unable to auto-detect IP:", err)
		localIP = "127.0.0.1" // Fallback to localhost
	}

	display := chooseMonitor()

	// Format port string for listening
	listenAddr := fmt.Sprintf(":%s", *port)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	// Display connection info with detected IP
	log.Println("🎥 Screen server started")
	log.Printf("📡 Server address: %s:%s\n", localIP, *port)
	log.Println("⏳ Waiting for client connection...")

	conn, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("✅ Client connected from:", conn.RemoteAddr())

	ticker := time.NewTicker(150 * time.Millisecond) // ~6 FPS stabil
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
		totalLen := uint32(len(frame) + 8) // 4 byte width + 4 byte height + frame
		binary.Write(conn, binary.BigEndian, totalLen)
		binary.Write(conn, binary.BigEndian, w)
		binary.Write(conn, binary.BigEndian, h)
		conn.Write(frame)
	}
}
