package config

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func DiscoverMoshConnections() []Connection {
	var conns []Connection

	conns = append(conns, discoverViaProcesses()...)
	conns = append(conns, discoverViaUFW()...)
	conns = append(conns, discoverViaNetstat()...)

	return deduplicate(conns)
}

func discoverViaProcesses() []Connection {
	data, err := exec.Command("pgrep", "-a", "mosh-server").Output()
	if err != nil {
		return nil
	}

	var conns []Connection
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pid := fields[0]
		var port, host string
		host = "localhost"

		for i, f := range fields {
			if f == "-p" && i+1 < len(fields) {
				port = fields[i+1]
			}
		}

		uptimeRaw := getProcessUptime(pid)
		uptime := formatUptime(uptimeRaw)
		if uptime == "" {
			uptime = "-"
		}

		if port == "" {
			continue
		}

		conns = append(conns, Connection{
			Name:   fmt.Sprintf("mosh-server (pid %s)", pid),
			Port:   port,
			Type:   TypeMosh,
			Host:   host,
			Pid:    pid,
			Uptime: uptime,
		})
	}
	return conns
}

func discoverViaUFW() []Connection {
	paths := []string{
		"/etc/ufw/user.rules",
		"/etc/ufw/user6.rules",
	}

	portRe := regexp.MustCompile(`\b(\d{4,5})/udp\b`)

	var conns []Connection
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		idx := 0
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "mosh") || strings.Contains(line, "Mosh") {
				matches := portRe.FindStringSubmatch(line)
				if len(matches) > 1 {
					idx++
					conns = append(conns, Connection{
						Name:   fmt.Sprintf("mosh-ufw-%d", idx),
						Port:   matches[1],
						Type:   TypeMosh,
						Host:   "0.0.0.0",
						Uptime: "-",
					})
				}
			}
		}
		f.Close()
	}
	return conns
}

func discoverViaNetstat() []Connection {
	out, err := exec.Command("ss", "-ulpn").Output()
	if err != nil {
		return nil
	}

	portRe := regexp.MustCompile(`:(\d{4,5})\b`)

	var conns []Connection
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	idx := 0
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "mosh") {
			continue
		}
		matches := portRe.FindStringSubmatch(line)
		if len(matches) > 1 {
			idx++
			conns = append(conns, Connection{
				Name:   fmt.Sprintf("mosh-net-%d", idx),
				Port:   matches[1],
				Type:   TypeMosh,
				Host:   "0.0.0.0",
				Uptime: "-",
			})
		}
	}
	return conns
}

func deduplicate(conns []Connection) []Connection {
	seen := make(map[string]bool)
	var result []Connection
	for _, c := range conns {
		key := c.Port + ":" + c.Host
		if !seen[key] {
			seen[key] = true
			result = append(result, c)
		}
	}
	return result
}
