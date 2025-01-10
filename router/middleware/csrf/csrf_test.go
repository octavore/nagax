package csrf

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/octavore/naga/service"
	"github.com/shoenig/test"

	"github.com/octavore/nagax/users/session"
	"github.com/octavore/nagax/util/memlogger"
)

type TestModule struct {
	*Module
}

func (m *TestModule) Init(c *service.Config) {
	c.Setup = func() error {
		m.Logger.Logger = &memlogger.MemoryLogger{}
		return nil
	}
}

type testEnv struct {
	module *Module
	logger *memlogger.MemoryLogger
	stop   func()
}

func setup() testEnv {
	module, stop := service.New(&TestModule{}).StartForTest()
	return testEnv{
		module: module.Module,
		logger: module.Logger.Logger.(*memlogger.MemoryLogger),
		stop:   stop,
	}
}

func TestNew(t *testing.T) {
	env := setup()
	defer env.stop()

	testHandler := env.module.New("/ignore", "/ignore2/:id")
	testPaths := []struct {
		path   string
		method string
		valid  bool
	}{
		{path: "/", method: "POST", valid: false},                // INVALID
		{path: "/path", method: "POST", valid: false},            // INVALID
		{path: "/path", method: "GET", valid: true},              // -> valid because method is ignored
		{path: "/ignore", method: "POST", valid: true},           // -> valid because ignored
		{path: "/ignore2", method: "POST", valid: false},         // INVALID
		{path: "/ignore2/foo", method: "POST", valid: true},      // -> valid because it matches pattern
		{path: "/ignore2/foo/sub", method: "POST", valid: false}, // INVALID
	}

	session, err := env.module.Session.NewSessionCookie(&session.UserSession{ID: "1", SessionID: "fake"})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range testPaths {
		req := httptest.NewRequest(c.method, c.path, nil)
		req.AddCookie(session)
		// req.Header.Set(csrfHeaderKey, "fake")
		rr := httptest.NewRecorder()
		testHandler(rr, req, func(rw http.ResponseWriter, req *http.Request) {})
		if c.valid {
			test.Eq(t, rr.Code, 200, test.Sprintf("%s %s should be valid", c.method, c.path))
		} else {
			test.Eq(t, rr.Code, 400, test.Sprintf("%s %s should be invalid", c.method, c.path))
		}
	}

}
