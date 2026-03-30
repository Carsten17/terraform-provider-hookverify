package provider

// TODO: Implement when topics API is available.
// Topics will allow grouping webhooks by event type (e.g., "order.created", "payment.failed")
// so that destinations can subscribe to specific event streams.

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTopic() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTopicCreate,
		ReadContext:   resourceTopicRead,
		UpdateContext: resourceTopicUpdate,
		DeleteContext: resourceTopicDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Topic name (e.g., 'order.created').",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Human-readable description of this topic.",
			},
		},
	}
}

func resourceTopicCreate(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.Errorf("hookverify_topic is not yet implemented — topics API is not available")
}

func resourceTopicRead(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.Errorf("hookverify_topic is not yet implemented — topics API is not available")
}

func resourceTopicUpdate(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.Errorf("hookverify_topic is not yet implemented — topics API is not available")
}

func resourceTopicDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.Errorf("hookverify_topic is not yet implemented — topics API is not available")
}
