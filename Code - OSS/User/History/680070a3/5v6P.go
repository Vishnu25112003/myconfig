package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ================== RECEIVER ==================
func handleConnection(conn net.Conn, outputDir string) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// 1. Read metadata line: "relative/path:size\n"
		meta, err := reader.ReadString('\n')
		if err == io.EOF {
			break // transfer finished
		}
		if err != nil {
			fmt.Println("Error reading metadata:", err)
			return
		}

		parts := strings.Split(strings.TrimSpace(meta), ":")
		if len(parts) != 2 {
			fmt.Println("Invalid metadata:", meta)
			return
		}
		relpath := parts[0]
		size, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			fmt.Println("Invalid file size:", err)
			return
		}

		// 2. Ensure directory exists
		fullpath := filepath.Join(outputDir, relpath)
		if err := os.MkdirAll(filepath.Dir(fullpath), 0755); err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}

		// 3. Create file
		file, err := os.Create(fullpath)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}

		// 4. Copy file content
		written, err := io.CopyN(file, reader, size)
		file.Close()
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}

		fmt.Printf("Received: %s (%d bytes)\n", relpath, written)
	}

	fmt.Println("✅ Folder received successfully!")
}

func startServer(outputDir string) {
	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println("📡 Listening on :9000 ...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn, outputDir)
	}
}

// ================== SENDER ==================
func sendFolder(conn net.Conn, folder string) error {
	return filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files or folders (like .git, .DS_Store)
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil // skip directories, we recreate on receiver
		}

		rel, _ := filepath.Rel(folder, path)
		size := info.Size()

		// Send metadata
		fmt.Fprintf(conn, "%s:%d\n", rel, size)

		// Send file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(conn, file)
		if err != nil {
			return err
		}

		fmt.Println("Sent:", rel, "(", size, "bytes )")
		return nil
	})
}

func sendToPeer(addr, folder string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("❌ Error connecting to peer:", err)
		return
	}
	defer conn.Close()

	fmt.Println("🚀 Sending folder:", folder)
	err = sendFolder(conn, folder)
	if err != nil {
		fmt.Println("❌ Error sending folder:", err)
		return
	}
	fmt.Println("✅ Folder sent successfully!")
}

// ================== MAIN ==================
func main() {
	// If arguments are provided → send mode
	if len(os.Args) == 3 {
		addr := os.Args[1]   // "192.168.x.x:9000"
		folder := os.Args[2] // folder to send
		sendToPeer(addr, folder)
		return
	}

	// Otherwise → receive mode
	outputDir := "received_folder"
	startServer(outputDir)
}
