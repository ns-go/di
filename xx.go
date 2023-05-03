package main

import (
	"fmt"
	"reflect"

	"github.com/ns-go/di/pkg/di"
)

type TT struct {
	tt1 TT2 `di.inject:""`
}

type TT2 struct {
	name string
}

func main() {
	t := reflect.TypeOf(new(TT)).Elem()
	xxx := reflect.New(t).Elem().Interface()
	yy := reflect.ValueOf(xxx).Elem()
	fmt.Println(xxx, yy)
	constainer := di.NewContainer()
	di.RegisterSingleton[TT](constainer, false)
	di.RegisterSingleton[TT2](constainer, false)
	xx, err := di.Resolve[TT](constainer)
	fmt.Println(xx, err)
}
