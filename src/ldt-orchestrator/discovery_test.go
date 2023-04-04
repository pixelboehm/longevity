package ldtorchestrator

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	setupDiscoveryConfig()
	code := m.Run()
	teardownDiscoveryConfig()
	os.Exit(code)
}

var c *DiscoveryConfig = nil

func setupDiscoveryConfig() {
	d1 := []byte("https://github.com/pixelboehm/ldt\n#https://github.com/pixelboehm/longevity\n")
	if err := os.WriteFile("./config", d1, 0644); err != nil {
		panic(err)
	}
	c = &DiscoveryConfig{
		repository_file: "./config",
	}
}

func teardownDiscoveryConfig() {
	os.Remove("./config")
}

func ensureConfigExists(t *testing.T) {
	require := require.New(t)
	if c != nil {
		require.FileExists(c.repository_file)
	} else {
		require.Fail("DiscoveryConfig is not initialized")
	}
}
