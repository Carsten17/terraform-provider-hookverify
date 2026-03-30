# Terraform Provider for HookVerify

Manage [HookVerify](https://hookverify.com) webhook destinations and monitor usage as infrastructure-as-code.

## Prerequisites

- [Go](https://go.dev/) 1.21+
- [Terraform](https://www.terraform.io/) 1.0+
- A HookVerify account with an API key (`hv_xxxxx`)

## Building

```bash
git clone https://github.com/hookverify/terraform-provider-hookverify.git
cd terraform-provider-hookverify
go build -o terraform-provider-hookverify
```

## Local Development

Add a dev override to your `~/.terraformrc` so Terraform uses your local build:

```hcl
provider_installation {
  dev_overrides {
    "hookverify/hookverify" = "/path/to/terraform-provider-hookverify"
  }
  direct {}
}
```

Then run Terraform without `terraform init` (dev overrides skip the registry):

```bash
cd examples
terraform plan -var="hookverify_api_key=hv_your_key_here"
```

## Provider Configuration

```hcl
provider "hookverify" {
  api_key  = var.hookverify_api_key   # Required. Also reads HOOKVERIFY_API_KEY env var.
  base_url = "https://hookverify.com" # Optional. Default: https://hookverify.com
}
```

| Attribute  | Type   | Required | Default                    | Description                     |
|------------|--------|----------|----------------------------|---------------------------------|
| `api_key`  | string | Yes      | `$HOOKVERIFY_API_KEY`      | Your HookVerify API key         |
| `base_url` | string | No       | `https://hookverify.com`   | API base URL                    |

## Resources

### hookverify_destination

Manages a webhook forwarding destination.

```hcl
resource "hookverify_destination" "production" {
  name        = "Production API"
  url         = "https://api.example.com/webhooks"
  active      = true
  max_retries = 5
  retry_delays = "[0,5,25,60]"
}
```

| Attribute      | Type   | Required | Default     | Description                              |
|----------------|--------|----------|-------------|------------------------------------------|
| `name`         | string | Yes      |             | Human-readable destination name          |
| `url`          | string | Yes      |             | HTTPS URL to forward webhooks to         |
| `active`       | bool   | No       | `true`      | Whether the destination is active        |
| `max_retries`  | int    | No       | `3`         | Maximum retry attempts                   |
| `retry_delays` | string | No       | `[0,5,25]`  | JSON array of retry delays (seconds)     |

**Read-only attributes:**

| Attribute    | Type   | Description                       |
|--------------|--------|-----------------------------------|
| `id`         | string | Server-assigned destination ID    |
| `created_at` | string | Creation timestamp                |

### hookverify_webhook

Send a test webhook to validate endpoint connectivity. This is a "fire and forget" resource — Create sends the webhook, Read returns current state, and Delete is a no-op (webhooks cannot be unsent).

```hcl
resource "hookverify_webhook" "connectivity_test" {
  url     = "https://api.example.com/webhooks"
  payload = jsonencode({
    event     = "terraform_test"
    timestamp = timestamp()
  })
}
```

| Attribute     | Type   | Required | Default            | Description                         |
|---------------|--------|----------|--------------------|-------------------------------------|
| `url`         | string | Yes      |                    | Destination URL to test             |
| `payload`     | string | No       | Test event JSON    | JSON payload to send                |

**Read-only attributes:**

| Attribute     | Type   | Description                       |
|---------------|--------|-----------------------------------|
| `delivery_id` | string | Server-assigned delivery ID       |
| `status`      | string | Delivery status (e.g., "queued")  |

> **Note:** All fields are `ForceNew` — changing any attribute destroys and recreates the resource, sending a new webhook.

### hookverify_topic (coming soon)

Stub resource for topic management. Returns "not yet implemented" until the topics API is available.

### hookverify_subscription (coming soon)

Stub resource for topic-to-destination subscriptions. Returns "not yet implemented" until the subscriptions API is available.

## Data Sources

### hookverify_usage

Read-only data source for monitoring account usage.

> **Note:** Requires the `GET /v1/usage` endpoint to be implemented on the HookVerify backend.

```hcl
data "hookverify_usage" "current" {}

output "tier" {
  value = data.hookverify_usage.current.tier
}

output "usage_pct" {
  value = data.hookverify_usage.current.usage_percentage
}
```

| Attribute          | Type   | Description                              |
|--------------------|--------|------------------------------------------|
| `tier`             | string | Current subscription tier                |
| `requests_used`    | int    | Requests used this billing period        |
| `monthly_limit`    | int    | Monthly request limit for current tier   |
| `usage_percentage` | float  | Percentage of limit used (0-100+)        |
| `overage_count`    | int    | Requests over the monthly limit          |

### hookverify_api_key

Read-only data source for API key metadata.

> **Note:** The backend `GET /v1/api-keys/me` endpoint currently uses session authentication. It will need to support `X-API-Key` header authentication for Terraform provider usage.

```hcl
data "hookverify_api_key" "current" {}

output "api_key_tier" {
  value = data.hookverify_api_key.current.tier
}

output "requests_this_month" {
  value = data.hookverify_api_key.current.requests_this_month
}
```

| Attribute             | Type   | Description                              |
|-----------------------|--------|------------------------------------------|
| `tier`                | string | Current subscription tier                |
| `requests_this_month` | int    | API requests made this month             |
| `active`              | bool   | Whether the API key is active            |
| `created_at`          | string | Key creation timestamp                   |
| `last_used`           | string | Last API call timestamp                  |

## Importing Existing Resources

Import an existing destination into Terraform state:

```bash
terraform import hookverify_destination.prod 42
```

Where `42` is the destination ID. The Read function fetches the full state from the API.

## Example

See the [examples/](examples/) directory for a complete working configuration.

```bash
cd examples
export HOOKVERIFY_API_KEY="hv_your_key_here"
terraform init
terraform plan
terraform apply
```

## API Reference

The provider communicates with the HookVerify API using these endpoints:

| Method | Path                      | Used By                  |
|--------|---------------------------|--------------------------|
| GET    | `/v1/endpoints`           | destination read (list)  |
| POST   | `/v1/endpoints`           | destination create       |
| PUT    | `/v1/endpoints/{id}`      | destination update       |
| DELETE | `/v1/endpoints/{id}`      | destination delete       |
| POST   | `/v1/webhooks`            | webhook resource (test)  |
| GET    | `/v1/api-keys/me`         | api_key data source      |
| GET    | `/v1/usage`               | usage data source        |

All requests include the `X-API-Key` header for authentication.
