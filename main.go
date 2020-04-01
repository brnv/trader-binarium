package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kovetskiy/godocs"
	"github.com/kovetskiy/lorg"
)

var (
	logger  = lorg.NewLog()
	version = "[manual build]"
)

var (
	kindOption = map[string]string{
		"ВВЕРХ": "--up",
		"ВНИЗ":  "--down",
	}

	binariumLock = sync.Mutex{}

	history map[string]bool = make(map[string]bool)

	staleMessageTimeoutSeconds = 2
	simultaneousMessagesMax    = 3
)

const usage = `trader-binarium

Usage:
    trader-binarium [options]
    trader-binarium -h | --help

Options:
    --debug    Enable debug output.
    --trace    Enable trace output.
    -h --help  Show this help.
`

func main() {
	args := godocs.MustParse(usage, version, godocs.UsePager)

	logger.SetIndentLines(true)

	if args["--debug"].(bool) {
		logger.SetLevel(lorg.LevelDebug)
	}

	if args["--trace"].(bool) {
		logger.SetLevel(lorg.LevelTrace)
	}

	go autoLogin()

	cmd := exec.Command("telegram-cli", "--disable-colors")

	stdin, _ := cmd.StdinPipe()

	go func() {
		mutex := sync.Mutex{}
		time.Sleep(time.Second * 1)
		stdin.Write([]byte("dialog_list\n"))
		time.Sleep(time.Second * 4)

		for {
			mutex.Lock()
			stdin.Write([]byte("history AT_BOT_m15\n"))
			mutex.Unlock()
			time.Sleep(time.Second * 3)
		}
	}()

	stdout, _ := cmd.StdoutPipe()

	cmd.Start()

	buf := bufio.NewReader(stdout)
	for {
		line, _, _ := buf.ReadLine()

		splitted := strings.Split(string(line), "\n")
		if len(splitted) > simultaneousMessagesMax {
			splitted = splitted[len(splitted)-simultaneousMessagesMax:]
		}

		for _, message := range splitted {
			re := regexp.MustCompile(`\[([0-9:]+)\](.*)`)

			matches := re.FindStringSubmatch(message)
			if len(matches) != 0 {
				tradingMessage := matches[0]
				parseTraderBotMessage(tradingMessage)
			}
		}

	}
}

func parseTraderBotMessage(message string) {
	re := regexp.MustCompile(`\[([0-9:]+)\](.*)\s([A-Z]+)\s[0-9]+\s[а-я]+\s([А-Я]+)`)

	matches := re.FindStringSubmatch(message)
	if len(matches) == 0 {
		return
	}

	asset := matches[3]
	messageTime := matches[1]
	messageMinutes := convertStringToMinutes(messageTime)
	tradingKey := fmt.Sprintf("%d-%s", messageMinutes, asset)
	if history[tradingKey] {
		return
	}

	history[tradingKey] = true

	now := time.Now()
	currentTime := fmt.Sprintf(
		"%02d:%02d", now.Hour(), now.Minute(),
	)
	currentMinutes := convertStringToMinutes(currentTime)

	if Abs(currentMinutes-messageMinutes) > staleMessageTimeoutSeconds {
		logger.Infof("[IGNORE SIGNAL] '%s'", message)
		return
	}

	kind := kindOption[matches[4]]
	logger.Infof("[NEW SIGNAL] '%s' '%s'", asset, kind)

	prevID := ""
	betAgainIfFailed := true
	waitBeforeBet(asset, kind)
	runBinarium(firstBetSum, asset, kind, betAgainIfFailed, prevID)
}

func convertStringToMinutes(time string) int {
	timeParts := strings.Split(time, ":")
	timeHours, _ := strconv.Atoi(timeParts[0])
	timeSeconds, _ := strconv.Atoi(timeParts[1])
	timeSeconds = timeSeconds + timeHours*60

	return timeSeconds
}

func waitBeforeBet(asset string, kind string) {
	rand.Seed(time.Now().UnixNano())
	waitTime := time.Second * time.Duration(rand.Intn(20))
	logger.Infof("[OPERATION] [%s] [%s] wait '%s'", asset, kind, waitTime)
	time.Sleep(waitTime)
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
