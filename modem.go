package modem

import (
	"bytes"
	"github.com/tarm/serial"
	"log"
	"sync"
	"time"
)

var OK []byte = []byte("OK")
var ERROR []byte = []byte("ERROR")
var NewLine []byte = []byte{byte('\r'), byte('\n')}
var Escape []byte = []byte{byte('\n'), byte(26)}
var Interactive []byte = []byte{62, 32}

type CommandResult struct {
	Output [][]byte
	Error  error
}

type Modem struct {
	sync.RWMutex
	sync.WaitGroup
	Name    string
	Timeout time.Duration

	bufferSize int
	port       *serial.Port
	cmd        []byte
	ans        chan [][]byte
	exit       chan struct{}
}

func (m *Modem) Connect() (err error) {
	c := &serial.Config{Name: m.Name, Baud: 115200, ReadTimeout: time.Second * 1}
	if m.port, err = serial.OpenPort(c); err != nil {
		return
	}
	m.bufferSize = 80
	m.ans = make(chan [][]byte)
	m.exit = make(chan struct{})
	m.Add(1)
	go m.HandleRead()
	return
}

func (m *Modem) Disconnect() (err error) {
	m.exit <- struct{}{}
	m.Wait()
	err = m.port.Close()
	return
}

func (m *Modem) SendCommand(cmd string) (result [][]byte, err error) {
	m.Lock()
	defer m.Unlock()

	m.cmd = []byte(cmd)
	if _, err = m.port.Write([]byte(cmd)); err != nil {
		return
	}
	if _, err = m.port.Write(NewLine); err != nil {
		return
	}
	select {
	case <-time.After(m.Timeout):
		err = TimeoutError
	case result = <-m.ans:
		if result == nil {
			err = InteractiveError
		}
	}
	return
}

func (m *Modem) Interactive(msg string) (result [][]byte, err error) {
	m.Lock()
	defer m.Unlock()

	if _, err = m.port.Write([]byte(msg)); err != nil {
		return
	}
	if _, err = m.port.Write(Escape); err != nil {
		return
	}
	select {
	case <-time.After(m.Timeout):
		err = TimeoutError
	case result = <-m.ans:
		if result == nil {
			err = ErrorResult
		}
	}
	return
}

func (m *Modem) HandleRead() {
	defer m.Done()

	var isOutput bool
	var output [][]byte
loop:
	for {
		select {
		case <-m.exit:
			break loop
		default:
			if lines, err := m.Read(); err != nil {
				continue
			} else {
				if len(lines) == 0 {
					continue
				}

				cmd := m.cmd
				if cmd == nil {
					for _, line := range lines {
						log.Printf("verbose: %v", string(line))
					}
					continue
				}

				for _, line := range lines {
					log.Printf("scan(%v): %v (%v)", string(cmd), string(line), len(line))
					switch {
					case bytes.Equal(line, cmd):
						isOutput = true
					case isOutput && bytes.Equal(line, OK):
						output = append(output, line)
						isOutput = false
					case isOutput && bytes.Equal(line, ERROR):
						output = append(output, line)
						isOutput = false
					case isOutput && bytes.Equal(line[:4], []byte("+CMS")):
						output = append(output, line)
						isOutput = false
					case isOutput && bytes.Equal(line, Interactive):
						m.ans <- nil
					case isOutput:
						output = append(output, line)
					default:
					}
				}
				if !isOutput && output != nil {
					m.ans <- output
					m.cmd = nil
					output = nil
				}
			}
		}
	}
}

func (m *Modem) Read() (result [][]byte, err error) {
	var n int
	var wordFound, complete bool
	var partialResult bool = true
	var partialResultIndex int = 0
	for !complete {
		buf := make([]byte, m.bufferSize)
		if n, err = m.port.Read(buf); err != nil || n == 0 {
			if err != nil {
				log.Printf("handle: read %v", err)
			}
			return
		}
		complete = n != m.bufferSize
		var startIndex int
		for i := 0; i < n; i++ {
			for i < n && (buf[i] == '\r' || buf[i] == '\n') {
				if wordFound {
					if partialResult && partialResultIndex > 0 {
						for _, val := range buf[startIndex:i] {
							result[partialResultIndex-1] = append(result[partialResultIndex-1], val)
						}
					} else {
						result = append(result, buf[startIndex:i])
						partialResultIndex++
					}
					wordFound = false
					partialResult = true
				} else {
					startIndex = i + 1
					partialResult = false
				}
				i++
			}
			if !wordFound && i < n {
				wordFound = true
			}
		}
		if wordFound {
			if partialResult && partialResultIndex > 0 {
				for _, val := range buf[startIndex:n] {
					result[partialResultIndex-1] = append(result[partialResultIndex-1], val)
				}
			} else {
				result = append(result, buf[startIndex:n])
				partialResultIndex++
				partialResult = true
			}
			wordFound = false
		}
	}
	return
}
