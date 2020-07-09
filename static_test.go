package martini

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/straightdave/inject"
)

var currentRoot, _ = os.Getwd()

func TestStatic(t *testing.T) {
	response := httptest.NewRecorder()
	response.Body = new(bytes.Buffer)

	m := New()
	r := NewRouter()

	m.Use(Static(currentRoot))
	m.Action(r.Handle)

	req, err := http.NewRequest("GET", "http://localhost:3000/martini.go", nil)
	if err != nil {
		t.Error(err)
	}
	m.ServeHTTP(response, req)
	expect(t, response.Code, http.StatusOK)
	expect(t, response.Header().Get("Expires"), "")
	if response.Body.Len() == 0 {
		t.Errorf("Got empty body for GET request")
	}
}

func TestStaticLocalPath(t *testing.T) {
	currentRoot = os.TempDir()
	response := httptest.NewRecorder()
	response.Body = new(bytes.Buffer)

	m := New()
	r := NewRouter()

	m.Use(Static(currentRoot))
	f, err := ioutil.TempFile(currentRoot, "static_content")
	if err != nil {
		t.Error(err)
	}

	t.Logf("===== currentroot: %s,  temp file at: %s", currentRoot, f.Name())

	f.WriteString("Expected Content")
	f.Close()
	m.Action(r.Handle)

	url := "http://localhost:3000/" + path.Base(f.Name())
	t.Logf("===== url: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Error(err)
	}
	m.ServeHTTP(response, req)
	expect(t, response.Code, http.StatusOK)
	expect(t, response.Header().Get("Expires"), "")
	expect(t, response.Body.String(), "Expected Content")
}

func TestStaticHead(t *testing.T) {
	response := httptest.NewRecorder()
	response.Body = new(bytes.Buffer)

	m := New()
	r := NewRouter()

	m.Use(Static("."))
	m.Action(r.Handle)

	req, err := http.NewRequest("HEAD", "http://localhost:3000/martini.go", nil)
	if err != nil {
		t.Error(err)
	}

	m.ServeHTTP(response, req)
	expect(t, response.Code, http.StatusOK)
	if response.Body.Len() != 0 {
		t.Errorf("Got non-empty body for HEAD request")
	}
}

func TestStaticAsPost(t *testing.T) {
	response := httptest.NewRecorder()

	m := New()
	r := NewRouter()

	m.Use(Static(currentRoot))
	m.Action(r.Handle)

	req, err := http.NewRequest("POST", "http://localhost:3000/martini.go", nil)
	if err != nil {
		t.Error(err)
	}

	m.ServeHTTP(response, req)
	expect(t, response.Code, http.StatusNotFound)
}

func TestStaticBadDir(t *testing.T) {
	response := httptest.NewRecorder()

	m := Classic()

	req, err := http.NewRequest("GET", "http://localhost:3000/martini.go", nil)
	if err != nil {
		t.Error(err)
	}

	m.ServeHTTP(response, req)
	refute(t, response.Code, http.StatusOK)
}

func TestStaticOptionsLogging(t *testing.T) {
	response := httptest.NewRecorder()

	var buffer bytes.Buffer
	m := &Martini{Injector: inject.New(), action: func() {}, logger: log.New(&buffer, "[martini] ", 0)}
	m.Map(m.logger)
	m.Map(defaultReturnHandler())

	opt := StaticOptions{}
	m.Use(Static(".", opt))

	req, err := http.NewRequest("GET", "http://localhost:3000/martini.go", nil)
	if err != nil {
		t.Error(err)
	}

	m.ServeHTTP(response, req)
	expect(t, response.Code, http.StatusOK)
	expect(t, buffer.String(), "[martini] [Static] Serving /martini.go\n")

	// Now without logging
	m.Handlers()
	buffer.Reset()

	// This should disable logging
	opt.SkipLogging = true
	m.Use(Static(currentRoot, opt))

	m.ServeHTTP(response, req)
	expect(t, response.Code, http.StatusOK)
	expect(t, buffer.String(), "")
}

func TestStaticOptionsServeIndex(t *testing.T) {
	response := httptest.NewRecorder()

	var buffer bytes.Buffer
	m := &Martini{Injector: inject.New(), action: func() {}, logger: log.New(&buffer, "[martini] ", 0)}
	m.Map(m.logger)
	m.Map(defaultReturnHandler())

	opt := StaticOptions{IndexFile: "martini.go"} // Define martini.go as index file
	m.Use(Static(".", opt))

	req, err := http.NewRequest("GET", "http://localhost:3000/", nil)
	if err != nil {
		t.Error(err)
	}

	m.ServeHTTP(response, req)
	expect(t, response.Code, http.StatusOK)
	expect(t, buffer.String(), "[martini] [Static] Serving /martini.go\n")
}

func TestStaticOptionsPrefix(t *testing.T) {
	response := httptest.NewRecorder()

	var buffer bytes.Buffer
	m := &Martini{Injector: inject.New(), action: func() {}, logger: log.New(&buffer, "[martini] ", 0)}
	m.Map(m.logger)
	m.Map(defaultReturnHandler())

	// Serve current directory under /public
	m.Use(Static(".", StaticOptions{Prefix: "/public"}))

	// Check file content behaviour
	req, err := http.NewRequest("GET", "http://localhost:3000/public/martini.go", nil)
	if err != nil {
		t.Error(err)
	}

	m.ServeHTTP(response, req)
	expect(t, response.Code, http.StatusOK)
	expect(t, buffer.String(), "[martini] [Static] Serving /martini.go\n")
}

func TestStaticOptionsExpires(t *testing.T) {
	response := httptest.NewRecorder()

	var buffer bytes.Buffer
	m := &Martini{Injector: inject.New(), action: func() {}, logger: log.New(&buffer, "[martini] ", 0)}
	m.Map(m.logger)
	m.Map(defaultReturnHandler())

	// Serve current directory under /public
	m.Use(Static(".", StaticOptions{Expires: func() string { return "46" }}))

	// Check file content behaviour
	req, err := http.NewRequest("GET", "http://localhost:3000/martini.go", nil)
	if err != nil {
		t.Error(err)
	}

	m.ServeHTTP(response, req)
	expect(t, response.Header().Get("Expires"), "46")
}

func TestStaticOptionsFallback(t *testing.T) {
	response := httptest.NewRecorder()

	var buffer bytes.Buffer
	m := &Martini{Injector: inject.New(), action: func() {}, logger: log.New(&buffer, "[martini] ", 0)}
	m.Map(m.logger)
	m.Map(defaultReturnHandler())

	// Serve current directory under /public
	m.Use(Static(".", StaticOptions{Fallback: "/martini.go"}))

	// Check file content behaviour
	req, err := http.NewRequest("GET", "http://localhost:3000/initram.go", nil)
	if err != nil {
		t.Error(err)
	}

	m.ServeHTTP(response, req)
	expect(t, response.Code, http.StatusOK)
	expect(t, buffer.String(), "[martini] [Static] Serving /martini.go\n")
}

func TestStaticRedirect(t *testing.T) {
	response := httptest.NewRecorder()

	m := New()
	m.Use(Static(currentRoot, StaticOptions{Prefix: "/public"}))

	req, err := http.NewRequest("GET", "http://localhost:3000/public?param=foo#bar", nil)
	if err != nil {
		t.Error(err)
	}

	m.ServeHTTP(response, req)
	expect(t, response.Code, http.StatusFound)
	expect(t, response.Header().Get("Location"), "/public/?param=foo#bar")
}
