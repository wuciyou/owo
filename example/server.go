package main

import (
	"github.com/wuciyou/dogo"
	"github.com/wuciyou/owo"
)

type User struct {
}

func (this *User) SayHello(name string, age *int) error {

	dogo.Dglog.Infof("Call SayHello success ... \n name:%s ", name)
	*age = 50
	return nil
}

func (this *User) RegisterUser(name string, age *int) error {
	dogo.Dglog.Infof("Call RegisterUser success... \n name:%s", name)
	*age = 100
	return nil
}
func main() {
	user := &User{}
	owo.Register(user)
	owo.Run()
}
