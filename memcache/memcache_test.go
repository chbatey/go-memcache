package memcache

import (
	"bufio"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	//	"time"
)

func TestSetAndGet(t *testing.T) {
	m := New()
	m.Start()
	conn, err := net.Dial("tcp", "localhost:8080")
	//	conn.SetReadDeadline(time.Now().Add(5000))
	assert.Nil(t, err)

	fmt.Fprintf(conn, "set pop 1 0 2\r\n")
	conn.Write([]byte("aa"))
	fmt.Fprint(conn, "\r\n")

	bReader := bufio.NewReader(conn)
	fmt.Println("Reading response to set")
	status, _, err := bReader.ReadLine()
	assert.Nil(t, err)
	assert.Equal(t, "STORED", string(status))

	fmt.Println("Sending Get")
	_, err = fmt.Fprint(conn, "gets pop\r\n")
	assert.Nil(t, err)
	fmt.Println("Reading response to get")
	getResponse, _, err := bReader.ReadLine()
	fmt.Println("Response to get is ", getResponse)
	assert.Nil(t, err)

	assert.Equal(t, "VALUE pop 1 2", string(getResponse))

	returnedData := make([]byte, 2, 2)
	length, err := bReader.Read(returnedData)
	assert.Nil(t, err)
	assert.Equal(t, length, 2)
	assert.Equal(t, returnedData, []byte("aa"))

	data, _, err := bReader.ReadLine()
	assert.Nil(t, err)
	assert.Equal(t, len(data), 0)

	end, _, err := bReader.ReadLine()
	assert.Nil(t, err)
	assert.Equal(t, []byte("END"), end)

	conn.Close()
	m.Stop()
}
