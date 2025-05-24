package main

import (
	"bytes"
	"fmt"
	"image/png"
	"net"
	"os"
	"time"

	"github.com/kbinani/screenshot"
)

func main() {
	// STEP 1: Pilih monitor
	var monitor int
	n := screenshot.NumActiveDisplays()
	fmt.Println("List Monitor:")
	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		fmt.Printf("[%d] - %v\n", i, bounds)
	}
	fmt.Print("Pilih nomor monitor untuk share: ")
	fmt.Scan(&monitor)

	// STEP 2: Listen TCP
	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	fmt.Println("Menunggu client connect di port 9090...")
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Accept error:", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("Client connected!")

	// STEP 3: Capture and send loop
	for {
		img, err := screenshot.CaptureDisplay(monitor)
		if err != nil {
			fmt.Println("Capture error:", err)
			continue
		}
		buf := new(bytes.Buffer)
		png.Encode(buf, img)
		data := buf.Bytes()

		// Send length (4 byte) + image data
		size := uint32(len(data))
		sizeBuf := []byte{
			byte(size >> 24),
			byte(size >> 16),
			byte(size >> 8),
			byte(size),
		}
		conn.Write(sizeBuf)
		conn.Write(data)
		time.Sleep(50 * time.Millisecond) // 20 FPS
	}
}
