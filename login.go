package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"time"
)

func autoLogin() {
	rand.Seed(time.Now().UnixNano())

	for {
		logger.Infof("[LOGIN] login to binarium")
		err := login()
		if err != nil {
			logger.Error(err)
		}
		loginDelay := 10 * time.Second * time.Duration(rand.Intn(600)+180)
		logger.Infof("[LOGIN] relogin after '%s'", loginDelay)
		time.Sleep(loginDelay)
	}
}

func login() error {
	cmd := exec.Command(
		"/home/operator/sources/binarium/binarium",
		"--login",
		"--email", "",
		"--password", "",
	)

	binariumLock.Lock()
	logger.Debug(cmd)
	out, err := cmd.CombinedOutput()
	logger.Debug(string(out))
	binariumLock.Unlock()
	if err != nil {
		return fmt.Errorf("can't login: '%s'", err.Error())
	}

	return nil
}
