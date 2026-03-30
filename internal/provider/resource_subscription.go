package provider

// TODO: Implement when subscriptions API is available.
// Subscriptions will link topics to destinations, allowing fine-grained
// control over which event types are forwarded to which endpoints.

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSubscription() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSubscriptionCreate,
		ReadContext:   resourceSubscriptionRead,
		UpdateContext: resourceSubscriptionUpdate,
		DeleteContext: resourceSubscriptionDelete,
		Schema: map[string]*schema.Schema{
			"topic_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the topic to subscribe to.",
			},
			"destination_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the destination to forward matching events to.",
			},
		},
	}
}

func resourceSubscriptionCreate(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.Errorf("hookverify_subscription is not yet implemented — subscriptions API is not available")
}

func resourceSubscriptionRead(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.Errorf("hookverify_subscription is not yet implemented — subscriptions API is not available")
}

func resourceSubscriptionUpdate(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.Errorf("hookverify_subscription is not yet implemented — subscriptions API is not available")
}

func resourceSubscriptionDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.Errorf("hookverify_subscription is not yet implemented — subscriptions API is not available")
}
