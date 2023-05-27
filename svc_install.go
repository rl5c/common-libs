package wsvc

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

func exePath() (string, error) {

	var err error
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}

	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}

	if filepath.Ext(p) == "" {
		var fi os.FileInfo
		p += ".exe"
		fi, err = os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}

func installService(name string, display string, description string) error {

	exePath, err := exePath()
	if err != nil {
		return err
	}

	svcMgr, err := mgr.Connect()
	if err != nil {
		return err
	}

	defer svcMgr.Disconnect()
	service, err := svcMgr.OpenService(name)
	if err == nil {
		service.Close()
		return fmt.Errorf("service %s already installed", name)
	}

	service, err = svcMgr.CreateService(
		name,
		exePath,
		mgr.Config{
			DisplayName: display,
			StartType:   mgr.StartAutomatic,
			Description: description,
		},
	)

	if err != nil {
		return err
	}

	defer service.Close()
	eventlog.Remove(name)
	if err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info); err != nil {
		service.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	return nil
}
