package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"regexp"
	"time"
)

var (
	firstBetSum  = "1"
	failedBetSum = "1.5"
)

func runBinarium(
	sum string,
	asset string,
	kind string,
	betAgainIfFailed bool,
	prevID string,
) {
	loggerPrefix := fmt.Sprintf("[OPERATION] [%s] [%s]", asset, kind)
	if prevID != "" {
		loggerPrefix = fmt.Sprintf("[OPERATION] [%s] [%s] [%s|_]", asset, kind, prevID)
	}

	logger.Infof("%s place sum '%s'", loggerPrefix, sum)

	command := getBinariumCommand(sum, asset, kind)
	cmd := exec.Command(command[0], command[1:]...)
	binariumLock.Lock()
	logger.Debug(cmd)
	out, err := cmd.CombinedOutput()
	logger.Debug(string(out))
	binariumLock.Unlock()
	if err != nil {
		logger.Errorf("%s error: '%s'", loggerPrefix, err.Error())
	}

	transactionID := getTransactionID(string(out))
	if transactionID != "" {
		go checkTransaction(transactionID, asset, kind, betAgainIfFailed, prevID)
	}

	loginRequired := checkLoginRequired(string(out))
	if loginRequired {
		logger.Infof("%s relogin required", loggerPrefix)
		err := login()
		if err != nil {
			logger.Error(err)
		} else {
			logger.Infof("%s restart after relogin", loggerPrefix)
			runBinarium(sum, asset, kind, betAgainIfFailed, prevID)
		}
	}
}

func checkTransaction(
	id string,
	asset string,
	kind string,
	betAgainIfFailed bool,
	prevID string,
) {
	loggerPrefix := fmt.Sprintf("[OPERATION] [%s] [%s] [%s]", asset, kind, id)
	if prevID != "" {
		loggerPrefix = fmt.Sprintf("[OPERATION] [%s] [%s] [%s|%s]", asset, kind, prevID, id)
	}

	checkTimeout := time.Minute * 16
	if !betAgainIfFailed {
		checkTimeout = time.Minute * 15
	}
	rand.Seed(time.Now().UnixNano())
	checkTimeout = checkTimeout + time.Second*time.Duration(rand.Intn(10))
	logger.Infof("%s wait check for '%s'", loggerPrefix, checkTimeout)
	time.Sleep(checkTimeout)
	logger.Infof("%s check after '%s'", loggerPrefix, checkTimeout)

	cmd := exec.Command(
		"/home/operator/sources/binarium/binarium",
		"--check-id",
		id,
	)
	binariumLock.Lock()
	logger.Debug(cmd)
	out, err := cmd.CombinedOutput()
	logger.Debug(string(out))
	binariumLock.Unlock()
	if err != nil {
		logger.Errorf("%s error: '%s'", loggerPrefix, err.Error())
		return
	}

	transactionSuccess := getTransactionSuccess(string(out))

	if !transactionSuccess {
		logger.Infof("%s transaction failed", loggerPrefix)

		if betAgainIfFailed {
			logger.Infof("%s bet again on failed", loggerPrefix)
			betAgainIfFailed = false
			runBinarium(failedBetSum, asset, kind, betAgainIfFailed, id)
		}
	}

	if transactionSuccess {
		logger.Infof("%s transaction succeeded", loggerPrefix)
	}
}

func getBinariumCommand(sum string, asset string, kind string) []string {
	command := []string{
		"/home/operator/sources/binarium/binarium",
		"--asset", fmt.Sprintf("%s", asset),
		"--sum", fmt.Sprintf("%s", sum),
		kind,
	}

	return command
}

func getTransactionID(output string) string {
	re := regexp.MustCompile(`.*transaction\sid:\s([0-9]+)`)

	matches := re.FindStringSubmatch(output)

	if len(matches) != 0 {
		return matches[1]
	}

	return ""
}

func getTransactionSuccess(output string) bool {
	re := regexp.MustCompile(`.*Success:\sfail`)

	matches := re.FindStringSubmatch(output)

	if len(matches) != 0 {
		return false
	}

	return true
}

func checkLoginRequired(output string) bool {
	re := regexp.MustCompile(`.*login\srequired`)

	matches := re.FindStringSubmatch(output)

	if len(matches) != 0 {
		return true
	}

	return false
}
