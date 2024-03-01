package maybe_test

import (
	"fmt"

	"golang.org/x/xerrors"

	"github.com/coder/coder/v2/coderd/util/maybe"
)

func ExampleMaybe() {
	m1 := maybe.Of("hello")
	m2 := maybe.Not[string](xerrors.New("goodbye"))
	_, _ = fmt.Println(m1.Valid())
	_, _ = fmt.Println(*m1.Value())
	_, _ = fmt.Println(m1.Error())
	_, _ = fmt.Println(m2.Valid())
	_, _ = fmt.Println(m2.Value())
	_, _ = fmt.Println(m2.Error())
	// Output: true
	// hello
	// <nil>
	// false
	// <nil>
	// goodbye
}
