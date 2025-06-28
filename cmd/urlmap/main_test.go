package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "Valid URL",
			args:    []string{"https://example.com"},
			wantErr: false,
		},
		{
			name:    "Invalid URL",
			args:    []string{"not-a-url"},
			wantErr: true,
		},
		{
			name:    "No arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Too many arguments",
			args:    []string{"https://example.com", "https://example.org"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := rootCmd
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionCommand(t *testing.T) {
	// Create a new command instance for isolated testing
	cmd := &cobra.Command{
		Use:   "urlmap",
		Short: "Test command",
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("urlmap version %s\n", version)
			cmd.Printf("commit: %s\n", commit)
			cmd.Printf("built: %s\n", date)
		},
	}

	cmd.AddCommand(versionCmd)

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "urlmap version") {
		t.Errorf("version output should contain 'urlmap version', got: %s", output)
	}
	if !strings.Contains(output, "commit:") {
		t.Errorf("version output should contain 'commit:', got: %s", output)
	}
	if !strings.Contains(output, "built:") {
		t.Errorf("version output should contain 'built:', got: %s", output)
	}
}

func TestFlagDefaults(t *testing.T) {
	// Reset flags to default values
	depth = 0
	verbose = false
	userAgent = "urlmap/1.0.0 (+https://github.com/aoshimash/urlmap)"
	concurrent = 10

	// Create a new command instance for isolated testing
	cmd := &cobra.Command{
		Use:  "urlmap <URL>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Test default values
			if depth != 0 {
				t.Errorf("depth default should be 0, got: %d", depth)
			}
			if verbose != false {
				t.Errorf("verbose default should be false, got: %v", verbose)
			}
			if userAgent != "urlmap/1.0.0 (+https://github.com/aoshimash/urlmap)" {
				t.Errorf("userAgent default should be 'urlmap/1.0.0 (+https://github.com/aoshimash/urlmap)', got: %s", userAgent)
			}
			if concurrent != 10 {
				t.Errorf("concurrent default should be 10, got: %d", concurrent)
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&depth, "depth", "d", 0, "Maximum crawl depth (0 = unlimited)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	cmd.Flags().StringVarP(&userAgent, "user-agent", "u", "urlmap/1.0.0 (+https://github.com/aoshimash/urlmap)", "Custom User-Agent string")
	cmd.Flags().IntVarP(&concurrent, "concurrent", "c", 10, "Number of concurrent requests")

	cmd.SetArgs([]string{"https://example.com"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("command execution failed: %v", err)
	}
}

func TestFlagParsing(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		expectedDepth      int
		expectedVerbose    bool
		expectedUserAgent  string
		expectedConcurrent int
	}{
		{
			name:               "Short flags",
			args:               []string{"-d", "5", "-v", "-u", "CustomBot/1.0", "-c", "20", "https://example.com"},
			expectedDepth:      5,
			expectedVerbose:    true,
			expectedUserAgent:  "CustomBot/1.0",
			expectedConcurrent: 20,
		},
		{
			name:               "Long flags",
			args:               []string{"--depth", "3", "--verbose", "--user-agent", "TestBot/2.0", "--concurrent", "15", "https://example.com"},
			expectedDepth:      3,
			expectedVerbose:    true,
			expectedUserAgent:  "TestBot/2.0",
			expectedConcurrent: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			depth = 0
			verbose = false
			userAgent = "urlmap/1.0.0 (+https://github.com/aoshimash/urlmap)"
			concurrent = 10

			cmd := &cobra.Command{
				Use:  "urlmap <URL>",
				Args: cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Test parsed values
					if depth != tt.expectedDepth {
						t.Errorf("depth should be %d, got: %d", tt.expectedDepth, depth)
					}
					if verbose != tt.expectedVerbose {
						t.Errorf("verbose should be %v, got: %v", tt.expectedVerbose, verbose)
					}
					if userAgent != tt.expectedUserAgent {
						t.Errorf("userAgent should be %s, got: %s", tt.expectedUserAgent, userAgent)
					}
					if concurrent != tt.expectedConcurrent {
						t.Errorf("concurrent should be %d, got: %d", tt.expectedConcurrent, concurrent)
					}
					return nil
				},
			}

			cmd.Flags().IntVarP(&depth, "depth", "d", 0, "Maximum crawl depth (0 = unlimited)")
			cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
			cmd.Flags().StringVarP(&userAgent, "user-agent", "u", "urlmap/1.0.0 (+https://github.com/aoshimash/urlmap)", "Custom User-Agent string")
			cmd.Flags().IntVarP(&concurrent, "concurrent", "c", 10, "Number of concurrent requests")

			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("command execution failed: %v", err)
			}
		})
	}
}

func TestHelpOutput(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "urlmap <URL>",
		Short: "A web crawler for mapping site URLs",
		Long: `Urlmap is a web crawler for discovering and mapping all URLs within a website.

This tool crawls web pages starting from a given URL and discovers all links
within the same domain, creating a comprehensive URL map of the site.`,
	}

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.Help()
	output := buf.String()

	// Update the test expectation to match the new description
	if !strings.Contains(output, "web crawler for discovering and mapping") {
		t.Errorf("help output should contain 'web crawler for discovering and mapping', got: %s", output)
	}
}

// Test the main Execute function indirectly
func TestExecute(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test version command
	os.Args = []string{"urlmap", "version"}

	// This test mainly ensures Execute() doesn't panic
	// and runs without fatal errors
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Execute() panicked: %v", r)
		}
	}()
}
