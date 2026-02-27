package fhir

import "encoding/json"

// NewPatient builds a FHIR Patient resource as JSON.
func NewPatient(given, family, dob, gender string) json.RawMessage {
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
	b, _ := json.Marshal(p)
	return b
}

// NewBloodPressureObservation builds a FHIR Observation for blood pressure.
func NewBloodPressureObservation(patientID string, systolic, diastolic int) json.RawMessage {
	obs := map[string]any{
		"resourceType": "Observation",
		"status":       "final",
		"code": map[string]any{
			"coding": []map[string]any{
				{
					"system":  "http://loinc.org",
					"code":    "85354-9",
					"display": "Blood pressure panel",
				},
			},
			"text": "Blood Pressure",
		},
		"subject": map[string]any{
			"reference": "Patient/" + patientID,
		},
		"component": []map[string]any{
			{
				"code": map[string]any{
					"coding": []map[string]any{
						{
							"system":  "http://loinc.org",
							"code":    "8480-6",
							"display": "Systolic blood pressure",
						},
					},
				},
				"valueQuantity": map[string]any{
					"value":  systolic,
					"unit":   "mmHg",
					"system": "http://unitsofmeasure.org",
					"code":   "mm[Hg]",
				},
			},
			{
				"code": map[string]any{
					"coding": []map[string]any{
						{
							"system":  "http://loinc.org",
							"code":    "8462-4",
							"display": "Diastolic blood pressure",
						},
					},
				},
				"valueQuantity": map[string]any{
					"value":  diastolic,
					"unit":   "mmHg",
					"system": "http://unitsofmeasure.org",
					"code":   "mm[Hg]",
				},
			},
		},
	}
	b, _ := json.Marshal(obs)
	return b
}

// NewWeightObservation builds a FHIR Observation for body weight.
func NewWeightObservation(patientID string, kg float64) json.RawMessage {
	obs := map[string]any{
		"resourceType": "Observation",
		"status":       "final",
		"code": map[string]any{
			"coding": []map[string]any{
				{
					"system":  "http://loinc.org",
					"code":    "29463-7",
					"display": "Body weight",
				},
			},
			"text": "Weight",
		},
		"subject": map[string]any{
			"reference": "Patient/" + patientID,
		},
		"valueQuantity": map[string]any{
			"value":  kg,
			"unit":   "kg",
			"system": "http://unitsofmeasure.org",
			"code":   "kg",
		},
	}
	b, _ := json.Marshal(obs)
	return b
}

// NewHeartRateObservation builds a FHIR Observation for heart rate.
func NewHeartRateObservation(patientID string, bpm int) json.RawMessage {
	obs := map[string]any{
		"resourceType": "Observation",
		"status":       "final",
		"code": map[string]any{
			"coding": []map[string]any{
				{
					"system":  "http://loinc.org",
					"code":    "8867-4",
					"display": "Heart rate",
				},
			},
			"text": "Heart Rate",
		},
		"subject": map[string]any{
			"reference": "Patient/" + patientID,
		},
		"valueQuantity": map[string]any{
			"value":  bpm,
			"unit":   "bpm",
			"system": "http://unitsofmeasure.org",
			"code":   "/min",
		},
	}
	b, _ := json.Marshal(obs)
	return b
}

// newSimpleObservation builds a FHIR Observation with a single valueQuantity.
func newSimpleObservation(patientID, loincCode, loincDisplay, text string, value float64, unit, unitCode string) json.RawMessage {
	obs := map[string]any{
		"resourceType": "Observation",
		"status":       "final",
		"code": map[string]any{
			"coding": []map[string]any{
				{
					"system":  "http://loinc.org",
					"code":    loincCode,
					"display": loincDisplay,
				},
			},
			"text": text,
		},
		"subject": map[string]any{
			"reference": "Patient/" + patientID,
		},
		"valueQuantity": map[string]any{
			"value":  value,
			"unit":   unit,
			"system": "http://unitsofmeasure.org",
			"code":   unitCode,
		},
	}
	b, _ := json.Marshal(obs)
	return b
}

func NewTemperatureObservation(patientID string, celsius float64) json.RawMessage {
	return newSimpleObservation(patientID, "8310-5", "Body temperature", "Temperature", celsius, "°C", "Cel")
}

func NewOxygenSaturationObservation(patientID string, percent int) json.RawMessage {
	return newSimpleObservation(patientID, "2708-6", "Oxygen saturation", "O2 Saturation", float64(percent), "%", "%")
}

func NewRespiratoryRateObservation(patientID string, perMin int) json.RawMessage {
	return newSimpleObservation(patientID, "9279-1", "Respiratory rate", "Respiratory Rate", float64(perMin), "/min", "/min")
}

func NewBloodGlucoseObservation(patientID string, mgDL float64) json.RawMessage {
	return newSimpleObservation(patientID, "2345-7", "Glucose [Mass/volume] in Blood", "Blood Glucose", mgDL, "mg/dL", "mg/dL")
}

func NewTotalCholesterolObservation(patientID string, mgDL float64) json.RawMessage {
	return newSimpleObservation(patientID, "2093-3", "Cholesterol [Mass/volume] in Serum or Plasma", "Total Cholesterol", mgDL, "mg/dL", "mg/dL")
}

func NewBMIObservation(patientID string, value float64) json.RawMessage {
	return newSimpleObservation(patientID, "39156-5", "Body mass index", "BMI", value, "kg/m2", "kg/m2")
}

func NewHbA1cObservation(patientID string, percent float64) json.RawMessage {
	return newSimpleObservation(patientID, "4548-4", "Hemoglobin A1c/Hemoglobin.total in Blood", "HbA1c", percent, "%", "%")
}

func NewCreatinineObservation(patientID string, mgDL float64) json.RawMessage {
	return newSimpleObservation(patientID, "2160-0", "Creatinine [Mass/volume] in Serum or Plasma", "Creatinine", mgDL, "mg/dL", "mg/dL")
}

func NewEGFRObservation(patientID string, value float64) json.RawMessage {
	return newSimpleObservation(patientID, "33914-3", "Glomerular filtration rate/1.73 sq M.predicted", "eGFR", value, "mL/min/1.73m2", "mL/min/{1.73_m2}")
}

// NewCondition builds a FHIR Condition resource with an ICD-10 code.
func NewCondition(patientID, icd10Code, display string) json.RawMessage {
	c := map[string]any{
		"resourceType":   "Condition",
		"clinicalStatus": map[string]any{"coding": []map[string]any{{"system": "http://terminology.hl7.org/CodeSystem/condition-clinical", "code": "active"}}},
		"code": map[string]any{
			"coding": []map[string]any{
				{
					"system":  "http://hl7.org/fhir/sid/icd-10-cm",
					"code":    icd10Code,
					"display": display,
				},
			},
			"text": display,
		},
		"subject": map[string]any{
			"reference": "Patient/" + patientID,
		},
	}
	b, _ := json.Marshal(c)
	return b
}

// NewCarePlan builds a FHIR CarePlan resource.
func NewCarePlan(patientID, title string) json.RawMessage {
	cp := map[string]any{
		"resourceType": "CarePlan",
		"status":       "active",
		"intent":       "plan",
		"title":        title,
		"subject": map[string]any{
			"reference": "Patient/" + patientID,
		},
		"activity": []any{},
	}
	b, _ := json.Marshal(cp)
	return b
}

// NewCarePlanActivity creates a CarePlan activity entry (for appending to a CarePlan).
func NewCarePlanActivity(description string, due string) map[string]any {
	detail := map[string]any{
		"status":      "not-started",
		"description": description,
	}
	if due != "" {
		detail["scheduledString"] = "By " + due
	}
	return map[string]any{
		"detail": detail,
	}
}

// BundleEntry creates a transaction bundle entry for a POST.
func BundleEntry(resourceType string, resource json.RawMessage) map[string]any {
	return map[string]any{
		"resource": json.RawMessage(resource),
		"request": map[string]any{
			"method": "POST",
			"url":    resourceType,
		},
	}
}

// TransactionBundle wraps entries into a FHIR transaction bundle.
func TransactionBundle(entries []map[string]any) json.RawMessage {
	b := map[string]any{
		"resourceType": "Bundle",
		"type":         "transaction",
		"entry":        entries,
	}
	raw, _ := json.Marshal(b)
	return raw
}
