package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	tea "charm.land/bubbletea/v2"

	"github.com/hech/mash/internal/tui"
	"github.com/hech/mash/internal/vault"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "add-pwd" {
		runAddPwd(os.Args[2:])
		return
	}

	m := tui.NewModel()
	tui.LoadConnections(&m)

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runAddPwd(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mash add-pwd <connection-name> <username>\n")
		os.Exit(1)
	}

	name := args[0]
	username := args[1]

	fmt.Printf("Password for %s (%s): ", name, username)
	passwd, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading password: %v\n", err)
		os.Exit(1)
	}
	if len(passwd) == 0 {
		fmt.Fprintln(os.Stderr, "password must not be empty")
		os.Exit(1)
	}

	if err := vault.Store(name, vault.Credential{
		Username: username,
		Password: strings.TrimSpace(string(passwd)),
	}); err != nil {
		fmt.Fprintf(os.Stderr, "error storing credential: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Credential stored for %q\n", name)
}
