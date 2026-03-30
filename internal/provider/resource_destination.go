package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDestination() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDestinationCreate,
		ReadContext:   resourceDestinationRead,
		UpdateContext: resourceDestinationUpdate,
		DeleteContext: resourceDestinationDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Human-readable name for this destination.",
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "HTTPS URL to forward webhooks to.",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if !strings.HasPrefix(v, "https://") {
						errs = append(errs, fmt.Errorf("%q must start with https://, got: %s", key, v))
					}
					return
				},
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether this destination is active.",
			},
			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3,
				Description: "Maximum number of delivery retry attempts.",
			},
			"retry_delays": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "[0,5,25]",
				Description: "JSON array of retry delay intervals in seconds.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the destination was created.",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// createRequest is the JSON body for POST /v1/endpoints.
type createRequest struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	MaxRetries  int    `json:"max_retries"`
	RetryDelays string `json:"retry_delays"`
}

// createResponse is the JSON response from POST /v1/endpoints.
type createResponse struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	Message string `json:"message"`
}

// endpointItem represents one item in the GET /v1/endpoints response list.
type endpointItem struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Active     bool   `json:"active"`
	MaxRetries int    `json:"max_retries"`
	RetryDelay []int  `json:"retry_delays"`
	CreatedAt  string `json:"created_at"`
}

// listResponse is the JSON response from GET /v1/endpoints.
type listResponse struct {
	Endpoints []endpointItem `json:"endpoints"`
}

func resourceDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	body := createRequest{
		Name:        d.Get("name").(string),
		URL:         d.Get("url").(string),
		MaxRetries:  d.Get("max_retries").(int),
		RetryDelays: d.Get("retry_delays").(string),
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to marshal request: %w", err))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.baseURL+"/v1/endpoints", bytes.NewReader(jsonBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create request: %w", err))
	}
	req.Header.Set("X-API-Key", client.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("API request failed: %w", err))
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return diag.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result createResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse response: %w", err))
	}

	d.SetId(strconv.Itoa(result.ID))

	// Read back full state
	return resourceDestinationRead(ctx, d, meta)
}

func resourceDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.baseURL+"/v1/endpoints", nil)
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

	if resp.StatusCode == 404 {
		d.SetId("")
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return diag.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var list listResponse
	if err := json.Unmarshal(respBody, &list); err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse response: %w", err))
	}

	// Find our endpoint by ID in the list
	targetID, _ := strconv.Atoi(d.Id())
	var found *endpointItem
	for i := range list.Endpoints {
		if list.Endpoints[i].ID == targetID {
			found = &list.Endpoints[i]
			break
		}
	}

	if found == nil {
		// Resource no longer exists
		d.SetId("")
		return nil
	}

	d.Set("name", found.Name)
	d.Set("url", found.URL)
	d.Set("active", found.Active)
	d.Set("max_retries", found.MaxRetries)
	if found.RetryDelay != nil {
		delayJSON, _ := json.Marshal(found.RetryDelay)
		d.Set("retry_delays", string(delayJSON))
	}
	d.Set("created_at", found.CreatedAt)

	return nil
}

// updateRequest is the JSON body for PUT /v1/endpoints/{id}.
type updateRequest struct {
	Name        string `json:"name,omitempty"`
	URL         string `json:"url,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
	MaxRetries  *int   `json:"max_retries,omitempty"`
	RetryDelays string `json:"retry_delays,omitempty"`
}

func resourceDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	body := updateRequest{}
	if d.HasChange("name") {
		body.Name = d.Get("name").(string)
	}
	if d.HasChange("url") {
		body.URL = d.Get("url").(string)
	}
	if d.HasChange("active") {
		v := d.Get("active").(bool)
		body.Enabled = &v
	}
	if d.HasChange("max_retries") {
		v := d.Get("max_retries").(int)
		body.MaxRetries = &v
	}
	if d.HasChange("retry_delays") {
		body.RetryDelays = d.Get("retry_delays").(string)
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to marshal request: %w", err))
	}

	url := fmt.Sprintf("%s/v1/endpoints/%s", client.baseURL, d.Id())
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(jsonBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create request: %w", err))
	}
	req.Header.Set("X-API-Key", client.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("API request failed: %w", err))
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return diag.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return resourceDestinationRead(ctx, d, meta)
}

func resourceDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	url := fmt.Sprintf("%s/v1/endpoints/%s", client.baseURL, d.Id())
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
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

	if resp.StatusCode == 404 {
		// Already deleted
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return diag.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	d.SetId("")
	return nil
}
