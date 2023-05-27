package wsvc

import (
	"fmt"

	"golang.org/x/sys/windows/svc/mgr"
)

func startService(name string) error {

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
	if err = service.Start(); err != nil {
		return fmt.Errorf("could not start service: %v", err)
	}
	return nil
}
