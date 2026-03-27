package version

import (
	"os"
	"strings"
)

var embeddedVersion string

func SetVersion(version string) {
	embeddedVersion = version
}

func Get() string {
	if embeddedVersion != "" {
		return strings.TrimSpace(embeddedVersion)
	}

	data, err := os.ReadFile("VERSION")
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(data))
}
