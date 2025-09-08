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

// --- Receiver side: handle incoming folder transfer ---
func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// Read metadata line
		meta, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Println("All files received ✅")
			return
		}
		if err != nil {
			fmt.Println("Error reading metadata:", err)
			return
		}

		meta = strings.TrimSpace(meta)
		if meta == "DONE" {
			fmt.Println("Transfer complete ✅")
			return
		}

		parts := strings.Split(meta, ":")
		if len(parts) != 2 {
			fmt.Println("Invalid metadata:", meta)
			continue
		}
		relPath := parts[0]
		fileSize, _ := strconv.Atoi(parts[1])

		// Make sure directories exist
		err = os.MkdirAll(filepath.Dir(relPath), os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directories:", err)
			return
		}

		// Create file
		outFile, err := os.Create(relPath)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}

		// Copy file content
		_, err = io.CopyN(outFile, reader, int64(fileSize))
		outFile.Close()
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}

		fmt.Println("Received:", relPath)
	}
}

// --- Sender side: walk folder and send files ---
func sendFolder(conn net.Conn, folder string) error {
	return filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil // skip dirs, only send files
		}

		// Open file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Get relative path
		relPath, _ := filepath.Rel(folder, path)

		// Send metadata
		fmt.Fprintf(conn, "%s:%d\n", relPath, info.Size())

		// Send file content
		_, err = io.Copy(conn, file)
		if err != nil {
			return err
		}

		fmt.Println("Sent:", relPath)
		return nil
	})
}

func main() {
	// --- Receiver mode ---
	go func() {
		ln, _ := net.Listen("tcp", ":9000")
		fmt.Println("Listening on :9000 ...")
		for {
			conn, _ := ln.Accept()
			go handleConnection(conn)
		}
	}()

	// --- Sender mode ---
	if len(os.Args) > 2 {
		addr := os.Args[1]   // e.g. "192.168.1.10:9000"
		folder := os.Args[2] // e.g. "./myfolder"

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			fmt.Println("Error connecting:", err)
			return
		}
		defer conn.Close()

		err = sendFolder(conn, folder)
		if err != nil {
			fmt.Println("Error sending folder:", err)
		}

		// Tell receiver we are done
		fmt.Fprintln(conn, "DONE")
	}
	select {}
}
