package memcache

import (
	"bufio"
	"fmt"
	mc "github.com/bradfitz/gomemcache/memcache"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

var (
	a = 8080
)

func nextPort() int {
	a++
	return a
}

func newM() (Memcache, int) {
	port := nextPort()
	return New("localhost", port), port
}

func TestSetAndGet(t *testing.T) {
	m, port := newM()
	m.Start()
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	assert.Nil(t, err)
	defer conn.Close()
	defer m.Stop()

	fmt.Fprintf(conn, "set pop 1 0 2\r\n")
	conn.Write([]byte("aa"))
	fmt.Fprint(conn, "\r\n")

	bReader := bufio.NewReader(conn)
	status, _, err := bReader.ReadLine()
	assert.Nil(t, err)
	assert.Equal(t, "STORED", string(status))

	_, err = fmt.Fprint(conn, "gets pop\r\n")
	assert.Nil(t, err)
	getResponse, _, err := bReader.ReadLine()
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
}

func TestKeyNotFound(t *testing.T) {
	m, port := newM()
	assert.NoError(t, m.Start())
	client := mc.New(fmt.Sprintf("localhost:%d", port))
	defer m.Stop()

	item, err := client.Get("doesnotexist")
	assert.Nil(t, item)
	assert.Equal(t, mc.ErrCacheMiss, err)
}

func TestMultiKeyGet(t *testing.T) {
	m, port := newM()
	m.Start()
	client := mc.New(fmt.Sprintf("localhost:%d",port))
	defer m.Stop()

	assert.Nil(t, client.Set(&mc.Item{Key: "one", Value: []byte{1, 2}}))
	assert.Nil(t, client.Set(&mc.Item{Key: "two", Value: []byte{3, 4}}))

	results, err := client.GetMulti([]string{"one", "two", "three"})
	assert.NoError(t, err)
	assert.Equal(t, results["one"], &mc.Item{Key: "one", Value: []byte{1, 2}})
	assert.Equal(t, results["two"], &mc.Item{Key: "two", Value: []byte{3, 4}})
}
