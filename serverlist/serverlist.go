// serverlist manages the list of servers valid for rof2plus
package serverlist

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

var server Server

type Server struct {
	LastUpdate time.Time      `yaml:"lastupdate"`
	Entries    []*ServerEntry `yaml:"entries"`
}

// Server is a server list entry
type ServerEntry struct {
	ShortName string `yaml:"shortname"`
	Name      string `yaml:"name"`
	PatchURL  string `yaml:"patchurl"`
}

// Fetch gets the latest server list
func Fetch() error {
	data, err := os.ReadFile("serverlist.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			err = download()
			if err != nil {
				return fmt.Errorf("download: %w", err)
			}
			data, err = os.ReadFile("serverlist.yaml")
			if err != nil {
				return fmt.Errorf("read download: %w", err)
			}
			return nil
		}

		return fmt.Errorf("read: %w", err)
	}

	err = yaml.Unmarshal(data, &server)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	return nil
}

func download() error {
	// make a mock serverlist
	server.Entries = []*ServerEntry{
		{
			Name:      "Test Server",
			ShortName: "test",
			PatchURL:  "https://example.com/patch",
		},
		{
			Name:      "Another Server",
			ShortName: "another",
			PatchURL:  "https://example.com/anotherpatch",
		},
	}
	server.LastUpdate = time.Now()

	data, err := yaml.Marshal(&server)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	err = os.WriteFile("serverlist.yaml", data, 0644)
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

// ByName returns a server entry by name
func ByShortName(shortName string) (*ServerEntry, error) {
	for _, entry := range server.Entries {
		if entry.ShortName == shortName {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("server %s not found", shortName)
}

// Servers returns the list of servers
func Servers() []*ServerEntry {
	return server.Entries
}
