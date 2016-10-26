package main

import (
	"./hs3000g"
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

	z, err := hex.DecodeString("0000004baa0000000000002cdbdc000000313631303236313532313435ff8057ca0040d078000000000116000000010015ffffff0000000000000004a3ff58113a1c0005000a0306000002cbc0")
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		return
	}

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
