package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type stella struct {
	addr string
	conn net.Conn
}

var links = make(map[int]stella)

func printConnections() {

	if len(links) == 0 {
		fmt.Println("No Sessions YET :P")
	}

	for k, v := range links {
		fmt.Printf("\nSessions\n---------------------------\n")
		fmt.Printf("Session #%v: %v\n", k, v.addr)
	}
}

func help() {
	fmt.Println("\nAvailable Commands:")
	fmt.Println("---------------------------")
	fmt.Println("sessions         List active connections")
	fmt.Println("use <number>  Open shell in specified session")
	fmt.Println("kill <number>  Kill the specified session")
	fmt.Println("help            Show this help menu")
	fmt.Println("exit            Exit the program")
	fmt.Println("---------------------------")
}

func monitorConnection(id int) {
	buffer := make([]byte, 1)
	for {
		_, err := links[id].conn.Read(buffer)
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "connection reset by peer") {
				log.Printf("\nClient %d disconnected from %s", id, links[id].addr)
			} else {
				log.Printf("\nError reading from client %d: %v", id, err)
			}

			// Clean up the connection
			links[id].conn.Close()
			delete(links, id)
			return
		}
	}
}

func killSession(i int) error {
	err := links[i].conn.Close()
	if err != nil {
		return err
	}
	delete(links, i)
	return nil
}

func menuPrompt() {
	reader := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\033[H\033[2J")
		help()
		fmt.Print("> ")

		if !reader.Scan() {
			log.Println("Error reading input")
			return
		}

		inp := strings.TrimSpace(reader.Text())
		if inp == "" {
			fmt.Println("No input received. Please provide a valid command.")
			continue
		}

		if inp == "sessions" {
			printConnections()
			fmt.Print("\nPress Enter to continue...")
			reader.Scan() // Wait for Enter
		} else if inp == "help" {
			help()
		} else if inp == "exit" {
			log.Println("chill ok.. chill.")
		} else if strings.Contains(inp, "use") {
			if len(links) == 0 {
				fmt.Println("Invalid session.")
				continue
			}
			_, after, found := strings.Cut(inp, " ")
			// log.Printf("before %v\nafter %v\n", before, after)
			if !found {
				log.Println("Invalid input, please enter a valid command")
				continue
			}
			conv, err := strconv.Atoi(after)
			if err != nil {
				log.Println("Error converting integer to string: ", err)
				return
			}

			if len(links) < conv {
				fmt.Println("Session doesn't exist YET :P")
				continue
			}

			startShell(links[conv].conn)

		} else if strings.Contains(inp, "kill") {
			if len(links) == 0 {
				fmt.Println("No sessions")
				continue
			}
			_, after, found := strings.Cut(inp, " ")
			// log.Printf("before %v\nafter %v\n", before, after)
			if !found {
				log.Println("Invalid input, please enter a valid command")
				continue
			}
			conv, err := strconv.Atoi(after)
			if err != nil {
				log.Println("Error converting integer to string: ", err)
				return
			}

			if len(links) < conv {
				fmt.Println("Session doesn't exist YET ;)")
				continue
			}

			err = killSession(conv)
			if err != nil {
				log.Println("Kill Session function errored")
			}
			log.Println("Killed session # ", conv)
		} else {
			log.Println("Invalid Command")
		}
	}
}

func startShell(conn net.Conn) {
	go func() {
		_, err := io.Copy(os.Stdout, conn)
		if err != nil {
			log.Printf("\nError reading from connection: %v\n", err)
		}
	}()

	log.Println("Opened shell")

	// Read from standard input and send to the connection.
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if scanner.Scan() {
			line := scanner.Text()
			if line == "exit" {
				log.Println("Closing shell..")
				return
			}
			_, err := conn.Write([]byte(line + "\n"))
			if err != nil {
				log.Printf("\nError writing to connection: %v\n", err)
				return
			}
		} else {
			break // end of input
		}
	}
}

func startTCPServer(host string, port string) error {
	address := host + ":" + port
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("TCP Server listening on %s", address)
	go menuPrompt()

	connID := 1
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("\nFailed to accept connection: %v\n", err)
			continue
		}
		links[connID] = stella{
			addr: conn.RemoteAddr().String(),
			conn: conn,
		}

		go monitorConnection(connID)
		connID++
		log.Printf("\nNew connection from %v", conn.RemoteAddr())
	}
}

func main() {
	if err := startTCPServer("localhost", "8080"); err != nil {
		log.Fatal(err)
	}
}
