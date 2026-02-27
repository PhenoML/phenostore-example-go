package fhir

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	labelStyle  = lipgloss.NewStyle().Width(14).Foreground(lipgloss.Color("8"))
	checkDone    = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("[x]")
	checkActive  = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("[~]")
	checkOpen    = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("[ ]")
)

// --- JSON access helpers ---

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getMap(m map[string]any, key string) map[string]any {
	if v, ok := m[key]; ok {
		if mm, ok := v.(map[string]any); ok {
			return mm
		}
	}
	return nil
}

func getSlice(m map[string]any, key string) []any {
	if v, ok := m[key]; ok {
		if s, ok := v.([]any); ok {
			return s
		}
	}
	return nil
}

func getNumber(m map[string]any, key string) float64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return n
		case json.Number:
			f, _ := n.Float64()
			return f
		}
	}
	return 0
}

// Parse unmarshals raw JSON into a map.
func Parse(raw json.RawMessage) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// ResourceID extracts the "id" field from a FHIR resource.
func ResourceID(raw json.RawMessage) string {
	m, err := Parse(raw)
	if err != nil {
		return ""
	}
	return getString(m, "id")
}

// PatientName extracts a display name from a FHIR Patient resource.
func PatientName(m map[string]any) string {
	names := getSlice(m, "name")
	if len(names) == 0 {
		return "(unknown)"
	}
	name, ok := names[0].(map[string]any)
	if !ok {
		return "(unknown)"
	}
	given := ""
	if gs := getSlice(name, "given"); len(gs) > 0 {
		if s, ok := gs[0].(string); ok {
			given = s
		}
	}
	family := getString(name, "family")
	return strings.TrimSpace(given + " " + family)
}

// PatientRef extracts the patient ID from a subject reference like "Patient/abc123".
func PatientRef(m map[string]any) string {
	sub := getMap(m, "subject")
	if sub == nil {
		return ""
	}
	ref := getString(sub, "reference")
	if strings.HasPrefix(ref, "Patient/") {
		return ref[len("Patient/"):]
	}
	return ref
}

// PrintPatient displays a Patient resource.
func PrintPatient(raw json.RawMessage) {
	m, err := Parse(raw)
	if err != nil {
		fmt.Println("Error parsing patient:", err)
		return
	}
	id := getString(m, "id")
	name := PatientName(m)

	fmt.Println(headerStyle.Render(fmt.Sprintf("Patient: %s (%s)", name, id)))
	fmt.Printf("  %s%s\n", labelStyle.Render("Gender:"), getString(m, "gender"))
	fmt.Printf("  %s%s\n", labelStyle.Render("Born:"), getString(m, "birthDate"))

	if telecoms := getSlice(m, "telecom"); len(telecoms) > 0 {
		for _, t := range telecoms {
			if tm, ok := t.(map[string]any); ok {
				system := getString(tm, "system")
				value := getString(tm, "value")
				label := system
				if len(label) > 0 {
					label = strings.ToUpper(label[:1]) + label[1:]
				}
				fmt.Printf("  %s%s\n", labelStyle.Render(label+":"), value)
			}
		}
	}

	if addrs := getSlice(m, "address"); len(addrs) > 0 {
		if addr, ok := addrs[0].(map[string]any); ok {
			var parts []string
			if lines := getSlice(addr, "line"); len(lines) > 0 {
				if line, ok := lines[0].(string); ok {
					parts = append(parts, line)
				}
			}
			city := getString(addr, "city")
			state := getString(addr, "state")
			postal := getString(addr, "postalCode")
			if city != "" {
				cityPart := city
				if state != "" {
					cityPart += ", " + state
				}
				if postal != "" {
					cityPart += " " + postal
				}
				parts = append(parts, cityPart)
			}
			if len(parts) > 0 {
				fmt.Printf("  %s%s\n", labelStyle.Render("Address:"), strings.Join(parts, ", "))
			}
		}
	}
}

// PrintPatientList displays a list of patients in a compact format.
func PrintPatientList(entries []json.RawMessage) {
	fmt.Println(headerStyle.Render(fmt.Sprintf("Patients (%d)", len(entries))))
	for _, raw := range entries {
		m, err := Parse(raw)
		if err != nil {
			continue
		}
		id := getString(m, "id")
		name := PatientName(m)
		gender := getString(m, "gender")
		dob := getString(m, "birthDate")
		fmt.Printf("  %-36s  %-20s  %-8s  %s\n", id, name, gender, dob)
	}
}

// PrintObservation displays a single Observation.
func PrintObservation(m map[string]any) {
	code := getMap(m, "code")
	display := ""
	if code != nil {
		display = getString(code, "text")
	}

	// Check for components (blood pressure)
	if components := getSlice(m, "component"); len(components) >= 2 {
		c1, _ := components[0].(map[string]any)
		c2, _ := components[1].(map[string]any)
		v1 := getNumber(getMap(c1, "valueQuantity"), "value")
		v2 := getNumber(getMap(c2, "valueQuantity"), "value")
		fmt.Printf("  %-16s  %d/%d mmHg\n", display, int(v1), int(v2))
		return
	}

	// Simple value
	vq := getMap(m, "valueQuantity")
	if vq != nil {
		val := getNumber(vq, "value")
		unit := getString(vq, "unit")
		if val == float64(int(val)) {
			fmt.Printf("  %-16s  %d %s\n", display, int(val), unit)
		} else {
			fmt.Printf("  %-16s  %.1f %s\n", display, val, unit)
		}
	}
}

// PrintObservationList displays multiple observations.
func PrintObservationList(entries []json.RawMessage) {
	fmt.Println(headerStyle.Render(fmt.Sprintf("Observations (%d)", len(entries))))
	for _, raw := range entries {
		m, err := Parse(raw)
		if err != nil {
			continue
		}
		PrintObservation(m)
	}
}

// PrintCondition displays a single Condition.
func PrintCondition(m map[string]any) {
	code := getMap(m, "code")
	if code == nil {
		return
	}
	display := getString(code, "text")
	icd := ""
	if codings := getSlice(code, "coding"); len(codings) > 0 {
		if c, ok := codings[0].(map[string]any); ok {
			icd = getString(c, "code")
		}
	}
	if icd != "" {
		fmt.Printf("  %s (%s)\n", display, icd)
	} else {
		fmt.Printf("  %s\n", display)
	}
}

// PrintConditionList displays multiple conditions.
func PrintConditionList(entries []json.RawMessage) {
	fmt.Println(headerStyle.Render(fmt.Sprintf("Conditions (%d)", len(entries))))
	for _, raw := range entries {
		m, err := Parse(raw)
		if err != nil {
			continue
		}
		PrintCondition(m)
	}
}

// carePlanProgress counts completed and total activities in a CarePlan.
func carePlanProgress(m map[string]any) (completed, total int) {
	for _, a := range getSlice(m, "activity") {
		act, ok := a.(map[string]any)
		if !ok {
			continue
		}
		detail := getMap(act, "detail")
		if detail == nil {
			continue
		}
		total++
		if getString(detail, "status") == "completed" {
			completed++
		}
	}
	return
}

// PrintCarePlan displays a CarePlan with its activities.
func PrintCarePlan(m map[string]any) {
	title := getString(m, "title")
	status := getString(m, "status")
	id := getString(m, "id")

	done, total := carePlanProgress(m)
	pct := 0
	if total > 0 {
		pct = done * 100 / total
	}

	fmt.Println(headerStyle.Render(fmt.Sprintf("Health Plan: %s (%s) [%s]", title, status, id)))
	if total > 0 {
		progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		fmt.Println(progressStyle.Render(fmt.Sprintf("  Progress: %d/%d complete (%d%%)", done, total, pct)))
	}

	activities := getSlice(m, "activity")
	for i, a := range activities {
		act, ok := a.(map[string]any)
		if !ok {
			continue
		}
		detail := getMap(act, "detail")
		if detail == nil {
			continue
		}
		desc := getString(detail, "description")
		st := getString(detail, "status")
		check := checkOpen
		if st == "completed" {
			check = checkDone
		} else if st == "in-progress" {
			check = checkActive
		}
		sched := getString(detail, "scheduledString")
		line := fmt.Sprintf("  %d. %s %s", i+1, check, desc)
		if sched != "" {
			line += fmt.Sprintf("  (%s)", sched)
		}
		fmt.Println(line)
	}
}

// PrintCarePlanList displays multiple care plans.
func PrintCarePlanList(entries []json.RawMessage) {
	for _, raw := range entries {
		m, err := Parse(raw)
		if err != nil {
			continue
		}
		PrintCarePlan(m)
		fmt.Println()
	}
}

// DashboardPlan holds a parsed care plan with its patient name for the clinic dashboard.
type DashboardPlan struct {
	PatientName string
	Title       string
	Completed   int
	Total       int
	Outstanding []DashboardItem
}

// DashboardItem represents an incomplete activity.
type DashboardItem struct {
	Description  string
	Status       string
	ScheduleNote string
}

// GetDashboardPlan extracts dashboard info from a CarePlan.
func GetDashboardPlan(carePlan map[string]any, patientName string) DashboardPlan {
	dp := DashboardPlan{
		PatientName: patientName,
		Title:       getString(carePlan, "title"),
	}
	for _, a := range getSlice(carePlan, "activity") {
		act, ok := a.(map[string]any)
		if !ok {
			continue
		}
		detail := getMap(act, "detail")
		if detail == nil {
			continue
		}
		dp.Total++
		if getString(detail, "status") == "completed" {
			dp.Completed++
		} else {
			dp.Outstanding = append(dp.Outstanding, DashboardItem{
				Description:  getString(detail, "description"),
				Status:       getString(detail, "status"),
				ScheduleNote: getString(detail, "scheduledString"),
			})
		}
	}
	return dp
}

// PrintClinicDashboard displays active plans grouped by patient with progress.
func PrintClinicDashboard(plans []DashboardPlan) {
	if len(plans) == 0 {
		fmt.Println("No outstanding items.")
		return
	}

	fmt.Println(headerStyle.Render("Clinic Dashboard — Outstanding Items"))
	fmt.Println()

	progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	currentPatient := ""
	for _, plan := range plans {
		if plan.PatientName != currentPatient {
			if currentPatient != "" {
				fmt.Println()
			}
			currentPatient = plan.PatientName
			fmt.Println(lipgloss.NewStyle().Bold(true).Render(plan.PatientName))
		}
		pct := 0
		if plan.Total > 0 {
			pct = plan.Completed * 100 / plan.Total
		}
		fmt.Printf("  %s  %s\n", plan.Title,
			progressStyle.Render(fmt.Sprintf("(%d/%d complete, %d%%)", plan.Completed, plan.Total, pct)))
		for _, item := range plan.Outstanding {
			check := checkOpen
			if item.Status == "in-progress" {
				check = checkActive
			}
			line := fmt.Sprintf("    %s %s", check, item.Description)
			if item.ScheduleNote != "" {
				line += fmt.Sprintf("  (%s)", item.ScheduleNote)
			}
			fmt.Println(line)
		}
	}
}

// labLoincCodes are LOINC codes that represent lab results rather than vital signs.
var labLoincCodes = map[string]bool{
	"2345-7":  true, // Blood Glucose
	"2093-3":  true, // Total Cholesterol
	"4548-4":  true, // HbA1c
	"2160-0":  true, // Creatinine
	"33914-3": true, // eGFR
}

// observationLoincCode extracts the primary LOINC code from an Observation.
func observationLoincCode(m map[string]any) string {
	code := getMap(m, "code")
	if code == nil {
		return ""
	}
	codings := getSlice(code, "coding")
	if len(codings) == 0 {
		return ""
	}
	if c, ok := codings[0].(map[string]any); ok {
		return getString(c, "code")
	}
	return ""
}

// PrintSummary displays a full patient summary with observations, conditions, and plans.
func PrintSummary(patient json.RawMessage, observations, conditions, plans []json.RawMessage) {
	PrintPatient(patient)
	fmt.Println()

	// Split observations into vital signs and lab results.
	var vitals, labs []json.RawMessage
	for _, raw := range observations {
		m, err := Parse(raw)
		if err != nil {
			continue
		}
		loinc := observationLoincCode(m)
		if labLoincCodes[loinc] {
			labs = append(labs, raw)
		} else {
			vitals = append(vitals, raw)
		}
	}

	if len(vitals) > 0 {
		fmt.Println(headerStyle.Render(fmt.Sprintf("Vital Signs (%d)", len(vitals))))
		for _, raw := range vitals {
			m, _ := Parse(raw)
			PrintObservation(m)
		}
		fmt.Println()
	}
	if len(labs) > 0 {
		fmt.Println(headerStyle.Render(fmt.Sprintf("Lab Results (%d)", len(labs))))
		for _, raw := range labs {
			m, _ := Parse(raw)
			PrintObservation(m)
		}
		fmt.Println()
	}

	if len(conditions) > 0 {
		PrintConditionList(conditions)
		fmt.Println()
	}
	if len(plans) > 0 {
		for _, raw := range plans {
			m, err := Parse(raw)
			if err != nil {
				continue
			}
			PrintCarePlan(m)
			fmt.Println()
		}
	}
}
