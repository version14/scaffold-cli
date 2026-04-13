package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	username string
	email    string
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			MarginTop(1).
			MarginBottom(1)

	subHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			MarginBottom(2)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")).
			PaddingLeft(2).
			PaddingRight(2)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)
)

func printHeader() {
	header := `
██╗   ██╗ ██╗██╗  ██╗      ███████╗ ██████╗ █████╗ ███████╗███████╗ ██████╗ ██╗     ██████╗
██║   ██║███║██║  ██║      ██╔════╝██╔════╝██╔══██╗██╔════╝██╔════╝██╔═══██╗██║     ██╔══██╗
██║   ██║╚██║███████║█████╗███████╗██║     ███████║█████╗  █████╗  ██║   ██║██║     ██║  ██║
╚██╗ ██╔╝ ██║╚════██║╚════╝╚════██║██║     ██╔══██║██╔══╝  ██╔══╝  ██║   ██║██║     ██║  ██║
 ╚████╔╝  ██║     ██║      ███████║╚██████╗██║  ██║██║     ██║     ╚██████╔╝███████╗██████╔╝
  ╚═══╝   ╚═╝     ╚═╝      ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝     ╚═╝      ╚═════╝ ╚══════╝╚═════╝
`
	fmt.Println(header)
	fmt.Println(headerStyle.Render("✨ Project Scaffolder"))
	fmt.Println(subHeaderStyle.Render("Generate production-ready projects in seconds"))
	fmt.Println()
}

func main() {
	printHeader()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("👤 Username").
				Placeholder("john_doe").
				Description("Your GitHub username or preferred name").
				Value(&username).
				Validate(func(s string) error {
					if len(s) < 3 {
						return fmt.Errorf("username must be at least 3 characters")
					}
					return nil
				}),
		).Description("Let's get started! Tell us about yourself."),

		huh.NewGroup(
			huh.NewInput().
				Title("📧 Email").
				Placeholder("you@example.com").
				Description("Your email address for the project").
				Value(&email).
				Validate(func(s string) error {
					if len(s) < 5 || !strings.Contains(s, "@") {
						return fmt.Errorf("please enter a valid email address")
					}
					return nil
				}),
		).Description("Where should we send updates?"),
	).
		WithTheme(huh.ThemeDracula())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println(successStyle.Render("✓ Profile Created Successfully!"))
	fmt.Println()
	fmt.Printf("  Username: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(username))
	fmt.Printf("  Email:    %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(email))
	fmt.Println()
	fmt.Println(infoStyle.Render("🚀 Ready to scaffold your project! (coming soon)"))
	fmt.Println()
}

