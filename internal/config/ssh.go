package config

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func ParseSSHConfig(path string) ([]Connection, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var conns []Connection
	var current *Connection

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		keyword := strings.ToLower(fields[0])
		value := fields[1]

		switch keyword {
		case "host":
			if current != nil {
				conns = append(conns, *current)
			}
			if value == "*" {
				current = nil
				continue
			}
			current = &Connection{
				Name:   value,
				Port:   "22",
				Type:   TypeSSH,
				Host:   value,
				Uptime: "-",
			}
		case "hostname":
			if current != nil {
				current.Host = value
			}
		case "port":
			if current != nil {
				current.Port = value
			}
		case "user":
			if current != nil {
				current.User = value
			}
		}
	}

	if current != nil {
		conns = append(conns, *current)
	}

	return conns, scanner.Err()
}

func detectActiveSSHUptime(conns []Connection) []Connection {
	out, err := exec.Command("pgrep", "-a", "-f", "^ssh ").Output()
	if err != nil {
		return conns
	}

	type activeInfo struct {
		host string
		pid  string
	}
	var active []activeInfo

	hostRe := regexp.MustCompile(`\bssh\b.*\b(\S+)$`)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid := fields[0]
		matches := hostRe.FindStringSubmatch(strings.Join(fields[1:], " "))
		if len(matches) > 1 {
			active = append(active, activeInfo{host: matches[1], pid: pid})
		}
	}

	for _, a := range active {
		uptime := getProcessUptime(a.pid)
		if uptime != "" {
			for i := range conns {
				if conns[i].Host == a.host || conns[i].Name == a.host {
					conns[i].Pid = a.pid
					conns[i].Uptime = formatUptime(uptime)
				}
			}
		}
	}

	return conns
}

func getProcessUptime(pid string) string {
	out, err := exec.Command("ps", "-o", "etime=", "-p", pid).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func formatUptime(elapsed string) string {
	elapsed = strings.TrimSpace(elapsed)
	if elapsed == "" {
		return "-"
	}

	totalSeconds := parseElapsed(elapsed)
	if totalSeconds < 3600 {
		return "< 1h"
	}

	days := totalSeconds / 86400
	hours := (totalSeconds % 86400) / 3600
	minutes := (totalSeconds % 3600) / 60

	if days > 0 {
		return "P" + strconv.FormatInt(days, 10) + "DT" + strconv.FormatInt(hours, 10) + "H" + strconv.FormatInt(minutes, 10) + "M"
	}
	return "PT" + strconv.FormatInt(hours, 10) + "H" + strconv.FormatInt(minutes, 10) + "M"
}

func parseElapsed(s string) int64 {
	parts := strings.Split(s, "-")
	if len(parts) == 2 {
		days, _ := strconv.ParseInt(parts[0], 10, 64)
		timeParts := strings.Split(parts[1], ":")
		hours, _ := strconv.ParseInt(timeParts[0], 10, 64)
		minutes, _ := strconv.ParseInt(timeParts[1], 10, 64)
		return days*86400 + hours*3600 + minutes*60
	}

	timeParts := strings.Split(s, ":")
	if len(timeParts) == 3 {
		hours, _ := strconv.ParseInt(timeParts[0], 10, 64)
		minutes, _ := strconv.ParseInt(timeParts[1], 10, 64)
		seconds, _ := strconv.ParseInt(timeParts[2], 10, 64)
		return hours*3600 + minutes*60 + seconds
	}
	return 0
}

func LoadAllSSHConnections() []Connection {
	var all []Connection

	home, err := os.UserHomeDir()
	if err == nil {
		userConfig := filepath.Join(home, ".ssh", "config")
		conns, err := ParseSSHConfig(userConfig)
		if err == nil {
			all = append(all, conns...)
		}
	}

	systemConns, err := ParseSSHConfig("/etc/ssh/ssh_config")
	if err == nil {
		all = append(all, systemConns...)
	}

	all = detectActiveSSHUptime(all)

	return all
}
