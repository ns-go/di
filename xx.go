package main

import (
	"fmt"

	"github.com/ns-go/di/pkg/di"
)

type TT struct {
	name TT2 `di.inject:""`
}

type TT2 struct {
	name TT `di.inject:""`
}

func main() {
	constainer := di.NewContainer()
	di.RegisterScoped[TT](constainer, true)
	xx, err := di.Resolve[TT](constainer)
	fmt.Println(xx, err)
}
