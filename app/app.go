package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/phenoml/phenostore-example-go/fhir"
	"github.com/phenoml/phenostore-sdk-go/phenostore"
	"github.com/phenoml/phenostore-sdk-go/phenostore/gen"
)

// App holds the shared client and configuration.
type App struct {
	Client *phenostore.Client
}

// Initialize loads environment variables and creates the PhenoStore client.
func (a *App) Initialize() error {
	_ = godotenv.Load()

	url := os.Getenv("PHENOSTORE_URL")
	clientID := os.Getenv("PHENOSTORE_CLIENT_ID")
	clientSecret := os.Getenv("PHENOSTORE_CLIENT_SECRET")
	tenant := os.Getenv("PHENOSTORE_TENANT")
	store := os.Getenv("PHENOSTORE_STORE")

	if url == "" || clientID == "" || clientSecret == "" || tenant == "" || store == "" {
		return fmt.Errorf("missing required environment variables: PHENOSTORE_URL, PHENOSTORE_CLIENT_ID, PHENOSTORE_CLIENT_SECRET, PHENOSTORE_TENANT, PHENOSTORE_STORE")
	}
	if err := validatePhenoStoreURL(url); err != nil {
		return err
	}

	client, err := phenostore.NewClient(url, clientID, clientSecret, tenant, store)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	a.Client = client
	return nil
}

func extractResources(bundle gen.Bundle) []json.RawMessage {
	if bundle.Entry == nil {
		return nil
	}
	var resources []json.RawMessage
	for _, entry := range *bundle.Entry {
		if entry.Resource != nil {
			resources = append(resources, *entry.Resource)
		}
	}
	return resources
}

func (a *App) fetchAllPatients(ctx context.Context) ([]json.RawMessage, error) {
	count := gen.SearchCount(100)
	params := &gen.SearchResourcesParams{
		UnderscoreCount: &count,
	}
	bundle, err := a.Client.SearchResources(ctx, "Patient", params)
	if err != nil {
		return nil, fmt.Errorf("searching patients: %w", err)
	}
	return extractResources(*bundle), nil
}

func validatePhenoStoreURL(rawURL string) error {
	parsed, err := neturl.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid PHENOSTORE_URL: must be an absolute URL")
	}

	switch strings.ToLower(parsed.Scheme) {
	case "https":
		return nil
	case "http":
		host := strings.ToLower(parsed.Hostname())
		if host == "localhost" || host == "127.0.0.1" || host == "::1" {
			return nil
		}
	}

	return fmt.Errorf("invalid PHENOSTORE_URL: must use https (http is only allowed for localhost)")
}

func (a *App) searchByPatient(ctx context.Context, resourceType, patientID string) ([]json.RawMessage, error) {
	count := gen.SearchCount(50)
	params := &gen.SearchResourcesParams{
		UnderscoreCount: &count,
	}
	resp, err := a.Client.Inner().SearchResourcesWithResponse(
		ctx, a.Client.Tenant(), a.Client.Store(),
		gen.SearchResourcesParamsResourceType(resourceType), params,
		func(ctx context.Context, req *http.Request) error {
			q := req.URL.Query()
			q.Set("patient", patientID)
			req.URL.RawQuery = q.Encode()
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("searching %s: %w", resourceType, err)
	}
	if resp.HTTPResponse.StatusCode >= 400 {
		return nil, fmt.Errorf("search %s failed: HTTP %d", resourceType, resp.HTTPResponse.StatusCode)
	}
	var bundle gen.Bundle
	if err := json.Unmarshal(resp.Body, &bundle); err != nil {
		return nil, fmt.Errorf("parsing %s response: %w", resourceType, err)
	}
	return extractResources(bundle), nil
}

func (a *App) searchCarePlans(ctx context.Context, patientID string) ([]json.RawMessage, error) {
	count := gen.SearchCount(50)
	params := &gen.SearchResourcesParams{
		UnderscoreCount: &count,
	}
	resp, err := a.Client.Inner().SearchResourcesWithResponse(
		ctx, a.Client.Tenant(), a.Client.Store(),
		gen.SearchResourcesParamsResourceType("CarePlan"), params,
		func(ctx context.Context, req *http.Request) error {
			q := req.URL.Query()
			q.Set("patient", patientID)
			q.Set("status", "active")
			req.URL.RawQuery = q.Encode()
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("searching care plans: %w", err)
	}
	if resp.HTTPResponse.StatusCode >= 400 {
		return nil, fmt.Errorf("search failed: HTTP %d", resp.HTTPResponse.StatusCode)
	}
	var bundle gen.Bundle
	if err := json.Unmarshal(resp.Body, &bundle); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return extractResources(bundle), nil
}

func (a *App) resolvePatientName(ctx context.Context, patientID string) string {
	raw, err := a.Client.ReadResource(ctx, "Patient", patientID)
	if err != nil {
		return patientID
	}
	m, err := fhir.Parse(raw)
	if err != nil {
		return patientID
	}
	return fhir.PatientName(m)
}

// searchByTag finds resource IDs tagged with the given _tag value.
func (a *App) searchByTag(ctx context.Context, resourceType, tag string) ([]string, error) {
	count := gen.SearchCount(200)
	params := &gen.SearchResourcesParams{
		UnderscoreCount: &count,
	}
	resp, err := a.Client.Inner().SearchResourcesWithResponse(
		ctx, a.Client.Tenant(), a.Client.Store(),
		gen.SearchResourcesParamsResourceType(resourceType), params,
		func(ctx context.Context, req *http.Request) error {
			q := req.URL.Query()
			q.Set("_tag", tag)
			req.URL.RawQuery = q.Encode()
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("searching %s: %w", resourceType, err)
	}
	if resp.HTTPResponse.StatusCode >= 400 {
		return nil, fmt.Errorf("search %s failed: HTTP %d", resourceType, resp.HTTPResponse.StatusCode)
	}
	var bundle gen.Bundle
	if err := json.Unmarshal(resp.Body, &bundle); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	var ids []string
	for _, raw := range extractResources(bundle) {
		if id := fhir.ResourceID(raw); id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
