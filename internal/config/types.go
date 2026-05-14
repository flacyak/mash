package config

type ConnType string

const (
	TypeSSH  ConnType = "SSH"
	TypeMosh ConnType = "Mosh"
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
