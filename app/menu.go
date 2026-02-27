package app

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// MainMenu runs the top-level interactive menu loop.
func (a *App) MainMenu() {
	for {
		fmt.Println()
		var choice string
		err := huh.NewSelect[string]().
			Title("Community Health Clinic").
			Options(
				huh.NewOption("Seed Sample Data", "seed"),
				huh.NewOption("Patient Summary", "summary"),
				huh.NewOption("Clinic Dashboard", "dashboard"),
				huh.NewOption("Manage Data", "manage"),
				huh.NewOption("Delete Seed Data", "unseed"),
				huh.NewOption("Exit", "exit"),
			).
			Value(&choice).
			Run()

		if err != nil {
			if isAbort(err) {
				fmt.Println("\nGoodbye!")
				return
			}
			ShowError(err)
			continue
		}

		switch choice {
		case "seed":
			a.SeedData()
		case "summary":
			a.PatientSummary()
		case "dashboard":
			a.ClinicDashboard()
		case "manage":
			a.manageMenu()
		case "unseed":
			a.DeleteSeedData()
		case "exit":
			fmt.Println("\nGoodbye!")
			return
		}
	}
}

func (a *App) manageMenu() {
	for {
		var choice string
		err := huh.NewSelect[string]().
			Title("Manage Data").
			Options(
				huh.NewOption("Patient Management", "patient"),
				huh.NewOption("Clinical Records", "clinical"),
				huh.NewOption("Health Plans", "health"),
				huh.NewOption("\u2190 Back", "back"),
			).
			Value(&choice).
			Run()

		if err != nil {
			if isAbort(err) {
				return
			}
			ShowError(err)
			continue
		}

		switch choice {
		case "patient":
			a.patientMenu()
		case "clinical":
			a.clinicalMenu()
		case "health":
			a.healthPlanMenu()
		case "back":
			return
		}
	}
}

func (a *App) patientMenu() {
	for {
		var choice string
		err := huh.NewSelect[string]().
			Title("Patient Management").
			Options(
				huh.NewOption("Register New Patient", "register"),
				huh.NewOption("List All Patients", "list"),
				huh.NewOption("View Patient Details", "view"),
				huh.NewOption("Update Contact Info", "update"),
				huh.NewOption("Delete Patient", "delete"),
				huh.NewOption("\u2190 Back", "back"),
			).
			Value(&choice).
			Run()

		if err != nil {
			if isAbort(err) {
				return
			}
			ShowError(err)
			continue
		}

		switch choice {
		case "register":
			a.RegisterPatient()
		case "list":
			a.ListPatients()
		case "view":
			a.ViewPatient()
		case "update":
			a.UpdateContact()
		case "delete":
			a.DeletePatient()
		case "back":
			return
		}
	}
}

func (a *App) clinicalMenu() {
	for {
		var choice string
		err := huh.NewSelect[string]().
			Title("Clinical Records").
			Options(
				huh.NewOption("Record Vital Signs", "vitals-add"),
				huh.NewOption("View Patient Vitals", "vitals-view"),
				huh.NewOption("Record Diagnosis", "diagnosis-add"),
				huh.NewOption("View Patient Diagnoses", "diagnosis-view"),
				huh.NewOption("\u2190 Back", "back"),
			).
			Value(&choice).
			Run()

		if err != nil {
			if isAbort(err) {
				return
			}
			ShowError(err)
			continue
		}

		switch choice {
		case "vitals-add":
			a.RecordVitals()
		case "vitals-view":
			a.ViewVitals()
		case "diagnosis-add":
			a.RecordDiagnosis()
		case "diagnosis-view":
			a.ViewDiagnoses()
		case "back":
			return
		}
	}
}

func (a *App) healthPlanMenu() {
	for {
		var choice string
		err := huh.NewSelect[string]().
			Title("Health Plans").
			Options(
				huh.NewOption("Create New Plan", "create"),
				huh.NewOption("Add Activity to Plan", "add"),
				huh.NewOption("Complete Activity", "complete"),
				huh.NewOption("View Plan Status", "status"),
				huh.NewOption("\u2190 Back", "back"),
			).
			Value(&choice).
			Run()

		if err != nil {
			if isAbort(err) {
				return
			}
			ShowError(err)
			continue
		}

		switch choice {
		case "create":
			a.CreatePlan()
		case "add":
			a.AddActivity()
		case "complete":
			a.CompleteActivity()
		case "status":
			a.ViewPlanStatus()
		case "back":
			return
		}
	}
}
