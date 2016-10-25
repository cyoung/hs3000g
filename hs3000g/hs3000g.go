package hs3000g

import (
	"errors"
)

/*
	SLIP_Parse().
	 - Assumes the 0xC0 start/end characters.
	 - Assumes a single message per call.
	 - 0xDBDC = 0xC0
	 - 0xDBDD = 0xDB
*/

func SLIP_Parse(msg []byte) ([]byte, error) {
	ret := make([]byte, 0)

	l := len(msg)
	for i := 0; i < l; i++ {
		if msg[i] == 0xC0 {
			continue // End character. Probably at the start or the end, since we assume only one message in 'msg'.
		}
		if msg[i] == 0xDB && i < l-1 {
			if msg[i+1] == 0xDC {
				ret = append(ret, byte(0xC0))
			} else if msg[i+1] == 0xDD {
				ret = append(ret, byte(0xDB))
			} else {
				return ret, errors.New("Invalid SLIP format.")
			}
			i = i + 2
		} else {
			ret = append(ret, msg[i])
		}
	}
	return ret, nil
}
