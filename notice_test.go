package gobrake_test

import (
	"errors"

	"github.com/Tixologi-Inc/gobrake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewNotice", func() {
	var notice *gobrake.Notice

	BeforeEach(func() {
		notice = gobrake.NewNotice(errors.New("test"), nil, 0)
	})

	It("returns correct backtrace", func() {
		Expect(notice.Errors[0].Backtrace[0].File).To(ContainSubstring("gobrake/notice_test.go"))
	})
})
