package wsvc

import (
	"fmt"

	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

func removeService(name string) error {

	svcMgr, err := mgr.Connect()
	if err != nil {
		return err
	}

	defer svcMgr.Disconnect()
	service, err := svcMgr.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", name)
	}

	defer service.Close()
	err = service.Delete()
	if err != nil {
		return err
	}

	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}
