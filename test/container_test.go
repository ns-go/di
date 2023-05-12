package test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/ns-go/di/pkg/di"
)

type Service1 struct {
	id int
}

type Service2 struct {
}

type Service3 struct {
	service4 Service4 `di.inject:""`
}

type Service4 struct {
	service1 *Service1 `di.inject:"test"`
	id       int
}

type Service5 struct {
	service1 *Service1 `di.inject:""`
	id       int
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
	di.RegisterTransient[Service3](constainer, false)

	s1, err := di.Resolve[Service1](constainer)

	if s1 == nil || err != nil {
		t.Errorf(`Resolve[Service1](constainer) = %v, %v; want %v, %v`, s1, err, Service1{}, nil)
	}

	s3, err := di.Resolve[Service3](constainer)
	if s3 != nil || err == nil {
		t.Errorf(`Resolve[Service3](constainer) = %v, %v; want %v, %v`, s1, err, nil, "error")
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
	s4.id = 100

	if s4 == nil || err != nil {
		t.Errorf("Resolve[Service4](scope) = %v,%v; want %v,%v", s4, err, Service4{}, nil)
	}

	s42, err := di.Resolve[Service4](scope)
	if s4.id != s42.id || err != nil {
		t.Errorf("s42.id = %v,%v; want %v,%v", s42.id, err, s4.id, nil)
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

	s1, err := di.ResolveByName[Service1](constainer, "test")
	if s1 == nil || err != nil {
		t.Errorf("ResolveByName[Service1](constainer, \"test\") = %v,%v; want %v,%v", s1, err, Service1{}, nil)
	} else {
		s1.id = 100

		s12, _ := di.ResolveByName[Service1](constainer, "test")
		if s12.id != 100 {
			t.Errorf("s12.id = %v; want %v", s12.id, 100)
		}
	}
}

func TestFactory(t *testing.T) {
	constainer := di.NewContainer()
	di.RegisterFactory(constainer, di.Singleton, func(c di.Container) Service1 { return Service1{} }, false)
	di.RegisterTransient[Service5](constainer, false)
	s5, err := di.Resolve[Service5](constainer)
	if s5 == nil || err != nil {
		t.Errorf("Resolve[Service4](constainer) = %v,%v; want %v,%v", s5, err, Service4{}, nil)
	}

	s1, err := di.Resolve[Service1](constainer)
	if s1 == nil || err != nil {
		t.Errorf("Resolve[Service1](constainer) = %v,%v; want %v,%v", s1, err, Service1{}, nil)
	} else {
		s1.id = 100

		s12, _ := di.Resolve[Service1](constainer)
		if s12.id != 100 {
			t.Errorf("s12.id = %v; want %v", s12.id, 100)
		}
	}
}

func TestRegisterValue(t *testing.T) {
	constainer := di.NewContainer()
	di.RegisterValue(constainer, Service1{}, false)
	di.RegisterTransient[Service5](constainer, false)
	s5, err := di.Resolve[Service5](constainer)
	if s5 == nil || err != nil {
		t.Errorf("Resolve[Service5](constainer) = %v,%v; want %v,%v", s5, err, Service5{}, nil)
	} else {
		if s5.service1 == nil {
			t.Errorf("s5.service1 = %v; want %v", s5.service1, Service1{})
		}
	}

	s1, err := di.Resolve[Service1](constainer)
	if s1 == nil || err != nil {
		t.Errorf("Resolve[Service1](constainer) = %v,%v; want %v,%v", s1, err, Service1{}, nil)
	} else {
		s1.id = 100

		s12, _ := di.Resolve[Service1](constainer)
		if s12.id != 100 {
			t.Errorf("s12.id = %v; want %v", s12.id, 100)
		}
	}
}
