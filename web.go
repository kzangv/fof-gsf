package gsf

import (
	"fmt"
	"gitee.com/kzangv/gsf-fof/logger"
	"github.com/urfave/cli/v2"
	"net/http"
	"os"
	"time"
)

const (
	CliWebIP           = "web-ip"
	CliWebPort         = "web-port"
	CliWebReadTimeout  = "web-r-timeout"
	CliWebWriteTimeout = "web-w-timeout"
	CliWebIdleTimeout  = "web-idle-timeout"
)

type WebConfig struct {
	IP      string `json:"ip"   yaml:"ip"`
	Port    int    `json:"port" yaml:"port"`
	Timeout struct {
		Read  int `json:"read"  yaml:"read"`
		Write int `json:"write" yaml:"write"`
		Idle  int `json:"idle"  yaml:"idle"`
	} `json:"timeout" yaml:"timeout"`
}

type WebService struct {
	Cfg                   WebConfig
	Handler               http.Handler
	BeforeRun, BeforeInit func(l logger.Interface) error
}

func (c *WebService) CliFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: CliWebIP, Value: "0.0.0.0", Usage: "web service ip", Action: func(_ *cli.Context, i string) error { c.Cfg.IP = i; return nil }},
		&cli.IntFlag{Name: CliWebPort, Value: 9980, Usage: "web service port", Action: func(_ *cli.Context, i int) error { c.Cfg.Port = i; return nil }},
		&cli.IntFlag{Name: CliWebReadTimeout, Value: 0, Usage: "web service read timeout", Action: func(_ *cli.Context, i int) error { c.Cfg.Timeout.Read = i; return nil }},
		&cli.IntFlag{Name: CliWebWriteTimeout, Value: 0, Usage: "web service write timeout", Action: func(_ *cli.Context, i int) error { c.Cfg.Timeout.Write = i; return nil }},
		&cli.IntFlag{Name: CliWebIdleTimeout, Value: 0, Usage: "web service idle timeout", Action: func(_ *cli.Context, i int) error { c.Cfg.Timeout.Idle = i; return nil }},
	}
}

func (c *WebService) Init(cfg *Config, _ *cli.Context) (logger.Interface, error) {
	v := &logger.Console{}
	switch cfg.Env() {
	case EnvLocal:
		v.Init(logger.Info, true, os.Stdout, os.Stderr)
	case EnvTest, EnvPrev, EnvRelease:
		v.Init(logger.Warn, false, os.Stdout, os.Stderr)
	}
	if c.BeforeInit != nil {
		if err := c.BeforeInit(v); err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (c *WebService) Run(l logger.Interface, cfg *Config) error {
	if c.BeforeRun != nil {
		if err := c.BeforeRun(l); err != nil {
			return err
		}
	}

	if c.Cfg.IP == "" {
		l.WarnForce("Listening Server [[[ Run-Mode: %s ]]] http://127.0.0.1:%d\n", cfg.EnvDesc(), c.Cfg.Port)
	} else {
		l.WarnForce("Listening Server [[[ Run-Mode: %s ]]] http://%s:%d\n", cfg.EnvDesc(), c.Cfg.IP, c.Cfg.Port)
	}

	// web
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", c.Cfg.IP, c.Cfg.Port),
		ReadTimeout:  time.Duration(c.Cfg.Timeout.Read) * time.Second,
		WriteTimeout: time.Duration(c.Cfg.Timeout.Write) * time.Second,
		IdleTimeout:  time.Duration(c.Cfg.Timeout.Idle) * time.Second,
		Handler:      c.Handler,
	}
	srv.SetKeepAlivesEnabled(true)
	return srv.ListenAndServe()
}

func (c *WebService) Close() {
}
