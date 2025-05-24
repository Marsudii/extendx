package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image/png"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "SERVER_IP:9090")
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for i := 0; ; i++ {
		// Read size (4 byte)
		sizeBuf := make([]byte, 4)
		_, err := reader.Read(sizeBuf)
		if err != nil {
			fmt.Println("Read size error:", err)
			return
		}
		size := int(sizeBuf[0])<<24 | int(sizeBuf[1])<<16 | int(sizeBuf[2])<<8 | int(sizeBuf[3])

		// Read image
		imgBuf := make([]byte, size)
		_, err = reader.Read(imgBuf)
		if err != nil {
			fmt.Println("Read image error:", err)
			return
		}
		img, err := png.Decode(bytes.NewReader(imgBuf))
		if err != nil {
			fmt.Println("Decode error:", err)
			continue
		}

		// Simpan file (untuk tes; bisa diganti render ke window)
		fileName := fmt.Sprintf("frame_%03d.png", i)
		file, _ := os.Create(fileName)
		png.Encode(file, img)
		file.Close()
		fmt.Println("Frame saved:", fileName)
	}
}
