package gsf

import (
	"errors"
	"fmt"
	"gitee.com/kzangv/gsf-fof/logger"
	"github.com/urfave/cli/v2"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	EnvLocal   = 0 // 本地
	EnvTest    = 1 // 测试
	EnvPrev    = 2 // 预发布
	EnvRelease = 3 // 线上

	CliAppEnv     = "app-env"
	CliAppVer     = "app-version"
	CliAppLogMore = "app-log-more"
	InvalidPath   = "__invalid__"
)

var (
	EnvVersion = "1.0.0"

	EnvLocalArg   = "local"   // 本地
	EnvTestArg    = "test"    // 测试
	EnvPrevArg    = "preview" // 预发布
	EnvReleaseArg = "release" // 线上
)

type _EnvFlag int

func (v _EnvFlag) Usage() string {
	ret := make([]string, 0, 4)
	ret = append(ret, fmt.Sprintf("%s-%s", EnvLocalArg, "本地环境"))
	ret = append(ret, fmt.Sprintf("%s-%s", EnvTestArg, "测试环境"))
	ret = append(ret, fmt.Sprintf("%s-%s", EnvPrevArg, "预发布环境"))
	ret = append(ret, fmt.Sprintf("%s-%s", EnvReleaseArg, "线上环境"))
	return "env(" + strings.Join(ret, ", ") + ")"
}

func (v *_EnvFlag) Action(_ *cli.Context, sv string) error {
	switch sv {
	case EnvLocalArg:
		*v = EnvLocal
	case EnvTestArg:
		*v = EnvTest
	case EnvPrevArg:
		*v = EnvPrev
	case EnvReleaseArg:
		*v = EnvRelease
	default:
		return errors.New("Env Is Invalid: " + sv)
	}
	return nil
}

type Config struct {
	env              _EnvFlag
	version, execDir string
	logMore          bool
}

func (v *Config) SetEnv(e int) {
	v.env = _EnvFlag(e)
}

func (v *Config) Env() int {
	return int(v.env)
}

func (v *Config) EnvDesc() string {
	switch v.env {
	case EnvLocal:
		return EnvLocalArg
	case EnvTest:
		return EnvTestArg
	case EnvPrev:
		return EnvPrevArg
	case EnvRelease:
		return EnvReleaseArg
	}
	return ""
}

func (v *Config) ExecDir() string {
	if v.execDir == "" {
		var err error
		if v.execDir, err = os.Getwd(); err != nil {
			v.execDir = InvalidPath
		}
	}
	return v.execDir
}

func (v *Config) Version() string {
	return v.version
}

func (v *Config) SetLogMore(lm bool) {
	v.logMore = lm
}

func (v *Config) LogMore() bool {
	return v.logMore
}

type Component interface {
	CliFlags() []cli.Flag
	Init(logger.Interface, Config) error
	Run(logger.Interface, Config) error
	Close(logger.Interface, Config) error
}

type Service interface {
	CliFlags() []cli.Flag
	Init(*Config, *cli.Context) (logger.Interface, error)
	Run(logger.Interface, *Config) error
	Close()
}

type Application struct {
	Component map[string]Component
	Log       logger.Interface
	Cfg       Config
	Ser       Service
}

func (app *Application) closeComponent() {
	if app.Component != nil && len(app.Component) > 0 {
		wg := sync.WaitGroup{}
		wg.Add(len(app.Component))
		for k := range app.Component {
			go func(v Component) {
				defer wg.Done()
				_ = v.Close(app.Log, app.Cfg)
			}(app.Component[k])
		}

		wg.Wait()
	}
}

func (app *Application) shutdown() {
	defer func() {
		fmt.Println("Application Had Exit")
		os.Exit(0)
	}()
	fmt.Println("Application To Exit")
}

func (app *Application) catchSignal() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for s := range c {
		switch s {
		case syscall.SIGHUP:
			fallthrough
		case syscall.SIGINT, syscall.SIGTERM:
			app.closeComponent()
			app.Ser.Close()
			app.shutdown()
		}
	}
}

func (app *Application) runComponent() error {
	if app.Component != nil && len(app.Component) > 0 {
		wg := sync.WaitGroup{}
		lock := sync.Mutex{}
		var errList []error

		wg.Add(len(app.Component))
		for k := range app.Component {
			go func(v Component) {
				defer wg.Done()
				err := v.Run(app.Log, app.Cfg)
				if err != nil {
					lock.Lock()
					if errList == nil {
						errList = make([]error, 0)
					}
					errList = append(errList, err)
					lock.Unlock()
				}
			}(app.Component[k])
		}

		wg.Wait()
		if len(errList) > 0 {
			errStr := make([]string, 0, len(errList))
			for k := range errList {
				errStr = append(errStr, errList[k].Error())
			}
			return errors.New("Component Run Error:" + strings.Join(errStr, ","))
		}
	}
	return nil
}

func (app *Application) run() error {
	var err error = nil

	runtime.GOMAXPROCS(runtime.NumCPU())
	go app.catchSignal()

	// 初始化随机种子
	rand.Seed(time.Now().UnixNano())

	// 启动服务
	err = app.runComponent()
	if err != nil {
		return err
	}

	return app.Ser.Run(app.Log, &app.Cfg)
}

func (app *Application) Start(cmd *cli.App, args []string, service Service) {
	if cmd == nil {
		cmd = cli.NewApp()
		cmd.Version = "1.0"
		cmd.Name = "Application"
		cmd.Usage = "Application Server"
		cmd.DisableSliceFlagSeparator = true
	}

	app.Ser = service

	// app
	cfs := make([][]cli.Flag, 0, len(app.Component)+2)
	cfs = append(cfs, []cli.Flag{
		// base
		&cli.StringFlag{Name: CliAppEnv, Usage: app.Cfg.env.Usage(), Action: app.Cfg.env.Action},
		&cli.StringFlag{Name: CliAppVer, Value: EnvVersion, Usage: "version", Destination: &app.Cfg.version},
		&cli.BoolFlag{Name: CliAppLogMore, Value: false, Usage: "log more", Destination: &app.Cfg.logMore},
	})
	fsLen := len(cfs[0])

	// service
	afs := app.Ser.CliFlags()
	if len(afs) > 0 {
		cfs = append(cfs, afs)
		fsLen += len(afs)
	}

	// component
	for k := range app.Component {
		v := app.Component[k].CliFlags()
		if len(v) > 0 {
			cfs = append(cfs, v)
			fsLen += len(v)
		}
	}

	// cli merge
	fs := make([]cli.Flag, 0, fsLen)
	for _, v := range cfs {
		fs = append(fs, v...)
	}

	// 执行命令行
	cmd.Flags = fs
	cmd.Action = func(ctx *cli.Context) error {
		var err error = nil
		app.Cfg.execDir, err = os.Getwd()
		if err == nil {
			// 用命令行初始化配置
			app.Log, err = app.Ser.Init(&app.Cfg, ctx)
			if err == nil {
				// 组件初始化
				for k := range app.Component {
					if err = app.Component[k].Init(app.Log, app.Cfg); err != nil {
						err = errors.New("Component Init Error: " + err.Error())
						break
					}
				}
				// 启动程序
				if err == nil {
					err = app.run()
				}
			}
		} else {
			err = errors.New("dir is invalid: " + app.Cfg.execDir)
		}
		return err
	}

	err := cmd.Run(args)
	if err != nil {
		fmt.Printf("Init Error: %s", err.Error())
		os.Exit(1)
	}
}
