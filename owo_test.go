package owo

import (
	"net"
	"testing"
)

type User struct {
}

func (this *User) SayHello(name string, age *int) error {

	return nil
}

func (this *User) RegisterUser(name string, age *int) error {

	return nil
}
func _TestApp(t *testing.T) {

	user := &User{}
	Register(user)
	Run()
}

func TestAddr(t *testing.T) {

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	t.Logf("addrs:%+v \n", addrs)

}
