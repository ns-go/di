package main

import (
	"fmt"

	"github.com/ns-go/di/pkg/di"
)

type TT struct {
	tt2 *TT2 `di.inject:""`
}

type TT2 struct {
	name string
}

func main() {
	// x := TT2{name: "dg"}
	// f := reflect.ValueOf(&x).Elem().FieldByName("name")
	// f.SetString("sdrejo")
	// fmt.Print(x)
	// m := make(map[string]*string)
	// fmt.Println(m["edrt"])
	// t := reflect.TypeOf(new(TT)).Elem()
	// xxx := reflect.New(t).Elem().Interface()
	// yy := reflect.ValueOf(xxx).Interface()
	// fmt.Println(xxx, yy)
	constainer := di.NewContainer()
	di.RegisterSingleton[TT](constainer, false)
	di.RegisterSingleton[TT2](constainer, false)
	di.RegisterByName(constainer, "xx", TT2{name: "Test"}, false)
	xx, err := di.Resolve[TT](constainer)
	fmt.Println(xx, err)
}
