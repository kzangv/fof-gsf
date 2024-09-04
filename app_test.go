package gsf

import (
	"gitee.com/kzangv/gsf-fof/logger"
	"github.com/urfave/cli/v2"
	"net/http"
	"testing"
)

type _TestComponent struct {
	name string
}

func (c *_TestComponent) CliFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "test-name", Value: "test", Usage: "test name", Destination: &c.name},
	}
}

func (c *_TestComponent) Init(log logger.Interface, cfg Config) error {
	log.InfoForce("Component Run %d-v:%s-d:%s", cfg.Env(), cfg.Version(), cfg.ExecDir())
	return nil
}

func (c *_TestComponent) Run(_ logger.Interface, _ Config) error {
	return nil
}

func (c *_TestComponent) Close(_ logger.Interface, _ Config) error {
	return nil
}

type _WebRouter struct {
	app *Application
}

func (r _WebRouter) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(http.StatusOK)
	resp.Header()
	resp.Write([]byte(req.URL.String() + " ## evn: " + r.app.Cfg.EnvDesc()))
}

func TestWeb(t *testing.T) {
	app := Application{
		Component: map[string]Component{
			"test": &_TestComponent{},
		},
	}
	app.Start(
		nil,
		[]string{"test", "--app-env=local", "--web-port=8888", "--web-r-timeout=10", "--web-w-timeout=9", "--web-idle-timeout=8"},
		&WebService{Handler: &_WebRouter{app: &app}},
	)
}

func TestCmd(t *testing.T) {
	app := Application{
		Component: map[string]Component{
			"test": &_TestComponent{},
		},
	}
	cmd := &CmdService{}
	cmd.AddCmdFunc("do", func(log logger.Interface) error {
		log.InfoForce("do cmd ## ver=%s:env=%s;log-more=%t",
			app.Cfg.Version(),
			app.Cfg.EnvDesc(),
			app.Cfg.LogMore())
		return nil
	})

	app.Start(
		nil,
		[]string{"test", "--app-env=release", "--app-version=v2.0", "--app-log-more=true", "--app-cmd=do"},
		cmd,
	)
}
