package api

import (
	"context"
	"encoding/json"
)

type CollectionsClient interface {
	GetWorkerMetadataViews(ctx context.Context) (WorkerTypeByPackage, error)
}

// { "prefect": {...}, "prefect-aws": {...} }.
type WorkerTypeByPackage map[string]MetadataByWorkerType

// { "ecs": {...} }.
type MetadataByWorkerType map[string]WorkerMetadata

type WorkerMetadata struct {
	Type                        string          `json:"type"`
	DocumentationURL            string          `json:"documentation_url"`
	DisplayName                 string          `json:"display_name"`
	InstallCommand              string          `json:"install_command"`
	Description                 string          `json:"description"`
	DefaultBaseJobConfiguration json.RawMessage `json:"default_base_job_configuration"`
}
