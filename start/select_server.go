package start

import (
	"fmt"
	"io"
	"strings"

	"github.com/xackery/rof2plus/serverlist"
)

func selectServer(serverName string) (*serverlist.ServerEntry, error) {
	var err error
	if serverName != "" {
		serverEntry, err := selectServerAttempt(serverName)
		if err == nil {
			return serverEntry, nil
		}
	}
	isServerChosen := false
	for !isServerChosen {
		serverName, err = selectServerPrompt()
		if err != nil {
			fmt.Println(err)
			continue
		}
		if serverName == "" {
			fmt.Println("Invalid server name. Please try again.")
			continue
		}
		serverEntry, err := selectServerAttempt(serverName)
		if err == nil {
			return serverEntry, nil
		}
		return serverEntry, nil
	}
	return nil, fmt.Errorf("failed to select server")
}

func selectServerAttempt(serverName string) (*serverlist.ServerEntry, error) {
	server, err := serverlist.ByShortName(serverName)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	return server, nil
}

func selectServerPrompt() (string, error) {
	servers := serverlist.Servers()
	for _, server := range servers {
		fmt.Printf("Server: %s\n", server.ShortName)
	}
	fmt.Printf("Please select a server to connect to: ")
	var selectedServer string
	_, err := fmt.Scanln(&selectedServer)
	if err != nil {
		if err == io.EOF {
			return "", nil
		}

		return "", fmt.Errorf("scan server: %w", err)
	}
	selectedServer = strings.TrimSpace(selectedServer)
	if selectedServer == "" {
		return "", fmt.Errorf("invalid server name")
	}
	return selectedServer, nil
}
