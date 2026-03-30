package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// apiClient holds the HTTP client and configuration for API calls.
type apiClient struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// Provider returns the HookVerify Terraform provider schema.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("HOOKVERIFY_API_KEY", nil),
				Description: "HookVerify API key (hv_xxxxx). Can also be set via HOOKVERIFY_API_KEY env var.",
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "https://hookverify.com",
				Description: "Base URL for the HookVerify API.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"hookverify_destination":  resourceDestination(),
			"hookverify_topic":        resourceTopic(),
			"hookverify_subscription": resourceSubscription(),
			"hookverify_webhook":      resourceWebhook(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hookverify_usage":    dataSourceUsage(),
			"hookverify_api_key":  dataSourceAPIKey(),
		},
		ConfigureContextFunc: configureProvider,
	}
}

func configureProvider(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey := d.Get("api_key").(string)
	baseURL := d.Get("base_url").(string)

	if apiKey == "" {
		return nil, diag.Errorf("api_key must be set")
	}

	client := &apiClient{
		httpClient: &http.Client{},
		baseURL:    baseURL,
		apiKey:     apiKey,
	}

	return client, nil
}

// newRequest creates an HTTP request with the X-API-Key header set.
func (c *apiClient) newRequest(method, path string, body interface{}) (*http.Request, error) {
	var req *http.Request
	var err error

	if body != nil {
		// Caller should pass an io.Reader for the body
		req, err = http.NewRequest(method, c.baseURL+path, body.(interface{ Read([]byte) (int, error) }))
	} else {
		req, err = http.NewRequest(method, c.baseURL+path, nil)
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
