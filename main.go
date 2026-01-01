package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/the9x/anneal/internal/config"
	"github.com/the9x/anneal/internal/jmap"
	"github.com/the9x/anneal/internal/ui"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Check if we have accounts configured
	if len(cfg.Accounts) == 0 {
		if err := setupFirstAccount(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
			os.Exit(1)
		}
	}

	// Get the default account
	account := cfg.DefaultAccount()
	if account == nil {
		fmt.Fprintf(os.Stderr, "No account configured\n")
		os.Exit(1)
	}

	// Get token from keyring
	token, err := config.GetToken(account.Email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "No API token found for %s\n", account.Email)
		fmt.Fprintf(os.Stderr, "Please set your token: tuimail set-token %s <token>\n", account.Email)
		os.Exit(1)
	}

	// Create JMAP client
	client, err := jmap.New(account.Email, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}

	// Create and run the app
	app := ui.NewApp(cfg, client)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func setupFirstAccount(cfg *config.Config) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to TuiMail!")
	fmt.Println("Let's set up your first Fastmail account.")
	fmt.Println()

	// Get account name
	fmt.Print("Account name (e.g., Work, Personal): ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Default"
	}

	// Get email
	fmt.Print("Email address: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Get API token
	fmt.Println()
	fmt.Println("To get your API token:")
	fmt.Println("1. Go to Fastmail Settings → Privacy & Security → Integrations")
	fmt.Println("2. Under 'API tokens', click 'Manage'")
	fmt.Println("3. Create a new token with Mail access")
	fmt.Println()
	fmt.Print("API Token: ")
	token, _ := reader.ReadString('\n')
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("API token is required")
	}

	// Add account
	if err := cfg.AddAccount(name, email, true); err != nil {
		return err
	}

	// Save token to keyring
	if err := config.SetToken(email, token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println("Account configured successfully!")
	fmt.Println()

	return nil
}
