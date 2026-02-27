package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/phenoml/phenostore-example-go/fhir"
)

// RegisterPatient collects patient details via a form and creates the resource.
func (a *App) RegisterPatient() {
	var given, family, dob, gender string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("First name").Value(&given),
			huh.NewInput().Title("Last name").Value(&family),
			huh.NewInput().Title("Date of birth (YYYY-MM-DD)").Value(&dob),
			huh.NewSelect[string]().
				Title("Gender").
				Options(huh.NewOptions("male", "female", "other", "unknown")...).
				Value(&gender),
		),
	)

	if err := form.Run(); err != nil {
		if !isAbort(err) {
			ShowError(err)
		}
		return
	}

	body := fhir.NewPatient(given, family, dob, gender)

	var created json.RawMessage
	var apiErr error

	err := spinner.New().
		Title("Registering patient...").
		Action(func() {
			created, apiErr = a.Client.CreateResource(context.Background(), "Patient", body, nil)
		}).
		Run()

	if err != nil {
		ShowError(err)
		PressEnter()
		return
	}
	if apiErr != nil {
		ShowError(fmt.Errorf("creating patient: %w", apiErr))
		PressEnter()
		return
	}

	id := fhir.ResourceID(created)
	fmt.Printf("\n  Created patient %s %s (ID: %s)\n", given, family, id)
	PressEnter()
}

// ListPatients fetches and displays all patients.
func (a *App) ListPatients() {
	var patients []json.RawMessage
	var fetchErr error
	var elapsed time.Duration

	err := spinner.New().
		Title("Loading patients...").
		Action(func() {
			start := time.Now()
			patients, fetchErr = a.fetchAllPatients(context.Background())
			elapsed = time.Since(start)
		}).
		Run()

	if err != nil {
		ShowError(err)
		PressEnter()
		return
	}
	if fetchErr != nil {
		ShowError(fetchErr)
		PressEnter()
		return
	}

	fmt.Println()
	if len(patients) == 0 {
		fmt.Println("  No patients found.")
	} else {
		fhir.PrintPatientList(patients)
		showTiming(fmt.Sprintf("Fetched %d patients", len(patients)), elapsed)
	}
	PressEnter()
}

// ViewPatient lets the user pick a patient and displays their details.
func (a *App) ViewPatient() {
	patientID, err := a.PickPatient()
	if err != nil {
		if !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}
	if patientID == "" {
		PressEnter()
		return
	}

	var raw json.RawMessage
	var apiErr error
	var elapsed time.Duration

	err = spinner.New().
		Title("Loading patient...").
		Action(func() {
			start := time.Now()
			raw, apiErr = a.Client.ReadResource(context.Background(), "Patient", patientID)
			elapsed = time.Since(start)
		}).
		Run()

	if err != nil {
		ShowError(err)
		PressEnter()
		return
	}
	if apiErr != nil {
		ShowError(fmt.Errorf("reading patient: %w", apiErr))
		PressEnter()
		return
	}

	fmt.Println()
	fhir.PrintPatient(raw)
	showTiming("Loaded patient", elapsed)
	PressEnter()
}

// UpdateContact lets the user pick a patient and update phone/email.
func (a *App) UpdateContact() {
	patientID, err := a.PickPatient()
	if err != nil {
		if !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}
	if patientID == "" {
		PressEnter()
		return
	}

	var phone, email string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Phone number (leave blank to skip)").Value(&phone),
			huh.NewInput().Title("Email address (leave blank to skip)").Value(&email),
		),
	)

	if err := form.Run(); err != nil {
		if !isAbort(err) {
			ShowError(err)
		}
		return
	}

	if phone == "" && email == "" {
		fmt.Println("\n  No changes provided.")
		PressEnter()
		return
	}

	var apiErr error
	err = spinner.New().
		Title("Updating patient...").
		Action(func() {
			ctx := context.Background()

			raw, err := a.Client.ReadResource(ctx, "Patient", patientID)
			if err != nil {
				apiErr = fmt.Errorf("reading patient: %w", err)
				return
			}

			var patient map[string]any
			if err := json.Unmarshal(raw, &patient); err != nil {
				apiErr = fmt.Errorf("parsing patient: %w", err)
				return
			}

			telecoms, _ := patient["telecom"].([]any)
			if phone != "" {
				telecoms = append(telecoms, map[string]any{"system": "phone", "value": phone})
			}
			if email != "" {
				telecoms = append(telecoms, map[string]any{"system": "email", "value": email})
			}
			patient["telecom"] = telecoms

			updated, err := json.Marshal(patient)
			if err != nil {
				apiErr = fmt.Errorf("marshaling patient: %w", err)
				return
			}

			_, err = a.Client.UpdateResource(ctx, "Patient", patientID, updated, nil)
			if err != nil {
				apiErr = fmt.Errorf("updating patient: %w", err)
				return
			}
		}).
		Run()

	if err != nil {
		ShowError(err)
		PressEnter()
		return
	}
	if apiErr != nil {
		ShowError(apiErr)
		PressEnter()
		return
	}

	fmt.Printf("\n  Updated patient %s\n", patientID)
	PressEnter()
}

// DeletePatient lets the user pick a patient and delete them after confirmation.
func (a *App) DeletePatient() {
	patientID, err := a.PickPatient()
	if err != nil {
		if !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}
	if patientID == "" {
		PressEnter()
		return
	}

	var confirm bool
	err = huh.NewConfirm().
		Title("Delete this patient?").
		Description("This action cannot be undone.").
		Value(&confirm).
		Run()
	if err != nil || !confirm {
		return
	}

	var apiErr error
	err = spinner.New().
		Title("Deleting patient...").
		Action(func() {
			apiErr = a.Client.DeleteResource(context.Background(), "Patient", patientID)
		}).
		Run()

	if err != nil {
		ShowError(err)
		PressEnter()
		return
	}
	if apiErr != nil {
		ShowError(fmt.Errorf("deleting patient: %w", apiErr))
		PressEnter()
		return
	}

	fmt.Printf("\n  Deleted patient %s\n", patientID)
	PressEnter()
}
