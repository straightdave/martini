package martini

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func TestNew(t *testing.T) {
	m := New()
	if m == nil {
		t.Error("martini.New() cannot return nil")
	}
}

func TestMartiniRunOnAddr(t *testing.T) {
	// just test that Run doesn't bomb
	go New().RunOnAddr("127.0.0.1:8080")
}

func TestMartiniRun(t *testing.T) {
	go New().Run()
}

func TestMartiniServeHTTP(t *testing.T) {
	result := ""
	response := httptest.NewRecorder()

	m := New()
	m.Use(func(c Context) {
		result += "foo"
		c.Next()
		result += "ban"
	})
	m.Use(func(c Context) {
		result += "bar"
		c.Next()
		result += "baz"
	})
	m.Action(func(res http.ResponseWriter, req *http.Request) {
		result += "bat"
		res.WriteHeader(http.StatusBadRequest)
	})

	m.ServeHTTP(response, (*http.Request)(nil))

	expect(t, result, "foobarbatbazban")
	expect(t, response.Code, http.StatusBadRequest)
}

func TestMartiniHandlers(t *testing.T) {
	result := ""
	response := httptest.NewRecorder()

	batman := func(c Context) {
		result += "batman!"
	}

	m := New()
	m.Use(func(c Context) {
		result += "foo"
		c.Next()
		result += "ban"
	})
	m.Handlers(
		batman,
		batman,
		batman,
	)
	m.Action(func(res http.ResponseWriter, req *http.Request) {
		result += "bat"
		res.WriteHeader(http.StatusBadRequest)
	})

	m.ServeHTTP(response, (*http.Request)(nil))

	expect(t, result, "batman!batman!batman!bat")
	expect(t, response.Code, http.StatusBadRequest)
}

func TestMartiniEarlyWrite(t *testing.T) {
	result := ""
	response := httptest.NewRecorder()

	m := New()
	m.Use(func(res http.ResponseWriter) {
		result += "foobar"
		res.Write([]byte("Hello world"))
	})
	m.Use(func() {
		result += "bat"
	})
	m.Action(func(res http.ResponseWriter) {
		result += "baz"
		res.WriteHeader(http.StatusBadRequest)
	})

	m.ServeHTTP(response, (*http.Request)(nil))

	expect(t, result, "foobar")
	expect(t, response.Code, http.StatusOK)
}

func TestMartiniWritten(t *testing.T) {
	response := httptest.NewRecorder()

	m := New()
	m.Handlers(func(res http.ResponseWriter) {
		res.WriteHeader(http.StatusOK)
	})

	ctx := m.createContext(response, (*http.Request)(nil))
	expect(t, ctx.Written(), false)

	ctx.run()
	expect(t, ctx.Written(), true)
}

func TestMartiniBasicNoRace(t *testing.T) {
	m := New()
	handlers := []Handler{func() {}, func() {}}
	// Ensure append will not realloc to trigger the race condition
	m.handlers = handlers[:1]
	req, _ := http.NewRequest("GET", "/", nil)
	for i := 0; i < 2; i++ {
		go func() {
			response := httptest.NewRecorder()
			m.ServeHTTP(response, req)
		}()
	}
}

func TestBreakHandlerStack(t *testing.T) {
	m := Classic()

	fn1 := func(c Context) string {
		c.Break()
		return "mid1"
	}

	fn2 := func(c Context) string {
		return "mid2"
	}

	m.Get("/", fn1, fn2, func() string {
		return "hello"
	})

	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()
	m.ServeHTTP(resp, req)
	expect(t, resp.Body.String(), "mid1")
}
