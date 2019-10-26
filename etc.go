package main

import (
	"bytes"
	"io/ioutil"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
)

func resolveHostIP() string {
	netInterfaceAddresses, err := net.InterfaceAddrs()

	if err != nil {
		return ""
	}

	for _, netInterfaceAddress := range netInterfaceAddresses {
		if networkIP, ok := netInterfaceAddress.(*net.IPNet); ok && !networkIP.IP.IsLoopback() && networkIP.IP.To4() != nil {
			return networkIP.IP.String()
		}
	}
	return ""
}

func getProjectLatestVersion(project string) string {
	all := redisConn.LRange(redisVersionsPrefix+project, 0, -1).Val()

	return all[0]
}

func getProjectSpiders(project, version string) ([]string, error) {
	out := bytes.NewBufferString("")
	cmd := exec.Command("python", "-m", "scrapy", "list")
	cmd.Dir = filepath.Join(*flagCacheDir, project, "src", version)
	cmd.Stderr = ioutil.Discard
	cmd.Stdout = out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(out.String()), "\n"), nil
}
