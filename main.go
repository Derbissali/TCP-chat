package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Network struct {
	clients map[net.Conn]string
	sync.Mutex
	ch Channels
}

type Channels struct {
	ms        chan string
	conn      chan net.Conn
	dconn     chan net.Conn
	counterch chan string
}

const logo = `         _nnnn_
        dGGGGMMb
       @p~qp~~qMb
       M|@||@) M|
       @,----.JM|
      JS^\__/  qKL
     dZP        qKRb
    dZP          qKKb
   fZP            SMMb
   HZM            MMMM
   FqM            MMMM
 __| ".        |\dS"qML
 |     .       |  ' \Zq
_)      \.___.,|     .'
\____   )MMMMMP|   .'
	  -'        --'` + "\n"

// var counter = 0

func NewNetwork() *Network {
	return &Network{
		clients: make(map[net.Conn]string),
		ch: Channels{
			ms:        make(chan string),
			conn:      make(chan net.Conn),
			dconn:     make(chan net.Conn),
			counterch: make(chan string),
		},
	}
}

func main() {
	arg := os.Args[1:]
	port := "8989"
	if len(arg) != 0 {
		port = arg[0]
	}
	fmt.Println("Listening on the port :" + port)
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Println(err.Error())
		return
	}
	if err := os.Truncate("info.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}
	counter := 0
	chat := NewNetwork()

	go func() {
		for {
			if counter >= 10 {
				continue
			}
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err.Error())
				return
			}
			counter++
			fmt.Println(counter)
			go chat.handleClient(conn)

		}
	}()

	for {
		f, err := os.OpenFile("info.txt", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}
		select {
		case msg := <-chat.ch.ms:
			name := strings.Split(msg, ":")
			mas := ""
			for i := 0; i < len(msg); i++ {
				if msg[i] == ':' {
					mas = string(msg[i+2:])
				}
			}
			chat.Lock()
			for connn := range chat.clients {
				if name[0] != chat.clients[connn] {
					//connn.Write([]byte("\n"))
					connn.Write([]byte("\n" + timer() + "[" + name[0] + "]:" + mas + timer() + "[" + chat.clients[connn] + "]" + ":"))

				} else {
					connn.Write([]byte(timer() + "[" + name[0] + "]" + ":"))

					newLine := (timer() + "[" + chat.clients[connn] + "]:" + mas)
					_, err = fmt.Fprint(f, newLine)
					if err != nil {
						fmt.Println(err)
						f.Close()
						return
					}
					err = f.Close()
					if err != nil {
						fmt.Println(err)
						return
					}
				}

			}
			chat.Unlock()
		case conn := <-chat.ch.conn:
			chat.Lock()
			for connn := range chat.clients {
				if connn != conn {
					connn.Write([]byte("\n" + chat.clients[conn] + " has joined our chat..." + "\n" + timer() + "[" + chat.clients[connn] + "]" + ":"))
				}
			}
			newLine := (chat.clients[conn] + " has joined our chat..." + "\n")
			_, err = fmt.Fprint(f, newLine)
			if err != nil {
				fmt.Println(err)
				f.Close()
				return
			}
			err = f.Close()
			if err != nil {
				fmt.Println(err)
				return
			}
			chat.Unlock()
		case dconn := <-chat.ch.dconn:
			chat.Lock()
			for connn := range chat.clients {

				connn.Write([]byte("\n" + chat.clients[dconn] + " has left our chat...\n" + timer() + "[" + chat.clients[connn] + "]" + ":"))

			}

			newLine := (chat.clients[dconn] + " has left our chat..." + "\n")
			_, err = fmt.Fprint(f, newLine)
			if err != nil {
				fmt.Println(err)
				f.Close()
				return
			}
			err = f.Close()
			if err != nil {
				fmt.Println(err)
				return
			}

			delete(chat.clients, dconn)
			// case ddconn := <-ddconn:
			chat.Unlock()
		case counterch := <-chat.ch.counterch:
			counterch += "s"
			fmt.Println(counter)
			counter--
		}
	}
}

func timer() string {
	word := "[" + time.Now().Format("2006-01-02 15:04:05") + "]"
	return word
}

func (n *Network) enter(conn net.Conn) string {
	for {
		conn.Write([]byte("[ENTER YOUR NAME]:"))
		newconn := bufio.NewReader(conn)
		word, err := newconn.ReadString('\n')
		word = strings.TrimSuffix(string(word), "\n")

		if err != nil {
			log.Println(err)
			fmt.Println(err)
			return "s"
		}
		word = strings.Trim(word, " ")
		fmt.Println(word)
		if len(word) > 0 {
			if n.repeatName(word, conn) {
				n.Lock()
				n.clients[conn] = word
				n.Unlock()
				return word
			} else {
				conn.Write([]byte("Change name\n"))
			}
		}
	}

}
func (n *Network) repeatName(str string, conn net.Conn) bool {
	if len(n.clients) > 0 {
		for _, x := range n.clients {
			fmt.Println(x, str)
			if x == str {
				return false
			}
		}
	} else {
		n.clients[conn] = str
		fmt.Println("apwu")
		return true
	}

	return true
}
func (n *Network) handleClient(conn net.Conn) {
	conn.Write([]byte("Welcome to TCP-Chat!\n"))
	conn.Write([]byte(logo))
	word := n.enter(conn)
	fi, _ := ioutil.ReadFile("info.txt")
	if len(string(fi)) != 0 {
		fi1 := strings.TrimSuffix(string(fi), "\n")
		conn.Write([]byte(fi1 + "\n"))

	}

	fmt.Println("Joined to chat " + string(word))
	if len(word) > 0 {
		n.ch.conn <- conn

		conn.Write([]byte(timer() + "[" + n.clients[conn] + "]" + ":"))
		for {
			newconn := bufio.NewReader(conn)
			word, err := newconn.ReadString('\n')
			if err != nil {
				break
			}
			trim := strings.Trim(word, " ")
			if len(trim) > 1 {
				n.ch.ms <- fmt.Sprintf("%v: %v", n.clients[conn], word)
			} else {
				conn.Write([]byte("empty message\n"))
				conn.Write([]byte(timer() + "[" + n.clients[conn] + "]" + ":"))
			}
		}
		n.ch.dconn <- conn
	}
	n.ch.counterch <- "-"

}
