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

// RecordDiagnosis guides the user through recording a condition.
func (a *App) RecordDiagnosis() {
	patientID, err := a.PickPatient()
	if err != nil || patientID == "" {
		if err != nil && !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	var code, display string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("ICD-10 code (e.g., I10)").Value(&code),
			huh.NewInput().Title("Display name (e.g., Hypertension)").Value(&display),
		),
	)

	if err := form.Run(); err != nil {
		if !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	body := fhir.NewCondition(patientID, code, display)

	var created json.RawMessage
	var apiErr error

	err = spinner.New().
		Title("Recording diagnosis...").
		Action(func() {
			created, apiErr = a.Client.CreateResource(context.Background(), "Condition", body, nil)
		}).
		Run()

	if err != nil {
		ShowError(err)
		PressEnter()
		return
	}
	if apiErr != nil {
		ShowError(fmt.Errorf("creating condition: %w", apiErr))
		PressEnter()
		return
	}

	id := fhir.ResourceID(created)
	fmt.Printf("\n  Recorded condition %s \u2014 %s (ID: %s)\n", code, display, id)
	PressEnter()
}

// ViewDiagnoses lets the user pick a patient and view their conditions.
func (a *App) ViewDiagnoses() {
	patientID, err := a.PickPatient()
	if err != nil || patientID == "" {
		if err != nil && !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	var conditions []json.RawMessage
	var fetchErr error
	var elapsed time.Duration

	err = spinner.New().
		Title("Loading diagnoses...").
		Action(func() {
			start := time.Now()
			conditions, fetchErr = a.searchByPatient(context.Background(), "Condition", patientID)
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
	if len(conditions) == 0 {
		fmt.Println("  No conditions found.")
	} else {
		fhir.PrintConditionList(conditions)
		showTiming(fmt.Sprintf("Fetched %d conditions", len(conditions)), elapsed)
	}
	PressEnter()
}
