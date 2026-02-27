package app

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/phenoml/phenostore-example-go/fhir"
)

var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
var timingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)

func mapStr(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}

func isAbort(err error) bool {
	return errors.Is(err, huh.ErrUserAborted)
}

// PickPatient fetches all patients and presents a filterable select.
// Returns ("", nil) if no patients exist.
func (a *App) PickPatient() (string, error) {
	ctx := context.Background()
	var patients []json.RawMessage
	var fetchErr error

	err := spinner.New().
		Title("Loading patients...").
		Action(func() {
			patients, fetchErr = a.fetchAllPatients(ctx)
		}).
		Run()
	if err != nil {
		return "", err
	}
	if fetchErr != nil {
		return "", fetchErr
	}

	if len(patients) == 0 {
		fmt.Println("\n  No patients found. Try seeding sample data first.")
		return "", nil
	}

	var options []huh.Option[string]
	for _, raw := range patients {
		m, err := fhir.Parse(raw)
		if err != nil {
			continue
		}
		id := fhir.ResourceID(raw)
		name := fhir.PatientName(m)
		dob := mapStr(m, "birthDate")
		label := fmt.Sprintf("%s (%s)", name, dob)
		options = append(options, huh.NewOption(label, id))
	}

	var patientID string
	err = huh.NewSelect[string]().
		Title("Select a patient").
		Options(options...).
		Value(&patientID).
		Filtering(true).
		Run()

	return patientID, err
}

// PickCarePlan fetches active care plans for a patient and presents a select.
// Returns ("", nil) if no plans exist.
func (a *App) PickCarePlan(patientID string) (string, error) {
	ctx := context.Background()
	var plans []json.RawMessage
	var fetchErr error

	err := spinner.New().
		Title("Loading care plans...").
		Action(func() {
			plans, fetchErr = a.searchCarePlans(ctx, patientID)
		}).
		Run()
	if err != nil {
		return "", err
	}
	if fetchErr != nil {
		return "", fetchErr
	}

	if len(plans) == 0 {
		fmt.Println("\n  No active care plans found for this patient.")
		return "", nil
	}

	var options []huh.Option[string]
	for _, raw := range plans {
		m, err := fhir.Parse(raw)
		if err != nil {
			continue
		}
		id := mapStr(m, "id")
		title := mapStr(m, "title")
		label := fmt.Sprintf("%s (%s)", title, id[:min(8, len(id))])
		options = append(options, huh.NewOption(label, id))
	}

	if len(options) == 0 {
		fmt.Println("\n  No active care plans found for this patient.")
		return "", nil
	}

	var cpID string
	err = huh.NewSelect[string]().
		Title("Select a care plan").
		Options(options...).
		Value(&cpID).
		Run()

	return cpID, err
}

// PressEnter waits for the user to press enter.
func PressEnter() {
	fmt.Print("\nPress enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// ShowError displays an error message.
func ShowError(err error) {
	fmt.Println(errorStyle.Render("\n  Error: " + err.Error()))
}

// showTiming prints a dimmed timing line after API results.
func showTiming(msg string, d time.Duration) {
	var dur string
	if d < time.Second {
		dur = fmt.Sprintf("%dms", d.Milliseconds())
	} else {
		dur = fmt.Sprintf("%.1fs", d.Seconds())
	}
	fmt.Println(timingStyle.Render(fmt.Sprintf("  %s in %s", msg, dur)))
}
