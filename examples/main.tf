terraform {
  required_providers {
    hookverify = {
      source  = "hookverify/hookverify"
      version = "~> 0.1"
    }
  }
}

variable "hookverify_api_key" {
  type      = string
  sensitive = true
}

provider "hookverify" {
  api_key  = var.hookverify_api_key
  base_url = "https://hookverify.com"
}

# Create a production destination
resource "hookverify_destination" "production" {
  name        = "Production API"
  url         = "https://api.example.com/webhooks"
  active      = true
  max_retries = 5
}

# Create a staging destination
resource "hookverify_destination" "staging" {
  name        = "Staging API"
  url         = "https://staging.example.com/webhooks"
  active      = true
  max_retries = 3
}

# Read current usage
data "hookverify_usage" "current" {}

# Look up API key info
data "hookverify_api_key" "current" {}

# Send a test webhook to verify connectivity
resource "hookverify_webhook" "test" {
  url     = hookverify_destination.production.url
  payload = jsonencode({
    event     = "terraform_test"
    timestamp = timestamp()
  })
}

output "current_tier" {
  value = data.hookverify_usage.current.tier
}

output "usage_percentage" {
  value = data.hookverify_usage.current.usage_percentage
}

output "production_destination_id" {
  value = hookverify_destination.production.id
}

output "api_key_tier" {
  value = data.hookverify_api_key.current.tier
}

output "test_delivery_id" {
  value = hookverify_webhook.test.delivery_id
}
