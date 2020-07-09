package martini

import (
	"os"
)

// Envs
const (
	Dev  string = "development"
	Prod        = "production"
	Test        = "test"
)

// Env is the environment that Martini is executing in.
// The MARTINI_ENV is read on initialization to set this variable.
var Env = Dev

// Root is the rooted path name of current directory of the application.
var Root string

func setENV(e string) {
	if len(e) > 0 {
		Env = e
	}
}

func init() {
	setENV(os.Getenv("MARTINI_ENV"))
	var err error
	Root, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}
