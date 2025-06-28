package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()
	log.Println("Client connected")

	cmd := exec.Command("./bin/api_wrapper", "wildberries-config.yaml")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Println("StdinPipe error:", err)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("StdoutPipe error:", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Println("Command start error:", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer stdin.Close()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read from websocket error:", err)
				break
			}
			if _, err := stdin.Write(message); err != nil {
				log.Println("Write to stdin error:", err)
				break
			}
			if _, err := stdin.Write([]byte("\n")); err != nil {
				log.Println("Write newline to stdin error:", err)
				break
			}
		}
	}()

	go func() {
		defer wg.Done()
		buf := make([]byte, 2048)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				log.Println("Read from stdout error:", err)
				break
			}
			if err := conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				log.Println("Write to websocket error:", err)
				break
			}
		}
	}()

	wg.Wait()
	log.Println("Client disconnected")
	cmd.Process.Kill()
}

func main() {
	// Build the original stdio-based application into the bin directory
	log.Println("Building api_wrapper...")
	if _, err := os.Stat("bin"); os.IsNotExist(err) {
		os.Mkdir("bin", 0755)
	}
	buildCmd := exec.Command("go", "build", "-o", "bin/api_wrapper", "./cmd/api_wrapper")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		log.Fatalf("Failed to build api_wrapper: %v", err)
	}
	log.Println("Build successful.")

	http.HandleFunc("/ws", handleWebSocket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	addr := ":" + port

	log.Printf("Starting WebSocket proxy on http://localhost%s/ws", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}
