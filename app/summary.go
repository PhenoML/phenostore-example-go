package app

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/phenoml/phenostore-example-go/fhir"
	"github.com/phenoml/phenostore-sdk-go/phenostore"
)

// PatientSummary lets the user pick a patient and displays a full summary.
func (a *App) PatientSummary() {
	patientID, err := a.PickPatient()
	if err != nil || patientID == "" {
		if err != nil && !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	ctx := context.Background()
	var patient json.RawMessage
	var observations, conditions, plans []json.RawMessage
	var apiErr error
	var elapsed time.Duration

	err = spinner.New().
		Title("Loading patient summary...").
		Action(func() {
			start := time.Now()

			var wg sync.WaitGroup
			var patientErr error
			var observationsErr error
			var conditionsErr error
			var plansErr error

			// Fire all 4 API calls in parallel.
			wg.Add(4)
			go func() {
				defer wg.Done()
				var err error
				patient, err = a.Client.ReadResource(ctx, "Patient", patientID)
				if err != nil {
					patientErr = err
				}
			}()
			go func() {
				defer wg.Done()
				observations, observationsErr = a.searchByPatient(ctx, "Observation", patientID)
			}()
			go func() {
				defer wg.Done()
				conditions, conditionsErr = a.searchByPatient(ctx, "Condition", patientID)
			}()
			go func() {
				defer wg.Done()
				plans, plansErr = a.searchByPatient(ctx, "CarePlan", patientID)
			}()
			wg.Wait()

			elapsed = time.Since(start)

			if phenostore.IsNotFound(patientErr) {
				apiErr = fmt.Errorf("patient %s not found", patientID)
				return
			}
			if patientErr != nil {
				apiErr = fmt.Errorf("reading patient: %w", patientErr)
				return
			}
			if observationsErr != nil {
				apiErr = fmt.Errorf("loading observations: %w", observationsErr)
				return
			}
			if conditionsErr != nil {
				apiErr = fmt.Errorf("loading conditions: %w", conditionsErr)
				return
			}
			if plansErr != nil {
				apiErr = fmt.Errorf("loading care plans: %w", plansErr)
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

	fmt.Println()
	fhir.PrintSummary(patient, observations, conditions, plans)
	total := len(observations) + len(conditions) + len(plans) + 1
	showTiming(fmt.Sprintf("Loaded patient summary (%d resources, 4 parallel API calls)", total), elapsed)
	PressEnter()
}
