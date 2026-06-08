package helpers_test

import (
	"testing"

	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

func TestIsCloudEndpoint(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		endpoint string
		want     bool
	}{
		{
			name:     "prefect cloud host",
			endpoint: "api.prefect.cloud",
			want:     true,
		},
		{
			name:     "prefect dev host",
			endpoint: "api.prefect.dev",
			want:     true,
		},
		{
			name:     "prefect staging dev host",
			endpoint: "api.stg.prefect.dev",
			want:     true,
		},
		{
			name:     "private cloud previous-api host",
			endpoint: "previous-api.private.prefect.cloud/api",
			want:     true,
		},
		{
			name:     "private cloud next-api host",
			endpoint: "next-api.private.prefect.cloud/api",
			want:     true,
		},
		{
			name:     "private cloud bare host",
			endpoint: "private.prefect.cloud",
			want:     true,
		},
		{
			name:     "private dev latest-api host (customer-managed test cluster)",
			endpoint: "latest-api.private.prefect.dev/api",
			want:     true,
		},
		{
			name:     "private dev bare host",
			endpoint: "private.prefect.dev",
			want:     true,
		},
		{
			name:     "self-hosted server host",
			endpoint: "prefect.example.com",
			want:     false,
		},
		{
			name:     "localhost server",
			endpoint: "localhost:4200",
			want:     false,
		},
		{
			name:     "empty endpoint",
			endpoint: "",
			want:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := helpers.IsCloudEndpoint(tc.endpoint)
			if got != tc.want {
				t.Errorf("IsCloudEndpoint(%q) = %v, want %v", tc.endpoint, got, tc.want)
			}
		})
	}
}
