package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/phenoml/phenostore-example-go/fhir"
)

// RecordVitals guides the user through recording an observation.
func (a *App) RecordVitals() {
	patientID, err := a.PickPatient()
	if err != nil || patientID == "" {
		if err != nil && !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	var obsType string
	err = huh.NewSelect[string]().
		Title("Vital sign type").
		Options(
			huh.NewOption("Blood Pressure", "bp"),
			huh.NewOption("Weight", "weight"),
			huh.NewOption("Heart Rate", "heart-rate"),
		).
		Value(&obsType).
		Run()

	if err != nil {
		if !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	var body json.RawMessage

	switch obsType {
	case "bp":
		var systolicStr, diastolicStr string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Systolic (mmHg)").Value(&systolicStr),
				huh.NewInput().Title("Diastolic (mmHg)").Value(&diastolicStr),
			),
		)
		if err := form.Run(); err != nil {
			if !isAbort(err) {
				ShowError(err)
				PressEnter()
			}
			return
		}
		systolic, err1 := strconv.Atoi(systolicStr)
		diastolic, err2 := strconv.Atoi(diastolicStr)
		if err1 != nil || err2 != nil {
			ShowError(fmt.Errorf("systolic and diastolic must be numbers"))
			PressEnter()
			return
		}
		body = fhir.NewBloodPressureObservation(patientID, systolic, diastolic)

	case "weight":
		var valueStr string
		if err := huh.NewInput().Title("Weight (kg)").Value(&valueStr).Run(); err != nil {
			if !isAbort(err) {
				ShowError(err)
				PressEnter()
			}
			return
		}
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			ShowError(fmt.Errorf("weight must be a number"))
			PressEnter()
			return
		}
		body = fhir.NewWeightObservation(patientID, value)

	case "heart-rate":
		var valueStr string
		if err := huh.NewInput().Title("Heart rate (bpm)").Value(&valueStr).Run(); err != nil {
			if !isAbort(err) {
				ShowError(err)
				PressEnter()
			}
			return
		}
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			ShowError(fmt.Errorf("heart rate must be a number"))
			PressEnter()
			return
		}
		body = fhir.NewHeartRateObservation(patientID, value)
	}

	var created json.RawMessage
	var apiErr error

	err = spinner.New().
		Title("Recording observation...").
		Action(func() {
			created, apiErr = a.Client.CreateResource(context.Background(), "Observation", body, nil)
		}).
		Run()

	if err != nil {
		ShowError(err)
		PressEnter()
		return
	}
	if apiErr != nil {
		ShowError(fmt.Errorf("creating observation: %w", apiErr))
		PressEnter()
		return
	}

	id := fhir.ResourceID(created)
	fmt.Printf("\n  Recorded %s observation (ID: %s)\n", obsType, id)
	PressEnter()
}

// ViewVitals lets the user pick a patient and view their observations.
func (a *App) ViewVitals() {
	patientID, err := a.PickPatient()
	if err != nil || patientID == "" {
		if err != nil && !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	var observations []json.RawMessage
	var fetchErr error
	var elapsed time.Duration

	err = spinner.New().
		Title("Loading observations...").
		Action(func() {
			start := time.Now()
			observations, fetchErr = a.searchByPatient(context.Background(), "Observation", patientID)
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
	if len(observations) == 0 {
		fmt.Println("  No observations found.")
	} else {
		fhir.PrintObservationList(observations)
		showTiming(fmt.Sprintf("Fetched %d observations", len(observations)), elapsed)
	}
	PressEnter()
}
