package source

import (
	"github.com/nikandfor/loc"
	"github.com/nikandfor/tlog"
)

func AssertionArgs(t interface{}) []string {

	tlog.Printw("testing.T", "type", tlog.FormatNext("%T"), t)

	pcs := loc.Callers(1, 10)

	for _, pc := range pcs {
		n, f, l := pc.NameFileLine()

		tlog.Printw("caller", "pc", uintptr(pc), "n", n, "f", f, "l", l)
	}

	return nil
}
