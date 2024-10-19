package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
	"strconv"
)

const (
	DEBUG         = false
	CTIMEOUT      = 100 * time.Second
	MAXLEN        = 1460
	CPUIDLEN      = 7
	MAXCLIENTS    = 1000
	VERSION_MAJOR = 1
	VERSION_MINOR = 0
	VERSION_PATCH = 0
	IP            = "::"
	DEFAULT_PORT  = 60050
)

type Client struct {
	cpuid     [CPUIDLEN]byte
	addr      net.UDPAddr
	lastTime  time.Time
}

var (
	clients   = make(map[[CPUIDLEN]byte]Client)
	clientsMu sync.Mutex
)

func findAndUpdateClient(cpuid [CPUIDLEN]byte, addr net.UDPAddr) (*Client, bool) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	client, exists := clients[cpuid]
	if exists {
		client.addr = addr
		client.lastTime = time.Now()
		clients[cpuid] = client
		return &client, true
	}
	return nil, false
}

func addClient(cpuid [CPUIDLEN]byte, addr net.UDPAddr) (*Client, bool) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	if len(clients) < MAXCLIENTS {
		now := time.Now()
		client := Client{
			cpuid:    cpuid,
			addr:     addr,
			lastTime: now,
		}
		clients[cpuid] = client
		return &client, true
	}
	return nil, false
}

func handlePacket(conn *net.UDPConn, buf []byte, addr net.UDPAddr) {
	if len(buf) < CPUIDLEN*2+6 || string(buf[:4]) != "NRL2" {
		return
	}

	var senderCpuid, receiverCpuid [CPUIDLEN]byte
	copy(senderCpuid[:], buf[6:6+CPUIDLEN])
	copy(receiverCpuid[:], buf[6+CPUIDLEN:6+CPUIDLEN*2])

	if client, _ := findAndUpdateClient(senderCpuid, addr); client == nil {
		if cpuid, _ := addClient(senderCpuid, addr); cpuid == nil {
			fmt.Println("[Info] Client full")
			return
		}
	}

	now := time.Now()
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for _, client := range clients {
		if now.Sub(client.lastTime) > CTIMEOUT {
			continue
		}

		if client.cpuid == receiverCpuid {
			_, err := conn.WriteTo(buf, &client.addr)
			if err != nil {
				fmt.Printf("[Error] Failed to send packet: %v\n", err)
			}
			if DEBUG {
				fmt.Printf("[Info] Packet forwarded to: %v\n", client.addr)
			}
			break
		}
	}
}

func main() {
	port := uint16(DEFAULT_PORT)
	if len(os.Args) > 2 && os.Args[1] == "-p" {
		if p, err := strconv.ParseUint(os.Args[2], 10, 16); err == nil {
			port = uint16(p)
		}
	}

	bindAddr := fmt.Sprintf("%s:%d", IP, port)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(IP), Port: int(port)})
	if err != nil {
		fmt.Printf("[Error] Couldn't bind to address: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Printf("[Info] NRL2730-Go %d.%d.%d server is running on %s\n", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH, bindAddr)

	buf := make([]byte, MAXLEN)

	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("[Error] Failed to read from UDP: %v\n", err)
			continue
		}
		if DEBUG {
			fmt.Printf("[Info] Received packet from: %v\n", addr)
		}
		handlePacket(conn, buf[:n], *addr)
	}
}
