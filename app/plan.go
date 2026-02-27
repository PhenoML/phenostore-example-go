package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/phenoml/phenostore-example-go/fhir"
	"github.com/phenoml/phenostore-sdk-go/phenostore/gen"
)

// CreatePlan lets the user pick a patient and create a new care plan.
func (a *App) CreatePlan() {
	patientID, err := a.PickPatient()
	if err != nil || patientID == "" {
		if err != nil && !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	var title string
	if err := huh.NewInput().Title("Plan title").Value(&title).Run(); err != nil {
		if !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	body := fhir.NewCarePlan(patientID, title)

	var created json.RawMessage
	var apiErr error

	err = spinner.New().
		Title("Creating care plan...").
		Action(func() {
			created, apiErr = a.Client.CreateResource(context.Background(), "CarePlan", body, nil)
		}).
		Run()

	if err != nil {
		ShowError(err)
		PressEnter()
		return
	}
	if apiErr != nil {
		ShowError(fmt.Errorf("creating care plan: %w", apiErr))
		PressEnter()
		return
	}

	id := fhir.ResourceID(created)
	fmt.Printf("\n  Created health plan %q (ID: %s)\n", title, id)
	PressEnter()
}

// AddActivity lets the user pick a patient, pick a plan, and add an activity.
func (a *App) AddActivity() {
	patientID, err := a.PickPatient()
	if err != nil || patientID == "" {
		if err != nil && !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	cpID, err := a.PickCarePlan(patientID)
	if err != nil || cpID == "" {
		if err != nil && !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	var description, due string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Activity description").Value(&description),
			huh.NewInput().Title("Due date (optional, YYYY-MM-DD)").Value(&due),
		),
	)

	if err := form.Run(); err != nil {
		if !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	var apiErr error

	err = spinner.New().
		Title("Adding activity...").
		Action(func() {
			ctx := context.Background()

			raw, err := a.Client.ReadResource(ctx, "CarePlan", cpID)
			if err != nil {
				apiErr = fmt.Errorf("reading care plan: %w", err)
				return
			}

			var carePlan map[string]any
			if err := json.Unmarshal(raw, &carePlan); err != nil {
				apiErr = fmt.Errorf("parsing care plan: %w", err)
				return
			}

			activities, _ := carePlan["activity"].([]any)
			activities = append(activities, fhir.NewCarePlanActivity(description, due))
			carePlan["activity"] = activities

			updated, err := json.Marshal(carePlan)
			if err != nil {
				apiErr = fmt.Errorf("marshaling care plan: %w", err)
				return
			}

			_, err = a.Client.UpdateResource(ctx, "CarePlan", cpID, updated, nil)
			if err != nil {
				apiErr = fmt.Errorf("updating care plan: %w", err)
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

	fmt.Printf("\n  Added activity: %s\n", description)
	PressEnter()
}

// CompleteActivity lets the user pick a patient, plan, and activity to mark as completed.
func (a *App) CompleteActivity() {
	patientID, err := a.PickPatient()
	if err != nil || patientID == "" {
		if err != nil && !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	cpID, err := a.PickCarePlan(patientID)
	if err != nil || cpID == "" {
		if err != nil && !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	// Read the care plan to show activities
	ctx := context.Background()
	var carePlanRaw json.RawMessage
	var apiErr error

	err = spinner.New().
		Title("Loading care plan...").
		Action(func() {
			carePlanRaw, apiErr = a.Client.ReadResource(ctx, "CarePlan", cpID)
		}).
		Run()

	if err != nil {
		ShowError(err)
		PressEnter()
		return
	}
	if apiErr != nil {
		ShowError(fmt.Errorf("reading care plan: %w", apiErr))
		PressEnter()
		return
	}

	var carePlan map[string]any
	if err := json.Unmarshal(carePlanRaw, &carePlan); err != nil {
		ShowError(fmt.Errorf("parsing care plan: %w", err))
		PressEnter()
		return
	}

	activities, _ := carePlan["activity"].([]any)
	if len(activities) == 0 {
		fmt.Println("\n  No activities in this care plan.")
		PressEnter()
		return
	}

	// Build options for incomplete activities
	var options []huh.Option[int]
	for i, a := range activities {
		act, ok := a.(map[string]any)
		if !ok {
			continue
		}
		detail, _ := act["detail"].(map[string]any)
		if detail == nil {
			continue
		}
		status, _ := detail["status"].(string)
		if status == "completed" {
			continue
		}
		desc, _ := detail["description"].(string)
		label := fmt.Sprintf("%d. %s", i+1, desc)
		options = append(options, huh.NewOption(label, i))
	}

	if len(options) == 0 {
		fmt.Println("\n  All activities are already completed.")
		PressEnter()
		return
	}

	var actIdx int
	err = huh.NewSelect[int]().
		Title("Select activity to complete").
		Options(options...).
		Value(&actIdx).
		Run()

	if err != nil {
		if !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	// Mark the activity as completed
	act, _ := activities[actIdx].(map[string]any)
	detail, _ := act["detail"].(map[string]any)
	detail["status"] = "completed"

	// Check if all activities are now completed
	allDone := true
	for _, a := range activities {
		am, _ := a.(map[string]any)
		d, _ := am["detail"].(map[string]any)
		if s, _ := d["status"].(string); s != "completed" {
			allDone = false
			break
		}
	}
	if allDone {
		carePlan["status"] = "completed"
	}

	updated, _ := json.Marshal(carePlan)

	err = spinner.New().
		Title("Updating care plan...").
		Action(func() {
			_, apiErr = a.Client.UpdateResource(ctx, "CarePlan", cpID, updated, nil)
		}).
		Run()

	if err != nil {
		ShowError(err)
		PressEnter()
		return
	}
	if apiErr != nil {
		ShowError(fmt.Errorf("updating care plan: %w", apiErr))
		PressEnter()
		return
	}

	desc, _ := detail["description"].(string)
	fmt.Printf("\n  Completed activity: %s\n", desc)
	if allDone {
		fmt.Println("  All activities completed \u2014 plan marked as completed.")
	}
	PressEnter()
}

// ViewPlanStatus lets the user pick a patient and view their care plans.
func (a *App) ViewPlanStatus() {
	patientID, err := a.PickPatient()
	if err != nil || patientID == "" {
		if err != nil && !isAbort(err) {
			ShowError(err)
			PressEnter()
		}
		return
	}

	var plans []json.RawMessage
	var fetchErr error
	var elapsed time.Duration

	err = spinner.New().
		Title("Loading care plans...").
		Action(func() {
			start := time.Now()
			plans, fetchErr = a.searchCarePlans(context.Background(), patientID)
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
	if len(plans) == 0 {
		fmt.Println("  No active health plans found.")
	} else {
		fhir.PrintCarePlanList(plans)
		showTiming(fmt.Sprintf("Fetched %d care plans", len(plans)), elapsed)
	}
	PressEnter()
}

// ClinicDashboard shows all active plans with progress across all patients.
func (a *App) ClinicDashboard() {
	ctx := context.Background()
	var entries []json.RawMessage
	var fetchErr error
	var elapsed time.Duration

	err := spinner.New().
		Title("Loading clinic dashboard...").
		Action(func() {
			start := time.Now()
			count := gen.SearchCount(100)
			params := &gen.SearchResourcesParams{
				UnderscoreCount: &count,
			}
			resp, err := a.Client.Inner().SearchResourcesWithResponse(
				ctx, a.Client.Tenant(), a.Client.Store(),
				gen.SearchResourcesParamsResourceType("CarePlan"), params,
				func(ctx context.Context, req *http.Request) error {
					q := req.URL.Query()
					q.Set("status", "active")
					req.URL.RawQuery = q.Encode()
					return nil
				},
			)
			if err != nil {
				fetchErr = fmt.Errorf("searching care plans: %w", err)
				return
			}
			if resp.HTTPResponse.StatusCode >= 400 {
				fetchErr = fmt.Errorf("search failed: HTTP %d", resp.HTTPResponse.StatusCode)
				return
			}
			var bundle gen.Bundle
			if err := json.Unmarshal(resp.Body, &bundle); err != nil {
				fetchErr = fmt.Errorf("parsing response: %w", err)
				return
			}
			entries = extractResources(bundle)
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

	if len(entries) == 0 {
		fmt.Println("\n  No active health plans found.")
		PressEnter()
		return
	}

	// Resolve patient names and collect dashboard plans
	patientNames := make(map[string]string)
	var allPlans []fhir.DashboardPlan

	for _, raw := range entries {
		m, err := fhir.Parse(raw)
		if err != nil {
			continue
		}
		patientID := fhir.PatientRef(m)
		name, ok := patientNames[patientID]
		if !ok {
			name = a.resolvePatientName(ctx, patientID)
			patientNames[patientID] = name
		}
		dp := fhir.GetDashboardPlan(m, name)
		allPlans = append(allPlans, dp)
	}

	fmt.Println()
	fhir.PrintClinicDashboard(allPlans)
	showTiming(fmt.Sprintf("Fetched %d active care plans across %d patients", len(entries), len(patientNames)), elapsed)
	PressEnter()
}
