package main

import (
	"fmt"
	"github.com/wuciyou/owo"
)

func main() {
	fmt.Println("start ...")
	call("User.SayHello", "我的名字叫吴赐有")
	call("User.RegisterUser", "我的名字叫Wells")
}

func call(serviceMethod string, name string) {

	var age int
	err := owo.Call(serviceMethod, name, &age)
	if err != nil {
		fmt.Printf("Call fail error:%+v \n", err)
	} else {
		fmt.Printf("Call succsss \n age:%d \n", age)
	}
}
