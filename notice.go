package gobrake

import (
	"encoding/json"
	"fmt"
	"go/build"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

var defaultContextOnce sync.Once
var defaultContext map[string]interface{}

func getDefaultContext() map[string]interface{} {
	defaultContextOnce.Do(func() {
		defaultContext = map[string]interface{}{
			"notifier": map[string]interface{}{
				"name":    notifierName,
				"version": notifierVersion,
				"url":     "https://github.com/airbrake/gobrake",
			},

			"language":     runtime.Version(),
			"os":           runtime.GOOS,
			"architecture": runtime.GOARCH,
		}

		if s, err := os.Hostname(); err == nil {
			defaultContext["hostname"] = s
		}

		if wd, err := os.Getwd(); err == nil {
			defaultContext["rootDirectory"] = wd
		}

		if s := gopath(); s != "" {
			defaultContext["gopath"] = s
		}
	})
	return defaultContext
}

// Returns the GOPATH of the application
func gopath() string {
	if path, ok := os.LookupEnv("GOPATH"); ok {
		return path
	}
	return build.Default.GOPATH
}

type StackFrame struct {
	File string         `json:"file"`
	Line int            `json:"line"`
	Func string         `json:"function"`
	Code map[int]string `json:"code,omitempty"`
}

type Error struct {
	Type      string       `json:"type"`
	Message   string       `json:"message"`
	Backtrace []StackFrame `json:"backtrace"`
}

type Notice struct {
	Id    string `json:"-"` // id returned by SendNotice
	Error error  `json:"-"` // error returned by SendNotice

	Errors  []Error                `json:"errors"`
	Context map[string]interface{} `json:"context"`
	Env     map[string]interface{} `json:"environment"`
	Session map[string]interface{} `json:"session"`
	Params  map[string]interface{} `json:"params"`
}

func (n *Notice) String() string {
	if len(n.Errors) == 0 {
		return "Notice<no errors>"
	}
	e := n.Errors[0]
	return fmt.Sprintf("Notice<%s: %s>", e.Type, e.Message)
}

func (n *Notice) SetRequest(ctx *gin.Context) {
	req := ctx.Request
	n.Context["url"] = req.URL.String()
	n.Context["httpMethod"] = req.Method
	if ua := req.Header.Get("User-Agent"); ua != "" {
		n.Context["userAgent"] = ua
	}
	n.Context["userAddr"] = remoteAddr(req)

	for k, v := range req.Header {
		if len(v) == 1 {
			n.Env[k] = v[0]
		} else {
			n.Env[k] = v
		}
	}

	setBody(req, n, ctx)
	setUser(req, n)
}

func setBody(req *http.Request, n *Notice, ctx *gin.Context) {
	if req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch {
		bodyString := ctx.GetString("BodyString")

		var body interface{}
		if err := json.Unmarshal([]byte(bodyString), &body); err != nil {
			n.Context["body"] = "error parsing body: " + err.Error()
		} else {
			n.Context["body"] = body
		}
	}
}

func setUser(req *http.Request, n *Notice) {
	if authHeader := req.Header.Get("Authorization"); authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		jwtToken := strings.TrimPrefix(authHeader, "Bearer ")
		claims := jwt.MapClaims{}
		token, _, err := jwt.NewParser().ParseUnverified(jwtToken, claims)
		if err != nil {
			n.Context["user"] = "error parsing jwt: " + err.Error()
		} else {
			n.Context["user"] = token.Claims
		}
	}
}

func remoteAddr(req *http.Request) string {
	if s := req.Header.Get("X-Forwarded-For"); s != "" {
		parts := strings.Split(s, ",")
		return parts[0]
	}

	if s := req.Header.Get("X-Real-Ip"); s != "" {
		return s
	}

	parts := strings.Split(req.RemoteAddr, ":")
	return parts[0]
}

func NewNotice(e interface{}, ctx *gin.Context, depth int) *Notice {
	notice, ok := e.(*Notice)
	if ok {
		return notice
	}

	typeName := getTypeName(e)
	notice = &Notice{
		Errors: []Error{{
			Type:    typeName,
			Message: fmt.Sprint(e),
		}},
		Context: make(map[string]interface{}),
		Env:     make(map[string]interface{}),
		Session: make(map[string]interface{}),
		Params:  make(map[string]interface{}),
	}

	for k, v := range getDefaultContext() {
		notice.Context[k] = v
	}

	if depth != -1 {
		packageName, backtrace := getBacktrace(e, depth+2)
		notice.Errors[0].Backtrace = backtrace
		notice.Context["component"] = packageName
	}

	if ctx != nil && ctx.Request != nil {
		notice.SetRequest(ctx)
	}

	return notice
}

// getTypeName returns the type name of e.
func getTypeName(e interface{}) string {
	if err, ok := e.(error); ok {
		e = errors.Cause(err)
	}
	return fmt.Sprintf("%T", e)
}
