package gittp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type StreamCode string

const (
	PackDataStreamCode = StreamCode("\u0001")
	ProgressStreamCode = StreamCode("\u0002")
	FatalStreamCode    = StreamCode("\u0003")
	DefaultStreamCode  = ProgressStreamCode
)

var (
	serviceRegexp         = regexp.MustCompile("(?:/info/refs\\?service=|/)(git-(?:receive|upload)-pack)$")
	noMatchingServiceErr  = errors.New("No matching service types found")
	couldNotCreateRepoErr = errors.New("Could not create repository")
)

func detectServiceType(url *url.URL) (string, error) {
	match := serviceRegexp.FindStringSubmatch(url.RequestURI())
	if len(match) < 2 {
		return "", noMatchingServiceErr
	}

	return match[1], nil
}

func runCmd(pack serviceType, repoPath string, input io.Reader, output io.Writer, advertise bool) error {
	args := []string{"--stateless-rpc"}

	if advertise {
		args = append(args, "--advertise-refs", repoPath)
	} else {
		args = append(args, repoPath)
	}

	cmd := exec.Command(pack.String(), args...)

	cmd.Dir = repoPath
	cmd.Stdin = input
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func parseRepoName(requestPath string) (string, error) {
	paths := strings.Split(requestPath, ".git")

	if len(paths) <= 1 {
		return "", errors.New("pushed url needs to be in user/project.git format\n" + requestPath)
	}

	return fmt.Sprintf("%s.git", paths[0]), nil
}

type ReceivePackResult struct {
	OldRef       string
	NewRef       string
	Branch       string
	Agent        string
	Capabilities []string
}

func newReceivePackResult(packetData []byte) ReceivePackResult {
	parsedPacketData, _ := readPacket(packetData)
	splits := bytes.Split(parsedPacketData, []byte("\x00"))

	pushInfo := strings.Split(string(splits[0]), " ")
	capabilities := strings.Split(string(splits[1]), " ")
	capLen := len(capabilities) - 1

	return ReceivePackResult{
		OldRef:       pushInfo[0],
		NewRef:       pushInfo[1],
		Branch:       pushInfo[2],
		Capabilities: capabilities[:capLen],
		Agent:        strings.TrimSuffix(capabilities[capLen], "0000PACK"),
	}
}

func initRepository(repoPath string) error {
	if !fileExists(repoPath) {
		if err := os.MkdirAll(repoPath, os.ModePerm|os.ModeDir); err != nil {
			return err
		}
	}

	cmd := exec.Command("git", "init", "--bare", repoPath)

	return cmd.Run()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return os.IsExist(err)
}

func handleMissingRepo(serviceStr serviceType, repoName string) error {
	if serviceStr.isReceivePack() && initRepository(repoName) != nil {
		return couldNotCreateRepoErr
	}

	return nil
}

// TODO needs tests
func writePacket(payload string) []byte {
	length := uint32(len(payload) + 4)
	return []byte(fmt.Sprintf("%04x%s0000", length, payload))
}

// TODO needs tests
func readPacket(packetData []byte) ([]byte, error) {
	buf := bytes.NewBuffer(packetData)

	packetLengthBytes := buf.Next(4)

	var packetLength int64
	var err error
	if packetLength, err = strconv.ParseInt(string(packetLengthBytes), 16, 32); err != nil {
		return nil, err
	}

	return buf.Next(int(packetLength)), nil
}

func encode(message string) string {
	return encodeWithPrefix(DefaultStreamCode, message)
}

func encodeWithPrefix(streamCode StreamCode, message string) string {
	packet := fmt.Sprintf("%s%s", streamCode, message)
	return fmt.Sprintf("%04X%s", len(packet)+4, packet)
}
