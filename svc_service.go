package wsvc

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

type Service interface {
	Startup() error
	Stop() error
}

type ConstructorInstanceFunc func() (Service, error)

var elog debug.Log

type WrapperServiceImpl struct {
	name     string
	instance Service
	stopCh   chan struct{}
}

func runService(name string, isDebug bool, constructor ConstructorInstanceFunc) {

	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}

	defer elog.Close()
	elog.Info(1, fmt.Sprintf("%s: starting.", name))
	elog.Info(1, fmt.Sprintf("%s: isDebug: %v", name, isDebug))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	elog.Info(1, fmt.Sprintf("%s: constructor start", name))
	instance, err := constructor()
	elog.Info(1, fmt.Sprintf("%s: constructor end : %v", name, err))
	if err != nil {
		elog.Info(1, fmt.Sprintf("%s constructor service failed: %v", name, err))
		elog.Error(1, fmt.Sprintf("%s constructor service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s: constructor successed", name))
	stopCh := make(chan struct{})
	err = run(name, &WrapperServiceImpl{
		name:     name,
		instance: instance,
		stopCh:   stopCh,
	})

	if err != nil {
		close(stopCh)
		elog.Info(1, fmt.Sprintf("%s service failed: %v", name, err))
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s: stopped", name))
}

func svcLauncher(name string, instance Service, stopCh <-chan struct{}) {

	go func() {
		if err := instance.Startup(); err != nil {
			elog.Error(1, fmt.Sprintf("%s service startup failed: %v", name, err))
			return
		}
		defer instance.Stop()
		elog.Info(1, fmt.Sprintf("%s: running.", name))
		<-stopCh
	}()
}

func (service *WrapperServiceImpl) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {

	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	fasttick := time.Tick(500 * time.Millisecond)
	slowtick := time.Tick(2 * time.Second)
	tick := fasttick
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	svcLauncher(service.name, service.instance, service.stopCh)

loop:
	for {
		select {
		case <-tick:
			// appLauncher()
			// elog.Info(1, "Launching app...")
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				close(service.stopCh)
				time.Sleep(500 * time.Millisecond)
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				tick = slowtick
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				tick = fasttick
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}
