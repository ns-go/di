package test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/ns-go/di/pkg/di"
)

type Service1 struct {
}

type Service2 struct {
}

type Service3 struct {
	service4 Service4 `di.inject:""`
}

type Service4 struct {
	service1 *Service1 `di.inject:"test"`
}

type Service5 struct {
	service1 *Service1 `di.inject:""`
}

func TestRegisterByName(t *testing.T) {
	constainer := di.NewContainer()
	err := di.RegisterByName(constainer, "test", Service1{}, true)

	if err != nil {
		t.Errorf(`RegisterByName(constainer, "test", Service1{}, true) = %v;  wont nil`, err)
	}

	err = di.RegisterByName(constainer, "test", Service1{}, true)

	if err == nil {
		t.Errorf(`RegisterByName(constainer, "test", Service1{}, true) = %v;  wont %v`, err, fmt.Errorf("item name '%s' is already registered", "test"))
	}

	err = di.RegisterByName(constainer, "test2", &Service1{}, true)
	if err == nil {
		t.Errorf(`RegisterByName(constainer, "test", Service1{}, true) = %v;  wont %v`, err, errors.New("cannot register type of pointer"))
	}
}

func TestRegisterByType(t *testing.T) {
	constainer := di.NewContainer()
	err := di.RegisterSingleton[Service1](constainer, true)

	if err != nil {
		t.Errorf(`RegisterSingleton[Service1](constainer, true) = %v;  wont nil`, err)
	}

	err = di.RegisterSingleton[Service2](constainer, true)

	if err != nil {
		t.Errorf(`RegisterSingleton[Service2](constainer, true) = %v;  wont nil`, err)
	}

	err = di.RegisterSingleton[Service2](constainer, true)

	if err == nil {
		t.Errorf(`RegisterSingleton[Service2](constainer, true) = %v;  wont %v`, err, fmt.Errorf("Type '%s' is already registered.", reflect.TypeOf(Service2{}).Name()))
	}

	err = di.RegisterSingleton[*Service2](constainer, true)

	if err == nil {
		t.Errorf(`RegisterSingleton[Service2](constainer, true) = %v;  wont %v`, err, errors.New("Cannot register type of pointer."))
	}
}

func TestResloveSingleton(t *testing.T) {
	constainer := di.NewContainer()
	di.RegisterSingleton[Service1](constainer, false)
	di.RegisterSingleton[Service2](constainer, false)

	s1, err := di.Resolve[Service1](constainer)
	if s1 == nil || err != nil {
		t.Errorf(`Resolve[Service1](constainer) = %v, %v; want %v, %v`, s1, err, Service1{}, "error")
	}

	s2, _ := di.Resolve[Service1](constainer)

	if s1 != s2 {
		t.Error("Singleton items not same value")
	}
}

func TestReslove(t *testing.T) {
	constainer := di.NewContainer()
	di.RegisterTransient[Service1](constainer, false)

	s1, err := di.Resolve[Service1](constainer)

	if s1 == nil || err != nil {
		t.Errorf(`Resolve[Service1](constainer) = %v, %v; want %v, %v`, s1, err, Service1{}, "error")
	}
}

func TestInjection(t *testing.T) {
	constainer := di.NewContainer()
	di.RegisterByName(constainer, "test", Service1{}, false)
	di.RegisterTransient[Service1](constainer, false)
	di.RegisterTransient[Service5](constainer, false)
	di.RegisterTransient[Service4](constainer, false)

	s5, err := di.Resolve[Service5](constainer)

	if s5 == nil || err != nil {
		t.Errorf(`Resolve[Service5](constainer) = %v, %v; want %v, %v`, s5, err, Service1{}, "error")
	}

	if s5.service1 == nil {
		t.Errorf("Injection usin tag 'di.inject\"\"' = %v, want %v", s5.service1, Service1{})
	}

	s4, err := di.Resolve[Service5](constainer)

	if s4 == nil || err != nil {
		t.Errorf(`Resolve[Service4](constainer) = %v, %v; want %v, %v`, s4, err, Service1{}, "error")
	}

	if s4.service1 == nil {
		t.Errorf("Injection usin tag 'di.inject\"test\"' = %v; want %v", s4.service1, Service1{})
	}
}

func TestScope(t *testing.T) {
	constainer := di.NewContainer()
	di.RegisterByName(constainer, "test", Service1{}, false)

	di.RegisterScoped[Service4](constainer, false)

	s4, err := di.Resolve[Service4](constainer)

	if s4 != nil || err == nil {
		t.Errorf("Resolve[Service4](constainer) = %v,%v; want %v,%v", s4, err, nil, errors.New("Cannot resolve scoped item with none scoped container."))
	}

	scope, err := constainer.NewScope()

	if scope == nil || err != nil {
		t.Errorf(`constainer.NewScope() = %v,%v; want %v, %v`, scope, err, di.Container{}, nil)
	}

	s4, err = di.Resolve[Service4](scope)

	if s4 == nil || err != nil {
		t.Errorf("Resolve[Service4](scope) = %v,%v; want %v,%v", s4, err, Service4{}, nil)
	}

	scope2, err := constainer.NewScope()

	if scope == nil || err != nil {
		t.Errorf(`constainer.NewScope() = %v,%v; want %v, %v`, scope2, err, di.Container{}, nil)
	}

	s4_2, err := di.Resolve[Service4](scope2)

	if s4_2 == nil || err != nil {
		t.Errorf("Resolve[Service4](scope2) = %v,%v; want %v,%v", s4, err, Service4{}, nil)
	}

	if s4 == s4_2 {
		t.Error("Difference scope must resolve not same value")
	}
}
