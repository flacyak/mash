package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// OpenTofuState represents the top-level JSON output of "tofu show -json".
type OpenTofuState struct {
	Values struct {
		RootModule struct {
			Resources []TofuResource `json:"resources"`
		} `json:"root_module"`
	} `json:"values"`
}

// TofuResource is a single managed resource from the OpenTofu state.
type TofuResource struct {
	Type    string                `json:"type"`
	Name    string                `json:"name"`
	Values  json.RawMessage       `json:"values"`
	Address string                `json:"address"`
}

// ec2InstanceValues mirrors the attributes produced by aws_instance.
type ec2InstanceValues struct {
	PublicIP    string            `json:"public_ip"`
	PrivateIP   string            `json:"private_ip"`
	Tags        map[string]string `json:"tags"`
}

// gcpInstanceValues mirrors google_compute_instance.
type gcpInstanceValues struct {
	NetworkInterface []struct {
		AccessConfig []struct {
			NatIP string `json:"nat_ip"`
		} `json:"access_config"`
		NetworkIP string `json:"network_ip"`
	} `json:"network_interface"`
	Name string `json:"name"`
}

// azureInstanceValues mirrors azurerm_linux_virtual_machine.
type azureInstanceValues struct {
	PublicIPAddress  string            `json:"public_ip_address"`
	PrivateIPAddress string            `json:"private_ip_address"`
	ComputerName     string            `json:"computer_name"`
	Tags             map[string]string `json:"tags"`
}

// DiscoverCloudConnections parses an OpenTofu state JSON file and returns
// Connection entries for every cloud VM instance found. If statePath is
// empty it shells out to "tofu show -json".
func DiscoverCloudConnections(statePath string) ([]Connection, error) {
	var raw []byte
	var err error

	if statePath != "" {
		raw, err = os.ReadFile(statePath)
	} else {
		raw, err = exec.Command("tofu", "show", "-json").Output()
	}
	if err != nil {
		return nil, fmt.Errorf("tofu state: %w", err)
	}

	var state OpenTofuState
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("parse tofu state: %w", err)
	}

	var conns []Connection

	for _, res := range state.Values.RootModule.Resources {
		switch res.Type {
		case "aws_instance":
			c := parseEC2(res)
			if c != nil {
				conns = append(conns, *c)
			}
		case "google_compute_instance":
			c := parseGCP(res)
			if c != nil {
				conns = append(conns, *c)
			}
		case "azurerm_linux_virtual_machine", "azurerm_windows_virtual_machine":
			c := parseAzure(res)
			if c != nil {
				conns = append(conns, *c)
			}
		}
	}

	return conns, nil
}

func parseEC2(res TofuResource) *Connection {
	var v ec2InstanceValues
	if err := json.Unmarshal(res.Values, &v); err != nil {
		return nil
	}

	host := v.PublicIP
	if host == "" {
		host = v.PrivateIP
	}
	if host == "" {
		return nil
	}

	name := res.Name
	if tagName, ok := v.Tags["Name"]; ok && tagName != "" {
		name = tagName
	}

	return &Connection{
		Name:   name,
		Port:   "22",
		Type:   TypeEC2,
		Host:   host,
		User:   "ec2-user",
		Uptime: "-",
	}
}

func parseGCP(res TofuResource) *Connection {
	var v gcpInstanceValues
	if err := json.Unmarshal(res.Values, &v); err != nil {
		return nil
	}

	if len(v.NetworkInterface) == 0 {
		return nil
	}

	host := v.NetworkInterface[0].NetworkIP
	if len(v.NetworkInterface[0].AccessConfig) > 0 &&
		v.NetworkInterface[0].AccessConfig[0].NatIP != "" {
		host = v.NetworkInterface[0].AccessConfig[0].NatIP
	}

	name := v.Name
	if name == "" {
		name = res.Name
	}

	return &Connection{
		Name:   name,
		Port:   "22",
		Type:   TypeGCP,
		Host:   host,
		User:   "gcp-user",
		Uptime: "-",
	}
}

func parseAzure(res TofuResource) *Connection {
	var v azureInstanceValues
	if err := json.Unmarshal(res.Values, &v); err != nil {
		return nil
	}

	host := v.PublicIPAddress
	if host == "" {
		host = v.PrivateIPAddress
	}
	if host == "" {
		return nil
	}

	name := v.ComputerName
	if name == "" {
		if tagName, ok := v.Tags["Name"]; ok && tagName != "" {
			name = tagName
		}
	}
	if name == "" {
		name = res.Name
	}

	return &Connection{
		Name:   name,
		Port:   "22",
		Type:   TypeAzure,
		Host:   host,
		User:   "azureuser",
		Uptime: "-",
	}
}

// cloudConnType maps a resource type string from OpenTofu to a ConnType.
func cloudConnType(resType string) ConnType {
	switch resType {
	case "aws_instance":
		return TypeEC2
	case "google_compute_instance":
		return TypeGCP
	case "azurerm_linux_virtual_machine", "azurerm_windows_virtual_machine":
		return TypeAzure
	default:
		return TypeSSH
	}
}

// CloudProviderName returns a human-readable provider name for a ConnType.
func CloudProviderName(ct ConnType) string {
	switch ct {
	case TypeEC2:
		return "AWS EC2"
	case TypeGCP:
		return "GCP Compute"
	case TypeAzure:
		return "Azure VM"
	case TypeTailscale:
		return "Tailscale"
	default:
		return string(ct)
	}
}
