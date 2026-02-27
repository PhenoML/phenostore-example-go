# PhenoStore Go SDK Example — Community Health Clinic

An interactive demo for managing patient records at a community health clinic, built with the [PhenoStore Go SDK](https://github.com/PhenoML/phenostore-sdk-go).

Demonstrates FHIR resource CRUD, search with query parameters, transaction bundles, and care plan management through a guided terminal interface.

## Prerequisites

- Go 1.25.7+
- A PhenoStore account with a configured tenant and store

## Configuration

Copy the example env file and fill in your credentials:

```sh
cp .env.example .env
# edit .env with your values
```

The `.env` file is loaded automatically on startup (existing environment variables take precedence). You can also export them directly:

```sh
export PHENOSTORE_URL="https://api.phenostore.example.com"
export PHENOSTORE_CLIENT_ID="your-client-id"
export PHENOSTORE_CLIENT_SECRET="your-client-secret"
export PHENOSTORE_TENANT="your-tenant-id"
export PHENOSTORE_STORE="your-store-id"
```

`PHENOSTORE_URL` must use `https://` in non-local environments (`http://` is only accepted for localhost).

## Build & Run

```sh
go build -o phenostore-example .
./phenostore-example
```

This launches an interactive session with menus and prompts — no flags or subcommands needed.

## Menu Structure

```
Main Menu
├── Seed Sample Data           → creates 5 patients with vitals, labs, conditions, and care plans
├── Patient Summary            → pick patient → full summary view (parallel API calls)
├── Clinic Dashboard           → all active care plans with progress across patients
├── Manage Data
│   ├── Patient Management
│   │   ├── Register New Patient  → form (name, DOB, gender)
│   │   ├── List All Patients     → table view
│   │   ├── View Patient Details  → pick patient → details
│   │   ├── Update Contact Info   → pick patient → phone/email form
│   │   └── Delete Patient        → pick patient → confirm → delete
│   ├── Clinical Records
│   │   ├── Record Vital Signs    → pick patient → pick type → value form
│   │   ├── View Patient Vitals   → pick patient → observation list
│   │   ├── Record Diagnosis      → pick patient → ICD-10 code + name
│   │   └── View Patient Diagnoses → pick patient → condition list
│   └── Health Plans
│       ├── Create New Plan       → pick patient → title
│       ├── Add Activity to Plan  → pick patient → pick plan → description + due date
│       ├── Complete Activity     → pick patient → pick plan → pick activity
│       └── View Plan Status      → pick patient → care plan list
├── Delete Seed Data           → removes only seed-created resources
└── Exit
```

Navigate with arrow keys, press Enter to select, and Ctrl+C to go back or exit.

## SDK Patterns Demonstrated

| Pattern | Where |
|---------|-------|
| `CreateResource` | Register patient, record vitals, record diagnosis, create plan |
| `ReadResource` | View patient, add/complete activity (read-modify-write) |
| `UpdateResource` | Update contact, add/complete activity |
| `DeleteResource` | Delete patient, delete seed data |
| `SearchResources` | List patients |
| `Inner().SearchResourcesWithResponse` | View vitals/diagnoses, plan status, clinic dashboard, tag search |
| `ProcessBundle` (transaction) | Seed sample data |
| `IsNotFound()` error handling | Patient summary |
| Read-modify-write pattern | Update contact, add activity, complete activity |
| Request editors for FHIR search params | View vitals/diagnoses (patient), plan status (patient+status), clinic dashboard (status), tag search |
| Parallel goroutines | Patient summary (4 concurrent API calls) |
| Composed reads | Patient summary (patient + observations + conditions + plans) |

## License

MIT
