resource "spacelift_ai_integration" "gemini" {
  name        = "gemini"
  description = "Gemini API integration"
  space_id    = "root"
  labels      = ["ai"]

  models = ["gemini-2.5-pro", "gemini-2.5-flash"]

  google {
    api_key_wo         = var.gemini_api_key
    api_key_wo_version = "1"
  }
}

# Exactly one provider block is set, and it is what decides the provider.
# Leaving models unset pins nothing, so the integration follows the default
# model list for its provider. Set `models = []` to go back to that after
# pinning models.
resource "spacelift_ai_integration" "claude" {
  name = "claude"

  anthropic {
    api_key_wo         = var.anthropic_api_key
    api_key_wo_version = "1"
  }
}

# base_url points at a self-hosted, provider-compatible gateway such as
# LiteLLM instead of calling the provider directly. Spacelift validates that
# it is reachable when the integration is created.
resource "spacelift_ai_integration" "openai_via_gateway" {
  name = "openai-via-gateway"

  openai {
    api_key_wo         = var.gateway_api_key
    api_key_wo_version = "1"
    base_url           = "https://litellm.example.com"
  }
}

# Bedrock is the exception: it authenticates through an AWS integration
# instead of an API key, and takes its models from the inference profiles
# rather than a models list. Spacelift assumes the integration's role and
# validates the profiles against AWS when the integration is saved.
resource "spacelift_ai_integration" "bedrock" {
  name = "bedrock"

  bedrock {
    integration_id = spacelift_aws_integration.this.id
    region         = "us-east-1"
    profiles = [
      "arn:aws:bedrock:us-east-1:123456789012:inference-profile/global.anthropic.claude-sonnet-4-6",
    ]
  }
}

# Spacelift also provides its own integrations, shared with every account. They
# cannot be created or deleted, so the resource is imported and the `spacelift`
# block marks it as one. Everything except `enabled` belongs to Spacelift, so
# the block takes no arguments and the rest of the resource is left empty.
data "spacelift_ai_integrations" "shared_anthropic" {
  spacelift_provided = true
  ai_provider        = "Anthropic"
}

import {
  to = spacelift_ai_integration.shared_anthropic
  id = one(data.spacelift_ai_integrations.shared_anthropic.integrations[*].integration_id)
}

resource "spacelift_ai_integration" "shared_anthropic" {
  enabled = false

  spacelift {}
}
