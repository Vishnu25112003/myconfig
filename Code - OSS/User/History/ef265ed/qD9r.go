package main

import "os"

func main() {
	listener, err := net.listener("tcp", ":8080")
	if err != nill {
		fmt.println("Error starting server:", err)
		os.Exit(1)
	}

}
