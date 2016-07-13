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

var (
	parseRepoNameRegexp   = regexp.MustCompile("((?:/info/refs\\?service=|/)(?:git-(?:receive|upload)-pack)$)")
	errCouldNotCreateRepo = errors.New("Could not create repository")
	errCouldNotGetArchive = errors.New("Could not get an archive of the pushed refs")
	errNotAGitRequest     = errors.New("requested url did not come from a git client")
	null                  = []byte("\x00")
)

// packetHeader represents the payload of git-send-pack
type packetHeader struct {
	Last         string
	Head         string
	Branch       string
	Agent        string
	Capabilities []string
}

func newPacketHeader(packHeader []byte) packetHeader {
	if len(packHeader) <= 4 {
		return packetHeader{}
	}

	splits := bytes.Split(packHeader[4:], null)
	pushInfo := strings.Split(string(splits[0]), " ")
	capabilities := strings.Split(string(splits[1]), " ")
	capLen := len(capabilities) - 1

	return packetHeader{
		Last:         pushInfo[0],
		Head:         pushInfo[1],
		Branch:       pushInfo[2],
		Capabilities: capabilities[:capLen],
		Agent:        strings.TrimPrefix(strings.TrimSuffix(capabilities[capLen], "0000"), "agent="),
	}
}

// TODO needs tests
func readPackInfo(packetData io.Reader) ([]byte, error) {
	packetLengthBytes := make([]byte, 4)
	_, err := io.ReadFull(packetData, packetLengthBytes)
	if err == io.EOF {
		return []byte{}, nil
	} else if err != nil {
		return []byte{}, errCouldNotReadReqBody
	}

	var packetLength int64
	if packetLength, err = strconv.ParseInt(string(packetLengthBytes), 16, 32); err != nil {
		return nil, err
	}

	// someone just sent a pkt-flush only
	if packetLength == 0 {
		return packetLengthBytes, nil
	}

	rawHeader := make([]byte, packetLength)
	if _, err := io.ReadFull(packetData, rawHeader); err != nil {
		return []byte{}, fmt.Errorf("Could not read %v length\n%v", packetLength, errCouldNotReadReqBody)
	}

	return append(packetLengthBytes, rawHeader...), nil
}

// TODO needs tests
func encodeWithPrefix(psc streamCode, data string) []byte {
	if len(data) == 0 {
		return []byte("0000")
	}

	return pktline(fmt.Sprintf("%s%s", psc, data))
}

func pktline(msg string) []byte {
	if len(msg) == 0 {
		return []byte("0000")
	}

	msgBytes := []byte(msg)
	packetLength := fmt.Sprintf("%04x", len(msgBytes)+4)
	return append([]byte(packetLength), msgBytes...)
}

func parseRepoName(requestPath string) (string, error) {
	match := parseRepoNameRegexp.FindStringSubmatch(requestPath)
	if len(match) < 2 {
		return "", errNotAGitRequest
	}

	path := strings.TrimSuffix(requestPath, match[1])
	return strings.TrimPrefix(path, "/"), nil
}

func gitArchive(fullRepoPath, hash string) ([]byte, error) {
	cmd := exec.Command("git", "archive", hash)
	cmd.Dir = fullRepoPath
	cmd.Stderr = os.Stdout

	tarArchive, err := cmd.Output()
	if err != nil {
		return nil, errCouldNotGetArchive
	}

	return tarArchive, nil
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

	return cmd.Run()
}

func initRepository(repoPath string) error {
	if _, err := os.Stat(repoPath); !!os.IsNotExist(err) {
		if err := os.MkdirAll(repoPath, os.ModePerm|os.ModeDir); err != nil {
			return err
		}

		cmd := exec.Command("git", "init", "--bare", repoPath)

		return cmd.Run()
	}

	return nil
}
