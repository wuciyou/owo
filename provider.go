package owo

import (
	"fmt"
	"net/rpc"
	"reflect"
	"unicode"
	"unicode/utf8"
)

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

type providerNode struct {
	Id    string
	Name  string
	Addr  string
	Level int
}

type provider struct {
	addr string
}

func initProvider(addr string) *provider {
	return &provider{addr: addr}
}

func (this *provider) Register(rcvr interface{}, name ...string) ([]*providerNode, error) {
	serverType := reflect.TypeOf(rcvr)
	serverCvr := reflect.ValueOf(rcvr)
	sname := reflect.Indirect(serverCvr).Type().Name()

	if len(name) > 0 && name[0] != "" {
		if err := rpc.RegisterName(name[0], rcvr); err != nil {
			return nil, err
		}
		sname = name[0]
	} else {
		if err := rpc.Register(rcvr); err != nil {
			return nil, err
		}
	}

	methods := this.suitableMethods(serverType)
	var providerNodes []*providerNode

	for _, method := range methods {
		providerNodes = append(providerNodes, &providerNode{
			Name:  fmt.Sprintf("%s.%s", sname, method),
			Addr:  this.addr,
			Level: 10,
		})
	}

	return providerNodes, nil
}

func (this *provider) suitableMethods(typ reflect.Type) []string {
	var methods []string
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}
		// Method needs three ins: receiver, *args, *reply.
		if mtype.NumIn() != 3 {
			continue
		}
		// First arg need not be a pointer.
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			continue
		}
		// Second arg must be a pointer.
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			continue
		}
		// Method needs one out.
		if mtype.NumOut() != 1 {
			continue
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfError {
			continue
		}
		methods = append(methods, mname)
	}
	return methods
}

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}
