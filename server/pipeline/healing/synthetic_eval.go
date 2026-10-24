// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package healing

import (
	"context"
	"fmt"
	"math"
)

// ClinicalParameter describes a single quantitative vital or lab measure in a patient record.
type ClinicalParameter struct {
	Code      string  `json:"code"`       // e.g. "8310-5" (Body temperature), "8867-4" (Heart rate)
	Name      string  `json:"name"`       // e.g. "Body Temperature", "Heart Rate"
	Value     float64 `json:"value"`      // e.g. 38.0
	Unit      string  `json:"unit"`       // e.g. "Cel", "beats/min"
	Threshold float64 `json:"threshold"`  // Decision boundary
}

// ClinicalEvaluationCase represents a test record that triggered a clinical decision boundary failure.
type ClinicalEvaluationCase struct {
	CaseID           string              `json:"case_id"`
	DiagnosisContext string              `json:"diagnosis_context"` // e.g. "Sepsis Early Warning"
	ExpectedOutcome  string              `json:"expected_outcome"`  // e.g. "FLAG_TRIAGE_HIGH"
	ActualOutcome    string              `json:"actual_outcome"`    // e.g. "FLAG_TRIAGE_NORMAL"
	Parameters       []ClinicalParameter `json:"parameters"`
}

// CounterfactualVector contains a bounded perturbation record to test boundary cliff sensitivity.
type CounterfactualVector struct {
	VectorID          string  `json:"vector_id"`
	OriginalValue     float64 `json:"original_value"`
	PerturbedValue    float64 `json:"perturbed_value"`
	DeltaPercentage   float64 `json:"delta_percentage"`
	PerturbedParam    string  `json:"perturbed_param"`
	FHIRResourceJSON  string  `json:"fhir_resource_json"`
	BoundaryCrossed   bool    `json:"boundary_crossed"`
}

// SyntheticCaseGenerator creates bounded counterfactual perturbations around decision boundaries.
type SyntheticCaseGenerator interface {
	GenerateCounterfactualVectors(ctx context.Context, evalCase ClinicalEvaluationCase) ([]CounterfactualVector, error)
}

// BoundedCounterfactualGenerator generates micro-perturbations (±1%, ±2%) adhering to FHIR schemas.
type BoundedCounterfactualGenerator struct{}

// NewBoundedCounterfactualGenerator creates a new generator instance.
func NewBoundedCounterfactualGenerator() *BoundedCounterfactualGenerator {
	return &BoundedCounterfactualGenerator{}
}

// GenerateCounterfactualVectors produces targeted synthetic edge-cases around the boundary cliff.
func (g *BoundedCounterfactualGenerator) GenerateCounterfactualVectors(ctx context.Context, evalCase ClinicalEvaluationCase) ([]CounterfactualVector, error) {
	var vectors []CounterfactualVector

	// Perturbation percentages bounded strictly to ±1.0% and ±2.0%
	deltas := []float64{-0.02, -0.01, 0.01, 0.02}

	for _, param := range evalCase.Parameters {
		for i, delta := range deltas {
			perturbedVal := math.Round((param.Value*(1.0+delta))*100) / 100

			// Determine if perturbation crosses the decision boundary
			boundaryCrossed := false
			if (param.Value < param.Threshold && perturbedVal >= param.Threshold) ||
				(param.Value >= param.Threshold && perturbedVal < param.Threshold) {
				boundaryCrossed = true
			}

			// Generate FHIR R4 Observation compliant JSON
			fhirJSON := fmt.Sprintf(`{
  "resourceType": "Observation",
  "status": "final",
  "code": {
    "coding": [{
      "system": "http://loinc.org",
      "code": "%s",
      "display": "%s"
    }]
  },
  "valueQuantity": {
    "value": %.2f,
    "unit": "%s",
    "system": "http://unitsofmeasure.org"
  },
  "interpretation": [{
    "coding": [{
      "code": "%s"
    }]
  }]
}`, param.Code, param.Name, perturbedVal, param.Unit, map[bool]string{true: "CRITICAL", false: "NORMAL"}[boundaryCrossed])

			vectors = append(vectors, CounterfactualVector{
				VectorID:         fmt.Sprintf("cf-%s-%s-%d", evalCase.CaseID, param.Code, i+1),
				OriginalValue:    param.Value,
				PerturbedValue:   perturbedVal,
				DeltaPercentage:  delta * 100.0,
				PerturbedParam:   param.Name,
				FHIRResourceJSON: fhirJSON,
				BoundaryCrossed:  boundaryCrossed,
			})
		}
	}

	return vectors, nil
}
