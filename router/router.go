package router

import (
	"fmt"
	"reflect"

	"github.com/go-router/router"
)

var r router.Router

func init() {
	r = router.New(router.PathID(), -1, router.BroadcastPolicy)
}

func Subscribe(path string, ch interface{}) (*router.RoutedChan, error) {
	return r.AttachRecvChan(router.PathID(path), ch)
}

func PermanentSubscribe(path string, ch interface{}) (*router.RoutedChan, error) {
	return r.AttachRecvChan(router.PathID(path), ch, make(chan *router.BindEvent, 1))
}

func Connect(path string, ch interface{}) (*router.RoutedChan, error) {
	return r.AttachSendChan(router.PathID(path), ch)
}

func Close(ch *router.RoutedChan) {
	ch.Close()
}

func Send(path string, ch interface{}, val interface{}) error {
	if reflect.TypeOf(ch).Kind() != reflect.Chan {
		return fmt.Errorf("passed %+v instead of a chan", reflect.TypeOf(ch))
	}
	rch, err := Connect(path, ch)
	if err != nil {
		return err
	}
	reflect.ValueOf(ch).Send(reflect.ValueOf(val))
	Close(rch)
	return nil
}
