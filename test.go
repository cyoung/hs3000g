package main

import (
	"./hs3000g"
	"encoding/hex"
	"fmt"
)

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

}
