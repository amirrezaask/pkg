package dump

import (
	"os"

	"github.com/davecgh/go-spew/spew"
)

func This(obj any) {
	spew.Dump(obj)
}

func AndDie(obj any) {
	spew.Dump(obj)
	os.Exit(0)
}
