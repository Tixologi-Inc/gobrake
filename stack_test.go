package gobrake

import (
	testpkg1 "github.com/Tixologi-Inc/gobrake/internal/testpkg1"
	testpkg2 "github.com/Tixologi-Inc/gobrake/internal/testpkg2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("backtraceFromErrorWithStackTrace", func() {
	It("returns package name", func() {
		var tests = []struct {
			err         error
			packageName string
		}{{
			err:         testpkg1.Foo(),
			packageName: "github.com/Tixologi-Inc/gobrake/internal/testpkg1",
		}, {
			err:         testpkg1.Bar(),
			packageName: "github.com/Tixologi-Inc/gobrake/internal/testpkg1",
		}, {
			err:         testpkg2.NewError(),
			packageName: "github.com/Tixologi-Inc/gobrake/internal/testpkg2",
		}}

		type stackTracer interface {
			StackTrace() errors.StackTrace
		}

		for _, test := range tests {
			v, ok := test.err.(stackTracer)

			Expect(ok).To(BeTrue())
			packageName, _ := backtraceFromErrorWithStackTrace(v)
			Expect(packageName).To(Equal(test.packageName))
		}
	})
})
