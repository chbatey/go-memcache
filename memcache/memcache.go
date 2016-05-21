package memcache

import (
	"bufio"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net"
	"strconv"
	"strings"
)

type Memcache interface {
	Start() error
	Stop() error
	WaitFor()
}

type entry struct {
	data   []byte
	flags  int
	ttl    int
	length int
}

type memcache struct {
	data           map[string]entry
	listen_address string
	listen_port    int
	closed         chan string
	listener       net.Listener
}

func New() Memcache {
	return &memcache{
		data:           make(map[string]entry),
		listen_address: "localhost",
		listen_port:    8080,
		closed:         make(chan string),
	}
}

func (m *memcache) WaitFor() {
	<-m.closed
}

func (m *memcache) Start() error {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", m.listen_address, m.listen_port))
	if err != nil {
		log.Warn("Unable to listen on host and port", err)
		return err
	}
	log.Debugf("Setting listener to ", ln)
	m.listener = ln
	go m.process()
	return nil
}

// TODO notify that the client has disconnected
func (m *memcache) process() {
	for {
		conn, err := m.listener.Accept()
		if err != nil {
			log.Info("Connection closed", err)
			return
		}
		go m.handleConnection(conn)
	}
}

func (m *memcache) handleConnection(conn net.Conn) {
	log.Info("Handling connection", conn)
	defer conn.Close()
	bReader := bufio.NewReader(conn)

	for {
		data, _, err := bReader.ReadLine()
		if err != nil {
			log.Debugf("Error reading from socket. Closing connection.", err)
			return
		}
		lines := strings.Split(string(data), " ")
		// TODO deal with partial commands
		if len(lines) < 2 {
			continue
		}
		command := lines[0]
		if command == "set" {
			key := lines[1]
			log.Debug("Processing set")
			flags, _ := strconv.Atoi(lines[2])
			ttl, _ := strconv.Atoi(lines[3])
			length, _ := strconv.Atoi(lines[4])

			log.Debug(lines, command, key, flags, ttl, length)
			if err != nil {
				log.Warnf("Error reading from socket {}", err)
				return
			}

			data := make([]byte, length, length)
			bReader.Read(data)
			log.Debug("Read data ", data)

			entry := entry{
				data:   data,
				flags:  flags,
				ttl:    ttl,
				length: length,
			}

			m.data[key] = entry

			bReader.ReadLine()
			log.Debug("Writing response to set")
			conn.Write([]byte("STORED\r\n"))
			log.Debug("Finished storing message")
		} else if command == "gets" || command == "get" {
			keys := lines[1:]
			log.Infof("Processing get for %s keys", keys)
			for _, key := range keys {
				val, ok := m.data[key]
				if ok {
					log.Debug("Found key %s", val)
					response := []byte(fmt.Sprintf("VALUE %s %d %d\r\n", key, val.flags, val.length))
					fullResponse := append(response, val.data...)
					fullResponse = append(fullResponse, []byte("\r\n")...)
					_, err = conn.Write(fullResponse)
					if err != nil {
						log.Warn("Failed to write response back", err)
					}
				}
			}
			conn.Write([]byte("END\r\n"))
		}
	}
}

func (m *memcache) Stop() error {
	log.Debugf("Closing down listener", m.listener)
	if m.listener != nil {
		m.listener.Close()
	}
	close(m.closed)
	return nil
}
