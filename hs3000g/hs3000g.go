package hs3000g

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dgryski/go-bitstream"
	"io"
)

/*
	SLIP_Parse().
	 - Assumes the 0xC0 start/end characters.
	 - Assumes a single message per call.
	 - 0xDBDC = 0xC0
	 - 0xDBDD = 0xDB
*/

const (
	POSITION     = 0xAA00
	POSITION_RSP = 0xFF01
	LOGIN_REQ    = 0xAA02
	LOGIN_RSP    = 0xFF03
	SET_REQ      = 0xAA04
	SET_RSP      = 0xFF05
	UPGRADE_REQ  = 0xAA06
	UPGRADE_RSP  = 0xFF07
	DOWN_REQ     = 0xAA08
	DOWN_RSP     = 0xFF09
	UPFAULT_REQ  = 0xAA12
	UPFAULT_RSP  = 0xFF13
	HSO_REQ      = 0xAA14
	HSO_RSP      = 0xFF15
	CTRL_REQ     = 0xAA16
	CTRL_RSP     = 0xFF17
)

func SLIP_Decode(msg []byte) ([]byte, error) {
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

func SLIP_Encode(msg []byte) []byte {
	ret := []byte{0xC0} // START.

	for i := 0; i < len(msg); i++ {
		if msg[i] == 0xC0 {
			ret = append(ret, byte(0xDB))
			ret = append(ret, byte(0xDC))
		} else if msg[i] == 0xDB {
			ret = append(ret, byte(0xDB))
			ret = append(ret, byte(0xDD))
		} else {
			ret = append(ret, msg[i])
		}
	}

	ret = append(ret, byte(0xC0)) // END.
	return ret
}

type HSMessage struct {
	bitstream    *bitstream.BitReader
	reader       io.Reader
	Command      uint32
	SerialNumber uint32
	SIM_ID       string
	VIN          string
	Device_Type  int
	Response     []byte
}

func (m *HSMessage) constructResponse(msg []byte) {
	var b bytes.Buffer
	responseWriter := bitstream.NewWriter(&b)
	responseWriter.WriteBits(0x0100, 16) //FIXME: Why?

	// Packet length.
	pktLen := 12 + len(msg) // Header is 12 bytes.
	responseWriter.WriteBits(uint64(pktLen), 16)

	// Command.
	cmdResp := uint64(0)
	switch m.Command {
	//	case POSITION:
	//		cmdResp = POSITION_RSP
	case LOGIN_REQ:
		cmdResp = LOGIN_RSP
	}

	if cmdResp == 0 {
		// Don't know how to respond to this command. Quit.
		return
	}
	responseWriter.WriteBits(cmdResp, 16)

	// CRC16. Not used.
	responseWriter.WriteBits(0, 16)

	// Serial number.
	responseWriter.WriteBits(uint64(m.SerialNumber), 32)

	// Write the content of the packet.
	for i := 0; i < len(msg); i++ {
		responseWriter.WriteByte(msg[i])
	}

	responseWriter.Flush(false)

	m.Response = SLIP_Encode(b.Bytes())
	fmt.Printf("response: %s\n", hex.Dump(m.Response))
}

func (m *HSMessage) parseLoginMessage() {
	for {
		fieldType, err := m.bitstream.ReadBits(16)
		if err != nil {
			break
		}
		fieldLen, err := m.bitstream.ReadBits(16)
		if err != nil {
			break
		}
		fieldLen = fieldLen - 4 // 'fieldLen' units are in bytes. 4 bytes used by the type and length fields.
		if fieldLen <= 0 {
			continue
		}

		// Read the value of the field into a []byte.
		fieldVal := make([]byte, fieldLen)
		readErr := false
		for i := 0; i < int(fieldLen); i++ {
			v, err := m.bitstream.ReadByte()
			if err != nil {
				readErr = true
				break
			}
			fieldVal[i] = v
		}
		if readErr {
			break // Don't use the values read.
		}
		switch fieldType {
		case 0x0001: // Firmware version.
			fmt.Printf("Firmware version=%s\n", string(fieldVal))
		case 0x0002: // VIN number.
			fmt.Printf("VIN number=%s\n", string(fieldVal))
			m.VIN = string(fieldVal)
		case 0x0003: // IMEI.
			fmt.Printf("IMEI=%s\n", string(fieldVal))
			m.SIM_ID = string(fieldVal)
		case 0x0008: // Device type.
			if len(fieldVal) == 0 {
				break
			}
			fmt.Printf("Device type=%d\n", fieldVal[0])
			m.Device_Type = int(fieldVal[0])
		}
	}

	//FIXME: Need some checking to declare a success, but for now always responding "successful login".
	m.constructResponse([]byte{0})
}

func (m *HSMessage) parseMessage() error {
	// "flag" CRC bit, etc. Ignored for now.
	flag, err := m.bitstream.ReadBits(4)
	if err != nil {
		return err
	}

	// Protocol version. Should always be 0.
	vers, err := m.bitstream.ReadBits(4)
	if err != nil {
		return err
	}

	// "Reserved" - unused.
	reserved, err := m.bitstream.ReadBits(8)
	if err != nil {
		return err
	}

	// Total length of the packet.
	packetLength, err := m.bitstream.ReadBits(16)
	if err != nil {
		return err
	}

	// Command identifier.
	cmd, err := m.bitstream.ReadBits(16)
	if err != nil {
		return err
	}

	// CRC16 of the packet.
	packetCRC, err := m.bitstream.ReadBits(16)
	if err != nil {
		return err
	}

	// Serial number of the packet.
	serialNum, err := m.bitstream.ReadBits(32)
	if err != nil {
		return err
	}
	m.SerialNumber = uint32(serialNum)

	fmt.Printf("flag=%02x, vers=%02x, reserved=%02x, packetLength=%04x, cmd=%04x, packetCRC=%04x, serialNum=%08x\n", flag, vers, reserved, packetLength, cmd, packetCRC, serialNum)

	m.Command = uint32(cmd)
	// Now do message-specific parsing.
	switch cmd {
	case LOGIN_REQ:
		m.parseLoginMessage()
	case POSITION:

	default:
		fmt.Printf("unknown message type=%04x\n", cmd)
	}

	return nil
}

func NewMessage(m []byte) (*HSMessage, error) {
	ret := new(HSMessage)

	// Parse SLIP.
	parsedMsg, err := SLIP_Decode(m)
	if err != nil {
		return ret, err
	}

	ret.reader = bytes.NewReader(parsedMsg)
	if err != nil {
		return ret, err
	}
	ret.bitstream = bitstream.NewReader(ret.reader)

	ret.parseMessage()
	return ret, nil
}
