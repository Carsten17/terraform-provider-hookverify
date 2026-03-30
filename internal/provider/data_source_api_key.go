package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NOTE: This data source calls GET /v1/api-keys/me which currently uses
// session-based authentication on the HookVerify backend. The endpoint
// will need to be updated to support X-API-Key header authentication
// for Terraform provider usage. Currently returns: created_at, last_used,
// requests_this_month. The tier and active fields will need to be added
// to the backend response.

func dataSourceAPIKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAPIKeyRead,
		Schema: map[string]*schema.Schema{
			"tier": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current subscription tier.",
			},
			"requests_this_month": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of API requests made this month.",
			},
			"active": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the API key is active.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the API key was created.",
			},
			"last_used": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the API key was last used.",
			},
		},
	}
}

// apiKeyResponse is the JSON response from GET /v1/api-keys/me.
type apiKeyResponse struct {
	Tier              string `json:"tier"`
	RequestsThisMonth int    `json:"requests_this_month"`
	Active            bool   `json:"active"`
	CreatedAt         string `json:"created_at"`
	LastUsed          string `json:"last_used"`
}

func dataSourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.baseURL+"/v1/api-keys/me", nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create request: %w", err))
	}
	req.Header.Set("X-API-Key", client.apiKey)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("API request failed: %w", err))
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return diag.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var info apiKeyResponse
	if err := json.Unmarshal(respBody, &info); err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse response: %w", err))
	}

	d.SetId("apikey-" + client.apiKey[:8])
	d.Set("tier", info.Tier)
	d.Set("requests_this_month", info.RequestsThisMonth)
	d.Set("active", info.Active)
	d.Set("created_at", info.CreatedAt)
	d.Set("last_used", info.LastUsed)

	return nil
}
