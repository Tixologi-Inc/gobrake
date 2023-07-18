package gobrake_test

import (
	"regexp"

	"github.com/Tixologi-Inc/gobrake"
)

func ExampleNewBlocklistKeysFilter() {
	notifier := gobrake.NewNotifier(1, "key")
	filter := gobrake.NewBlocklistKeysFilter("password", regexp.MustCompile("(?i)(user)"))
	notifier.AddFilter(filter)

	notice := &gobrake.Notice{
		Params: map[string]interface{}{
			"password": "slds2&LP",
			"User":     "username",
			"email":    "john@example.com",
		},
	}
	notifier.Notify(notice, nil)
}
