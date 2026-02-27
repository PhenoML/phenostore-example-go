package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/phenoml/phenostore-example-go/app"
)

func main() {
	a := &app.App{}
	if err := a.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	banner := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("Community Health Clinic — PhenoStore SDK Demo")

	fmt.Println()
	fmt.Println(banner)

	a.MainMenu()
}
