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
	z, err := hex.DecodeString("C00000007EAA020000000000010001001047315F48312E305F56312E3000030013383632393530303238353334333036000400144C342D56374C673979497A7A2D724A6D0005000501000600084341524400070008434152440008000500000900183839383630303530313931343436313130393134000A0009434D4E4554C0")
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		return
	}

	_, err = hs3000g.NewMessage(z)
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		return
	}

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
