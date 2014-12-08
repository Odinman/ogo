package ogo

import (
	"os"
	"path/filepath"

	"github.com/Odinman/ogo/libs/config"
)

/* {{{ type Environment struct
 */
type Environment struct {
	WorkPath      string // working path(abs)
	AppPath       string // application path
	ProcName      string // proc name
	Worker        string // worker name
	AppConfigPath string // config file path
	RunMode       string // run mode, "dev" or "prod"
	Daemonize     bool   // daemonize or not
	DebugLevel    int    // debug level
	PidFile       string // pidfile abs path
	Port          string // http port
	IndentJSON    bool   // indent JSON
}

/* }}} */

/* {{{ func ParseConfig() (err error)
 * ParseConfig parsed default config file.
 */
func ParseConfig() (err error) {
	appConfig, err = config.NewConfig("ini", env.AppConfigPath)
	if err != nil {
		appConfig = config.NewFakeConfig()
		return err
	} else {

		if port := appConfig.String("Port"); port != "" {
			env.Port = port
		}
		os.Setenv("PORT", env.Port) // pass to bind

		if runmode := appConfig.String("RunMode"); runmode != "" {
			env.RunMode = runmode
		}

		if workerName := appConfig.String("Worker"); workerName != "" {
			env.Worker = workerName
		}

		if daemonize, err := appConfig.Bool("Daemonize"); err == nil {
			env.Daemonize = daemonize
		}
		if indentJson, err := appConfig.Bool("IndentJson"); err == nil {
			env.IndentJSON = indentJson
		}
		if pidfile := appConfig.String("PidFile"); pidfile != "" {
			// make sure pidfile is abs path
			if filepath.IsAbs(pidfile) {
				env.PidFile = pidfile
			} else {
				env.PidFile = filepath.Join(env.AppPath, pidfile)
			}
		}
		if level, err := appConfig.Int("DebugLevel"); err == nil {
			env.DebugLevel = level
		}

	}
	return nil
}

/* }}} */
