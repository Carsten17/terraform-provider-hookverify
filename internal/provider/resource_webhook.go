package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWebhookCreate,
		ReadContext:   resourceWebhookRead,
		DeleteContext: resourceWebhookDelete,
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Destination URL to send the test webhook to.",
			},
			"payload": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     `{"event":"terraform_test","source":"terraform"}`,
				Description: "JSON payload to send. Defaults to a test event.",
			},
			"delivery_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server-assigned delivery ID.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Delivery status (e.g., 'queued').",
			},
		},
	}
}

// webhookRequest is the JSON body for POST /v1/webhooks.
type webhookRequest struct {
	URL     string      `json:"url"`
	Payload interface{} `json:"payload"`
}

// webhookResponse is the JSON response from POST /v1/webhooks.
type webhookResponse struct {
	DeliveryID string `json:"deliveryId"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	// Parse the payload string into a generic JSON object
	var payload interface{}
	payloadStr := d.Get("payload").(string)
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		return diag.FromErr(fmt.Errorf("payload must be valid JSON: %w", err))
	}

	body := webhookRequest{
		URL:     d.Get("url").(string),
		Payload: payload,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to marshal request: %w", err))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.baseURL+"/v1/webhooks", bytes.NewReader(jsonBody))
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

	var result webhookResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse response: %w", err))
	}

	d.SetId(result.DeliveryID)
	d.Set("delivery_id", result.DeliveryID)
	d.Set("status", result.Status)

	return nil
}

func resourceWebhookRead(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Webhooks are immutable after creation — return current state as-is.
	return nil
}

func resourceWebhookDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Webhooks cannot be unsent — just remove from Terraform state.
	d.SetId("")
	return nil
}
