package modem

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"receiver/errors"
)

type Huawei struct {
	Modem
}

var UnexpectedResult error = errors.New("unexpected result")
var ErrorResult error = errors.New("error result")
var TimeoutError error = errors.New("timeout")
var InteractiveError error = errors.New("need intercative input")

func (h *Huawei) Connect() (err error) {
	if err = h.Modem.Connect(); err != nil {
		return
	}
	return
}

//ATZ
func (h *Huawei) Reset() (err error) {
	var lines [][]byte
	if lines, err = h.SendCommand("ATZ"); err != nil {
		return
	}
	if len(lines) != 1 {
		return UnexpectedResult
	}
	if !bytes.Equal(OK, lines[0]) {
		return ErrorResult
	}
	return
}

//ATE
func (h *Huawei) Echo(value bool) (err error) {
	var lines [][]byte
	var command string
	if value {
		command = "ATE1"
	} else {
		command = "ATE0"
	}
	if lines, err = h.SendCommand(command); err != nil {
		return
	}
	if len(lines) != 1 {
		return UnexpectedResult
	}
	if !bytes.Equal(OK, lines[0]) {
		return ErrorResult
	}
	return
}

//CURC
func (h *Huawei) ReportRssi(value bool) (err error) {
	var command string
	if value {
		command = "AT^CURC=10"
	} else {
		command = "AT^CURC=0"
	}

	var lines [][]byte
	if lines, err = h.SendCommand(command); err != nil {
		return
	}
	if len(lines) != 1 {
		return UnexpectedResult
	}
	if !bytes.Equal(OK, lines[0]) {
		return ErrorResult
	}
	return
}

//ATI
func (h *Huawei) Information() (result []string, err error) {
	var lines [][]byte
	if lines, err = h.SendCommand("ATI"); err != nil {
		return
	}
	if len(lines) == 0 {
		return nil, UnexpectedResult
	}
	if !bytes.Equal(OK, lines[len(lines)-1]) {
		return nil, ErrorResult
	}
	for i := 1; i < len(lines); i++ {
		result = append(result, string(lines[i]))
	}
	return
}

//CMGF
func (h *Huawei) MessageFormat(value int) (err error) {
	var lines [][]byte
	command := fmt.Sprintf("AT+CMGF=%v", value)
	if lines, err = h.SendCommand(command); err != nil {
		return
	}
	if len(lines) != 1 {
		return UnexpectedResult
	}
	if !bytes.Equal(OK, lines[0]) {
		return ErrorResult
	}
	return
}

//CMGS (Format = 0)
func (h *Huawei) PduMessage(pdu []byte) (err error) {
	p1 := fmt.Sprintf("AT+CMGS=%d", len(pdu))
	if _, err = h.SendCommand(p1); err != InteractiveError {
		return
	}

	var lines [][]byte
	if lines, err = h.Interactive(hex.EncodeToString(pdu)); err != nil {
		return
	}
	if len(lines) < 1 {
		return UnexpectedResult
	}
	if !bytes.Equal(OK, lines[len(lines)-1]) {
		return ErrorResult
	}
	return
}
