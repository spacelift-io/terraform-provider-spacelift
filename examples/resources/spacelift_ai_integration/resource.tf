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
# Leaving models unset accepts whatever default list the provider offers.
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
