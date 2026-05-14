package config

type ConnType string

const (
	TypeSSH       ConnType = "SSH"
	TypeMosh      ConnType = "Mosh"
	TypeEC2       ConnType = "EC2"
	TypeGCP       ConnType = "GCP"
	TypeAzure     ConnType = "Azure"
	TypeTailscale ConnType = "Tailscale"
)

type Connection struct {
	Name   string
	Port   string
	Type   ConnType
	Host   string
	User   string
	Pid    string
	Uptime string
}
