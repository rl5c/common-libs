package wsvc

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func controlService(name string, c svc.Cmd, to svc.State) error {

	svcMgr, err := mgr.Connect()
	if err != nil {
		return err
	}

	defer svcMgr.Disconnect()
	service, err := svcMgr.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}

	defer service.Close()
	status, err := service.Control(c)
	if err != nil {
		return fmt.Errorf("could not send control=%d: %v", c, err)
	}

	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = service.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}
	return nil
}
