package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/phenoml/phenostore-example-go/fhir"
)

const seedTagQuery = "phenostore-example|seed"

var seedMeta = map[string]any{
	"tag": []map[string]any{
		{"system": "phenostore-example", "code": "seed"},
	},
}

// addSeedTag injects a meta.tag into a FHIR resource so it can be found later
// for cleanup. This avoids modifying the shared fhir.New* builders.
func addSeedTag(resource json.RawMessage) json.RawMessage {
	var m map[string]any
	_ = json.Unmarshal(resource, &m)
	m["meta"] = seedMeta
	b, _ := json.Marshal(m)
	return b
}

// obs is a shorthand for adding a tagged observation bundle entry.
func obs(entry map[string]any) map[string]any {
	raw := entry["resource"].(json.RawMessage)
	entry["resource"] = json.RawMessage(addSeedTag(raw))
	return entry
}

// SeedData loads sample patients with observations, conditions, and care plans.
func (a *App) SeedData() {
	var confirm bool
	err := huh.NewConfirm().
		Title("Seed sample data?").
		Description("Creates 5 patients with vitals, lab results, conditions, and care plans.").
		Value(&confirm).
		Run()
	if err != nil || !confirm {
		return
	}

	var entries []map[string]any
	p := func(urn string) string { return urn } // alias for readability

	// --- Patient 1: Maria Garcia ---
	// 39-year-old woman managing hypertension and anxiety. Elevated BP, on a
	// low-sodium diet plan. Recently started therapy for anxiety.
	p1 := p("urn:uuid:patient-1")
	entries = append(entries, bundleEntryWithUrn(p1, "Patient",
		addSeedTag(seedPatient("Maria", "Garcia", "1985-03-22", "female", "555-0101", "maria.garcia@email.com",
			&seedAddress{line: "Rua das Flores 142", city: "Rio de Janeiro", state: "RJ", postalCode: "20040-020"}))))
	// Vitals
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBloodPressureObservation(p1, 142, 91))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBloodPressureObservation(p1, 138, 88))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewWeightObservation(p1, 68.2))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewHeartRateObservation(p1, 78))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewTemperatureObservation(p1, 36.6))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewOxygenSaturationObservation(p1, 97))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewRespiratoryRateObservation(p1, 16))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBMIObservation(p1, 24.8))))
	// Labs
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewTotalCholesterolObservation(p1, 218))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBloodGlucoseObservation(p1, 92))))
	// Conditions
	entries = append(entries, fhir.BundleEntry("Condition", addSeedTag(fhir.NewCondition(p1, "I10", "Essential Hypertension"))))
	entries = append(entries, fhir.BundleEntry("Condition", addSeedTag(fhir.NewCondition(p1, "F41.1", "Generalized Anxiety Disorder"))))
	// Care plans
	entries = append(entries, bundleEntryWithUrn("urn:uuid:cp-1a", "CarePlan",
		addSeedTag(carePlanWithActivities(p1, "Hypertension Management", []seedActivity{
			{description: "Initial blood pressure screening", status: "completed"},
			{description: "Start low-sodium diet program", status: "in-progress", schedule: "By 2025-04-15"},
			{description: "Follow-up BP check in 30 days", status: "not-started", schedule: "By 2025-05-01"},
			{description: "Evaluate need for medication adjustment", status: "not-started", schedule: "By 2025-06-01"},
		}))))
	entries = append(entries, bundleEntryWithUrn("urn:uuid:cp-1b", "CarePlan",
		addSeedTag(carePlanWithActivities(p1, "Mental Health Support", []seedActivity{
			{description: "PHQ-9 screening questionnaire", status: "completed"},
			{description: "Cognitive behavioral therapy referral", status: "completed"},
			{description: "4-week therapy check-in", status: "not-started", schedule: "By 2025-05-15"},
		}))))

	// --- Patient 2: Wei Chen ---
	// 32-year-old man, generally healthy. Came in for a wellness visit. Mild
	// seasonal allergies, otherwise unremarkable. Good baseline vitals.
	p2 := p("urn:uuid:patient-2")
	entries = append(entries, bundleEntryWithUrn(p2, "Patient",
		addSeedTag(seedPatient("Wei", "Chen", "1992-07-14", "male", "555-0202", "",
			&seedAddress{line: "Av. Atlântica 1702", city: "Rio de Janeiro", state: "RJ", postalCode: "22021-001"}))))
	// Vitals
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBloodPressureObservation(p2, 118, 76))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewWeightObservation(p2, 79.5))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewHeartRateObservation(p2, 68))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewTemperatureObservation(p2, 36.5))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewOxygenSaturationObservation(p2, 99))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewRespiratoryRateObservation(p2, 14))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBMIObservation(p2, 24.1))))
	// Labs
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewTotalCholesterolObservation(p2, 185))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBloodGlucoseObservation(p2, 88))))
	// Conditions
	entries = append(entries, fhir.BundleEntry("Condition", addSeedTag(fhir.NewCondition(p2, "J30.2", "Seasonal Allergic Rhinitis"))))
	// Care plans
	entries = append(entries, bundleEntryWithUrn("urn:uuid:cp-2", "CarePlan",
		addSeedTag(carePlanWithActivities(p2, "Annual Wellness", []seedActivity{
			{description: "Comprehensive metabolic panel", status: "completed"},
			{description: "Lipid panel blood draw", status: "completed"},
			{description: "Flu vaccination", status: "not-started", schedule: "By 2025-10-01"},
			{description: "Schedule next annual physical", status: "not-started", schedule: "By 2026-03-01"},
		}))))

	// --- Patient 3: Alex Thompson ---
	// 47-year-old non-binary patient with multiple comorbidities — diabetes,
	// hypertension, and obesity. Complex care needs with two active plans.
	p3 := p("urn:uuid:patient-3")
	entries = append(entries, bundleEntryWithUrn(p3, "Patient",
		addSeedTag(seedPatient("Alex", "Thompson", "1978-11-03", "other", "555-0303", "alex.t@email.com",
			&seedAddress{line: "Rua Visconde de Pirajá 330", city: "Rio de Janeiro", state: "RJ", postalCode: "22410-002"}))))
	// Vitals
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBloodPressureObservation(p3, 148, 94))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBloodPressureObservation(p3, 145, 92))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewWeightObservation(p3, 104.3))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewWeightObservation(p3, 101.8))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewHeartRateObservation(p3, 88))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewTemperatureObservation(p3, 36.8))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewOxygenSaturationObservation(p3, 96))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewRespiratoryRateObservation(p3, 18))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBMIObservation(p3, 36.2))))
	// Labs
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewHbA1cObservation(p3, 7.8))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBloodGlucoseObservation(p3, 156))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewTotalCholesterolObservation(p3, 242))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewCreatinineObservation(p3, 1.1))))
	// Conditions
	entries = append(entries, fhir.BundleEntry("Condition", addSeedTag(fhir.NewCondition(p3, "E11.9", "Type 2 Diabetes Mellitus"))))
	entries = append(entries, fhir.BundleEntry("Condition", addSeedTag(fhir.NewCondition(p3, "I10", "Essential Hypertension"))))
	entries = append(entries, fhir.BundleEntry("Condition", addSeedTag(fhir.NewCondition(p3, "E66.01", "Morbid Obesity due to Excess Calories"))))
	// Care plans
	entries = append(entries, bundleEntryWithUrn("urn:uuid:cp-3a", "CarePlan",
		addSeedTag(carePlanWithActivities(p3, "Diabetes Care Plan", []seedActivity{
			{description: "HbA1c lab test", status: "completed"},
			{description: "Start metformin 500mg twice daily", status: "completed"},
			{description: "Diabetic retinal exam", status: "not-started", schedule: "By 2025-06-01"},
			{description: "Complete diabetes self-management education", status: "not-started", schedule: "By 2025-05-15"},
			{description: "Repeat HbA1c in 3 months", status: "not-started", schedule: "By 2025-07-01"},
		}))))
	entries = append(entries, bundleEntryWithUrn("urn:uuid:cp-3b", "CarePlan",
		addSeedTag(carePlanWithActivities(p3, "Weight Management", []seedActivity{
			{description: "Nutrition counseling intake session", status: "completed"},
			{description: "Begin supervised exercise program (3x/week)", status: "in-progress"},
			{description: "Monthly weigh-in and progress review", status: "not-started", schedule: "By 2025-05-01"},
			{description: "Evaluate for bariatric surgery referral if <5% loss in 6 months", status: "not-started", schedule: "By 2025-10-01"},
		}))))

	// --- Patient 4: Sarah Johnson ---
	// 23-year-old college athlete getting sports clearance. Excellent vitals.
	// Mild exercise-induced asthma, well-controlled. Mostly done with her plan.
	p4 := p("urn:uuid:patient-4")
	entries = append(entries, bundleEntryWithUrn(p4, "Patient",
		addSeedTag(seedPatient("Sarah", "Johnson", "2001-05-28", "female", "", "sarah.j@university.edu",
			&seedAddress{line: "Rua Jardim Botânico 920", city: "Rio de Janeiro", state: "RJ", postalCode: "22460-030"}))))
	// Vitals
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBloodPressureObservation(p4, 108, 68))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewWeightObservation(p4, 61.2))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewHeartRateObservation(p4, 52))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewTemperatureObservation(p4, 36.4))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewOxygenSaturationObservation(p4, 99))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewRespiratoryRateObservation(p4, 12))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBMIObservation(p4, 21.3))))
	// Conditions
	entries = append(entries, fhir.BundleEntry("Condition", addSeedTag(fhir.NewCondition(p4, "J45.990", "Exercise-Induced Bronchospasm"))))
	// Care plans
	entries = append(entries, bundleEntryWithUrn("urn:uuid:cp-4", "CarePlan",
		addSeedTag(carePlanWithActivities(p4, "Sports Clearance", []seedActivity{
			{description: "Pre-participation physical exam", status: "completed"},
			{description: "ECG screening", status: "completed"},
			{description: "Pulmonary function test", status: "completed"},
			{description: "Rescue inhaler prescription renewal", status: "not-started", schedule: "By 2025-08-01"},
		}))))

	// --- Patient 5: James Williams ---
	// 60-year-old man with chronic kidney disease, hypertension, and high
	// cholesterol. Multiple specialists involved. Highest-acuity patient.
	p5 := p("urn:uuid:patient-5")
	entries = append(entries, bundleEntryWithUrn(p5, "Patient",
		addSeedTag(seedPatient("James", "Williams", "1965-09-10", "male", "555-0505", "jwilliams@email.com",
			&seedAddress{line: "Av. Niemeyer 776", city: "Rio de Janeiro", state: "RJ", postalCode: "22450-221"}))))
	// Vitals
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBloodPressureObservation(p5, 162, 99))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBloodPressureObservation(p5, 155, 96))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewWeightObservation(p5, 88.4))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewHeartRateObservation(p5, 82))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewTemperatureObservation(p5, 36.7))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewOxygenSaturationObservation(p5, 95))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewRespiratoryRateObservation(p5, 18))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBMIObservation(p5, 28.6))))
	// Labs
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewCreatinineObservation(p5, 1.8))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewEGFRObservation(p5, 42))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewTotalCholesterolObservation(p5, 261))))
	entries = append(entries, obs(fhir.BundleEntry("Observation", fhir.NewBloodGlucoseObservation(p5, 108))))
	// Conditions
	entries = append(entries, fhir.BundleEntry("Condition", addSeedTag(fhir.NewCondition(p5, "I10", "Essential Hypertension"))))
	entries = append(entries, fhir.BundleEntry("Condition", addSeedTag(fhir.NewCondition(p5, "N18.3", "Chronic Kidney Disease, Stage 3"))))
	entries = append(entries, fhir.BundleEntry("Condition", addSeedTag(fhir.NewCondition(p5, "E78.5", "Hyperlipidemia, Unspecified"))))
	// Care plans
	entries = append(entries, bundleEntryWithUrn("urn:uuid:cp-5a", "CarePlan",
		addSeedTag(carePlanWithActivities(p5, "CKD Monitoring", []seedActivity{
			{description: "Baseline kidney function labs (GFR, creatinine)", status: "completed"},
			{description: "Nephrology referral", status: "in-progress", schedule: "By 2025-04-15"},
			{description: "Start renal-protective diet (low protein, low sodium)", status: "not-started", schedule: "By 2025-05-01"},
			{description: "Repeat GFR in 3 months", status: "not-started", schedule: "By 2025-07-01"},
		}))))
	entries = append(entries, bundleEntryWithUrn("urn:uuid:cp-5b", "CarePlan",
		addSeedTag(carePlanWithActivities(p5, "Cardiovascular Risk Reduction", []seedActivity{
			{description: "Fasting lipid panel", status: "completed"},
			{description: "Start atorvastatin 20mg daily", status: "completed"},
			{description: "Recheck lipids in 6 weeks", status: "not-started", schedule: "By 2025-05-15"},
			{description: "Cardiology consult for stress test", status: "not-started", schedule: "By 2025-06-01"},
		}))))

	bundle := fhir.TransactionBundle(entries)

	var created int
	var apiErr error
	var elapsed time.Duration

	err = spinner.New().
		Title("Seeding sample data...").
		Action(func() {
			start := time.Now()
			result, err := a.Client.ProcessBundle(context.Background(), bundle)
			elapsed = time.Since(start)
			if err != nil {
				apiErr = err
				return
			}
			if result.Entry != nil {
				for _, entry := range *result.Entry {
					if entry.Response != nil && entry.Response.Status != nil && strings.HasPrefix(*entry.Response.Status, "201") {
						created++
					}
				}
			}
		}).
		Run()

	if err != nil {
		ShowError(err)
		PressEnter()
		return
	}
	if apiErr != nil {
		ShowError(fmt.Errorf("processing bundle: %w", apiErr))
		PressEnter()
		return
	}

	fmt.Printf("\n  Seeded %d resources (5 patients with vitals, labs, conditions, and care plans)\n", created)
	showTiming(fmt.Sprintf("Created %d resources via transaction bundle", created), elapsed)
	PressEnter()
}

type seedAddress struct {
	line       string
	city       string
	state      string
	postalCode string
}

// seedPatient builds a Patient resource with optional contact info and address.
func seedPatient(given, family, dob, gender, phone, email string, addr *seedAddress) json.RawMessage {
	p := map[string]any{
		"resourceType": "Patient",
		"name": []map[string]any{
			{
				"given":  []string{given},
				"family": family,
			},
		},
		"birthDate": dob,
		"gender":    gender,
	}
	var telecoms []map[string]any
	if phone != "" {
		telecoms = append(telecoms, map[string]any{"system": "phone", "value": phone})
	}
	if email != "" {
		telecoms = append(telecoms, map[string]any{"system": "email", "value": email})
	}
	if len(telecoms) > 0 {
		p["telecom"] = telecoms
	}
	if addr != nil {
		a := map[string]any{
			"city":       addr.city,
			"state":      addr.state,
			"postalCode": addr.postalCode,
		}
		if addr.line != "" {
			a["line"] = []string{addr.line}
		}
		p["address"] = []map[string]any{a}
	}
	b, _ := json.Marshal(p)
	return b
}

type seedActivity struct {
	description string
	status      string
	schedule    string
}

func carePlanWithActivities(patientID, title string, activities []seedActivity) json.RawMessage {
	acts := make([]any, len(activities))
	for i, a := range activities {
		detail := map[string]any{
			"status":      a.status,
			"description": a.description,
		}
		if a.schedule != "" {
			detail["scheduledString"] = a.schedule
		}
		acts[i] = map[string]any{
			"detail": detail,
		}
	}
	cp := map[string]any{
		"resourceType": "CarePlan",
		"status":       "active",
		"intent":       "plan",
		"title":        title,
		"subject": map[string]any{
			"reference": patientID,
		},
		"activity": acts,
	}
	b, _ := json.Marshal(cp)
	return b
}

func bundleEntryWithUrn(urn, resourceType string, resource json.RawMessage) map[string]any {
	return map[string]any{
		"fullUrl":  urn,
		"resource": json.RawMessage(resource),
		"request": map[string]any{
			"method": "POST",
			"url":    resourceType,
		},
	}
}

// DeleteSeedData removes all resources that were created by SeedData.
// It searches by the meta.tag added during seeding, so user-created
// resources are never touched.
func (a *App) DeleteSeedData() {
	var confirm bool
	err := huh.NewConfirm().
		Title("Delete all seed data?").
		Description("Only removes resources created by \"Seed Sample Data\". Your own data is safe.").
		Value(&confirm).
		Run()
	if err != nil || !confirm {
		return
	}

	ctx := context.Background()
	var deleted int
	var apiErr error
	var elapsed time.Duration

	// Delete dependents before patients to avoid referential issues.
	resourceTypes := []string{"CarePlan", "Observation", "Condition", "Patient"}

	err = spinner.New().
		Title("Deleting seed data...").
		Action(func() {
			start := time.Now()
			for _, rt := range resourceTypes {
				ids, err := a.searchByTag(ctx, rt, seedTagQuery)
				if err != nil {
					apiErr = err
					return
				}
				for _, id := range ids {
					if err := a.Client.DeleteResource(ctx, rt, id); err != nil {
						apiErr = fmt.Errorf("deleting %s/%s: %w", rt, id, err)
						return
					}
					deleted++
				}
			}
			elapsed = time.Since(start)
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

	if deleted == 0 {
		fmt.Println("\n  No seed data found.")
	} else {
		fmt.Printf("\n  Deleted %d seed resources.\n", deleted)
		showTiming(fmt.Sprintf("Deleted %d resources", deleted), elapsed)
	}
	PressEnter()
}
