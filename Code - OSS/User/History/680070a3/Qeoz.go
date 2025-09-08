package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

// Handle incoming file from peer
func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// 1. Read metadata: "filename:size\n"
	meta, _ := reader.ReadString('\n')
	parts := strings.Split(strings.TrimSpace(meta), ":")
	if len(parts) != 2 {
		fmt.Println("Invalid metadata")
		return
	}
	filename := parts[0]
	filesize, _ := strconv.Atoi(parts[1])

	fmt.Printf("Receiving file: %s (%d bytes)\n", filename, filesize)

	// 2. Create file to save
	file, err := os.Create("received_" + filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// 3. Copy data from connection to file
	written, err := io.CopyN(file, reader, int64(filesize))
	if err != nil {
		fmt.Println("Error receiving file:", err)
		return
	}

	fmt.Printf("File received successfully (%d bytes)\n", written)
}

func main() {
	// 1. Start listener (server mode)
	go func() {
		ln, _ := net.Listen("tcp", ":9000")
		fmt.Println("Listening on :9000 for incoming files...")
		for {
			conn, _ := ln.Accept()
			go handleConnection(conn)
		}
	}()

	// 2. If args provided, send file (client mode)
	if len(os.Args) > 2 {
		addr := os.Args[1]
		filepath := os.Args[2]

		// Open file
		file, err := os.Open(filepath)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()

		// Get file info
		info, _ := file.Stat()
		filesize := info.Size()
		filename := info.Name()

		// Connect to peer
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			fmt.Println("Error connecting to peer:", err)
			return
		}
		defer conn.Close()

		// Send metadata
		fmt.Fprintf(conn, "%s:%d\n", filename, filesize)

		// Send file data
		written, err := io.Copy(conn, file)
		if err != nil {
			fmt.Println("Error sending file:", err)
			return
		}

		fmt.Printf("File sent successfully (%d bytes)\n", written)
	}

	select {} // keep running
}
