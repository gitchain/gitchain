package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscribeConnect(t *testing.T) {
	rch := make(chan string)
	sch := make(chan string)
	if _, err := Subscribe("/test/*", rch); err != nil {
		t.Errorf("can't subscribe to /test/* because of %v", err)
	}
	Connect("/test/this", sch)
	sch <- "test"
	msg := <-rch
	assert.Equal(t, msg, "test")
}

func TestSubscribeSend(t *testing.T) {
	rch := make(chan string)
	if _, err := Subscribe("/test/*", rch); err != nil {
		t.Errorf("can't subscribe to /test/* because of %v", err)
	}
	Send("/test/this", make(chan string), "test")
	msg := <-rch
	assert.Equal(t, msg, "test")
}
