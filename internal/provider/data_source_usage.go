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

// NOTE: This data source calls GET /v1/usage which must be implemented on the
// HookVerify backend. It should return:
//   { "tier": "...", "requests_used": N, "monthly_limit": N,
//     "usage_percentage": N.N, "overage_count": N }

func dataSourceUsage() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUsageRead,
		Schema: map[string]*schema.Schema{
			"tier": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current subscription tier.",
			},
			"requests_used": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of webhook requests used this month.",
			},
			"monthly_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Monthly webhook limit for the current tier.",
			},
			"usage_percentage": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Percentage of monthly limit used (0-100+).",
			},
			"overage_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of requests over the monthly limit.",
			},
		},
	}
}

// usageResponse is the JSON response from GET /v1/usage.
type usageResponse struct {
	Tier            string  `json:"tier"`
	RequestsUsed    int     `json:"requests_used"`
	MonthlyLimit    int     `json:"monthly_limit"`
	UsagePercentage float64 `json:"usage_percentage"`
	OverageCount    int     `json:"overage_count"`
}

func dataSourceUsageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.baseURL+"/v1/usage", nil)
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

	var usage usageResponse
	if err := json.Unmarshal(respBody, &usage); err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse response: %w", err))
	}

	// Use the API key as a stable ID for this data source
	d.SetId("usage-" + client.apiKey[:8])
	d.Set("tier", usage.Tier)
	d.Set("requests_used", usage.RequestsUsed)
	d.Set("monthly_limit", usage.MonthlyLimit)
	d.Set("usage_percentage", usage.UsagePercentage)
	d.Set("overage_count", usage.OverageCount)

	return nil
}
