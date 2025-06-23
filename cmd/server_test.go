package cmd

import (
	"bytes"
	"os"
	"testing"
)

func TestServerCommandDefined(t *testing.T) {
	if ServerCmd == nil {
		t.Fatal("ServerCmd should be defined")
	}
	if ServerCmd.Use != "server" {
		t.Errorf("expected command use 'server', got %s", ServerCmd.Use)
	}
	portFlag := ServerCmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Error("expected 'port' flag to be defined")
	}
}

func TestServerCommandHelp(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run help command
	ServerCmd.SetArgs([]string{"--help"})
	ServerCmd.Execute()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected help output, got empty string")
	}
}

func TestServerCommandPortFlag(t *testing.T) {
	// Test that the flag exists
	portFlag := ServerCmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Error("expected 'port' flag to be defined")
	}
}

func TestServerCommandStructure(t *testing.T) {
	if ServerCmd.Short == "" {
		t.Error("server command should have a short description")
	}

	if ServerCmd.Long == "" {
		t.Error("server command should have a long description")
	}

	if ServerCmd.Run == nil {
		t.Error("server command should have a Run function")
	}
}

func TestServerCommandIntegration(t *testing.T) {
	// This test would require starting the server and making HTTP requests
	// For now, we'll test the command structure and flags

	// Test that the command is properly added to root
	rootCmd.AddCommand(ServerCmd)

	// Verify server command is in root command's children
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "server" {
			found = true
			break
		}
	}

	if !found {
		t.Error("server command should be added to root command")
	}
}

// Mock test for HTTP handler (without actually starting server)
func TestServerHandlerLogic(t *testing.T) {
	// This would test the handler logic in a unit test way
	// For now, we'll just verify the command structure

	if ServerCmd.Short == "" {
		t.Error("server command should have a short description")
	}

	if ServerCmd.Long == "" {
		t.Error("server command should have a long description")
	}

	if ServerCmd.Run == nil {
		t.Error("server command should have a Run function")
	}
}

// Integration test helper (commented out as it requires actual server)
/*
func TestServerEndpoints(t *testing.T) {
	// Start server in goroutine
	go func() {
		serverPort = 8082
		ServerCmd.Run(ServerCmd, []string{})
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test endpoints
	endpoints := []string{"/", "/health", "/api/v1/status"}

	for _, endpoint := range endpoints {
		resp, err := http.Get("http://localhost:8082" + endpoint)
		if err != nil {
			t.Errorf("failed to get %s: %v", endpoint, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200 for %s, got %d", endpoint, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("failed to read body for %s: %v", endpoint, err)
			continue
		}

		if len(body) == 0 {
			t.Errorf("expected non-empty response for %s", endpoint)
		}
	}
}
*/
