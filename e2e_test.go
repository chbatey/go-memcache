package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bradfitz/gomemcache/memcache"
	s "github.com/chbatey/go-memcache/memcache"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEndToEnd(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ms := s.New()
	ms.Start()

	mc := memcache.New("localhost:8080")
	err := mc.Set(&memcache.Item{Key: "pop", Value: []byte("aa")})
	assert.Nil(t, err)

	item, err := mc.Get("pop")
	assert.Nil(t, err)
	assert.NotNil(t, item)
}
