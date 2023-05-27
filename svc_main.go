package wsvc

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/sys/windows/svc"
)

type ServiceConfig struct {
	ServiceName string
	DisplayName string
	Description string
	Constructor ConstructorInstanceFunc
}

func usage(errmsg string) {
	fmt.Fprintf(os.Stderr,
		"%s\n\n"+
			"usage: %s <command>\n"+
			"       where <command> is one of\n"+
			"       install, remove, debug, start, stop, pause or continue.\n",
		errmsg, os.Args[0])
	os.Exit(2)
}

func StartService(svcName string) error {
	return startService(svcName)
}

func StopService(svcName string) error {
	return controlService(svcName, svc.Stop, svc.Stopped)
}

func RemoveService(svcName string) error {
	return removeService(svcName)
}

func SvcMain(config ServiceConfig) {

	var (
		as  bool
		err error
	)

	as, err = svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}

	if !as {
		runService(config.ServiceName, false, config.Constructor)
		return
	}

	if len(os.Args) < 2 {
		usage("no command specified")
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "debug":
		runService(config.ServiceName, true, config.Constructor)
		return
	case "install":
		err = installService(config.ServiceName, config.DisplayName, config.Description)
	case "remove":
		err = removeService(config.ServiceName)
	case "start":
		err = startService(config.ServiceName)
	case "stop":
		err = controlService(config.ServiceName, svc.Stop, svc.Stopped)
	case "pause":
		err = controlService(config.ServiceName, svc.Pause, svc.Paused)
	case "continue":
		err = controlService(config.ServiceName, svc.Continue, svc.Running)
	default:
		usage(fmt.Sprintf("invalid command %s", cmd))
	}

	if err != nil {
		log.Fatalf("failed to %s %s: %v", cmd, config.ServiceName, err)
	}
	return
}
