package main

import (
	"log"
	"math/rand"
	"os/exec"
	"time"
)

func autoLogin() {
	rand.Seed(time.Now().UnixNano())

	for {
		login()
		loginDelay := 10 * time.Second * time.Duration(rand.Intn(600)+180)
		logger.Infof("[LOGIN] relogin after '%s'", loginDelay)
		time.Sleep(loginDelay)
	}
}

func login() {
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
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}
