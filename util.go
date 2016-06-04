package gittp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type streamCode []byte

var (
	parseRepoNameRegexp   = regexp.MustCompile("((?:/info/refs\\?service=|/)(?:git-(?:receive|upload)-pack)$)")
	errCouldNotCreateRepo = errors.New("Could not create repository")
	errCouldNotGetArchive = errors.New("Could not get an archive of the pushed refs")
	errNotAGitRequest     = errors.New("requested url did not come from a git client")

	packDataStreamCode = streamCode([]byte("\u0001"))
	progressStreamCode = streamCode([]byte("\u0002"))
	fatalStreamCode    = streamCode([]byte("\u0003"))
	defaultStreamCode  = progressStreamCode
)

func parseRepoName(requestPath string) (string, error) {
	match := parseRepoNameRegexp.FindStringSubmatch(requestPath)
	if len(match) < 2 {
		return "", errNotAGitRequest
	}

	path := strings.TrimSuffix(requestPath, match[1])
	return strings.TrimPrefix(path, "/"), nil
}

func gitArchive(fullRepoPath, hash string) (io.Reader, error) {
	cmd := exec.Command("git", "archive", hash)
	cmd.Dir = fullRepoPath
	cmd.Stderr = os.Stdin

	tarArchive, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return nil, errCouldNotGetArchive
	}

	return bytes.NewBuffer(tarArchive), nil
}

func runCmd(pack string, repoPath string, input io.Reader, output io.Writer, advertise bool) error {
	args := []string{"--stateless-rpc"}

	if advertise {
		args = append(args, "--advertise-refs")
	}

	args = append(args, repoPath)

	cmd := exec.Command(string(pack), args...)

	cmd.Dir = repoPath
	cmd.Stdin = input
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// receivePackResult represents the payload of git-send-pack
type receivePackResult struct {
	OldRef       string
	NewRef       string
	Branch       string
	Agent        string
	Capabilities []string
}

// TODO needs tests
func newReceivePackResult(packData []byte) receivePackResult {
	parsedPacketData, _ := readPackInfo(packData)
	splits := bytes.Split(parsedPacketData, []byte("\x00"))

	pushInfo := strings.Split(string(splits[0]), " ")
	capabilities := strings.Split(string(splits[1]), " ")
	capLen := len(capabilities) - 1

	return receivePackResult{
		OldRef:       pushInfo[0],
		NewRef:       pushInfo[1],
		Branch:       strings.TrimPrefix(pushInfo[2], "refs/heads/"),
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

// TODO needs tests
func writePacket(payload string) []byte {
	length := uint32(len(payload) + 4)
	return []byte(fmt.Sprintf("%04x%s0000", length, payload))
}

// TODO needs tests
func readPackInfo(packetData []byte) ([]byte, error) {
	buf := bytes.NewBuffer(packetData)

	packetLengthBytes := buf.Next(4)

	var packetLength int64
	var err error
	if packetLength, err = strconv.ParseInt(string(packetLengthBytes), 16, 32); err != nil {
		return nil, err
	}
	return buf.Next(int(packetLength)), nil
}

func encodeBytes(streamCode streamCode, msg []byte) []byte {
	packet := append(streamCode, msg...)
	packetLength := fmt.Sprintf("%04X", len(packet)+4)
	return append([]byte(packetLength), packet...)
}
