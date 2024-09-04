package gsf

import (
	"errors"
	"github.com/kzangv/gsf-fof/logger"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

var (
	CmdNoFind = errors.New(" Command No Find")
)

type CmdArg struct {
	Name, Default string
}

type CmdHandle interface {
	Flags() []CmdArg
	Run(logger.Interface, map[string]string) error
}

type _DefaultCmdHandle struct {
	handler func(logger.Interface) error
}

func (m *_DefaultCmdHandle) Flags() []CmdArg {
	return nil
}
func (m *_DefaultCmdHandle) Run(l logger.Interface, _ map[string]string) error {
	return m.handler(l)
}

type CmdService struct {
	cmd, arg              string
	Router                map[string]CmdHandle
	BeforeRun, BeforeInit func(l logger.Interface) error
}

func (c *CmdService) AddCmdFunc(name string, handle func(logger.Interface) error) *CmdService {
	return c.AddCmdHandle(name, &_DefaultCmdHandle{handler: handle})
}

func (c *CmdService) AddCmdHandle(name string, handle CmdHandle) *CmdService {
	if c.Router == nil {
		c.Router = make(map[string]CmdHandle)
	}
	c.Router[name] = handle
	return c
}

func (c *CmdService) CliFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "app-cmd", Value: "", Usage: "cmd", Destination: &c.cmd},
		&cli.StringFlag{Name: "app-args", Value: "", Usage: "cmd args", Destination: &c.arg},
	}
}

func (c *CmdService) Init(cfg *Config, _ *cli.Context) (logger.Interface, error) {
	v := &logger.Console{}
	switch cfg.Env() {
	case EnvLocal:
		v.Init(logger.Info, true, os.Stdout, os.Stderr)
	case EnvTest, EnvPrev, EnvRelease:
		v.Init(logger.Warn, false, os.Stdout, os.Stderr)
	}

	if _, ok := c.Router[c.cmd]; !ok {
		return nil, CmdNoFind
	}

	if c.BeforeInit != nil {
		if err := c.BeforeInit(v); err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (c *CmdService) Run(l logger.Interface, _ *Config) error {
	if c.BeforeRun != nil {
		if err := c.BeforeRun(l); err != nil {
			return err
		}
	}

	if handle, ok := c.Router[c.cmd]; ok {
		// 初始化
		flags := handle.Flags()
		args := make(map[string]string)
		for k := range flags {
			args[flags[k].Name] = flags[k].Default
		}
		if len(c.arg) > 0 {
			argRaws := strings.Split(c.arg, ";")
			for k := range argRaws {
				pVal := strings.Split(argRaws[k], ":")
				if len(pVal) >= 2 {
					if _, ok := args[pVal[0]]; ok {
						args[pVal[0]] = pVal[1]
					}
				}
			}
		}
		return handle.Run(l, args)
	}
	return CmdNoFind
}

func (c *CmdService) Close() {
}
