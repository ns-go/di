package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ns-go/di/pkg/di"
)

func TestMiddleware(t *testing.T) {
	constainer := di.NewContainer()
	di.RegisterSingleton[Service1](constainer, false)
	di.RegisterScoped[Service2](constainer, false)

	middleware := di.Middleware(constainer, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, ok := r.Context().Value(di.ContextContainerKey).(*di.Container)
		if !ok {
			t.Errorf(`c, ok := r.Context().Value(di.ContextContainerKey).(*di.Container) ok = %v; want %v`, ok, true)
		}
		if c == nil {
			t.Errorf(`c, ok := r.Context().Value(di.ContextContainerKey).(*di.Container) c = %v; want %v`, c, di.Container{})
		}

		s2, err := di.Resolve[Service2](c)
		if s2 == nil || err != nil {
			t.Errorf(`s2, err := di.Resolve[Service2](c) = (%v, %v); want (%v, %v)`, s2, err, Service2{}, nil)
		}
	}))
	server := httptest.NewServer(middleware)

	_, err := http.Get(server.URL)
	if err != nil {
		t.Errorf(`_, err = http.Get(server.URL) err = %v; want %v`, err, nil)
	}
}
