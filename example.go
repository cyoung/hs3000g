package main

import (
	"github.com/cyoung/hs3000g"
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	defer func() {
		fmt.Printf("connection closed: %s\n", conn.RemoteAddr().String())
		conn.Close()
	}()

	fmt.Printf("new connection! %s\n", conn.RemoteAddr().String())

	for {
		b, err := reader.ReadBytes(0xC0)
		if err != nil {
			return
		}

		parsedMsg, err := hs3000g.NewMessage(b)
		if err != nil {
			fmt.Printf("Err: %s\n", err.Error())
			continue
		}
		if len(parsedMsg.Response) > 0 {
			conn.Write(parsedMsg.Response)
		}
	}
}

func main() {

	z, err := hex.DecodeString("0000004baa00000000000589a5a5a55a3f0080000000313631303236323334343539ff8057cf0040d07400000000011b000000010015ffffff0000000000000004aeff5811a9bd0005000ac0")
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		return
	}

	fmt.Printf("%d\n", len(z))

	_, err = hs3000g.NewMessage(z)
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		return
	}
	return

	ln, err := net.Listen("tcp", ":8006")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}

}
