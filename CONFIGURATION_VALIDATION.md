# Configuration Validation

This document describes all configuration validation rules performed at startup.

## Validation Flow

1. **Load TOML file** - Parse configuration file
2. **Set defaults** - Apply default values
3. **Parse API keys** - Resolve environment variables
4. **Validate configuration** - Check all rules
5. **Start server** - Only if validation passes

## Server Configuration Validation

### Port
```toml
[server]
port = 8082  # Must be 1-65535
```

**Error:** `invalid server port: 0`

## Provider Configuration Validation

### Required Fields
All providers must have:
- `name` - Unique identifier (string)
- `type` - Provider type (string)
- `api_base_url` - API endpoint URL (string)
- `api_key` - API key configuration (string)
- `models` - List of supported models (array of strings)

### Provider Name
```toml
[[providers]]
name = "openai"  # Required, must be unique
```

**Errors:**
- `provider 0: name is required`
- `duplicate provider name: openai`

### Provider Type
```toml
[[providers]]
type = "openai"  # Required
```

**Error:** `provider openai: type is required`

### API Base URL
```toml
[[providers]]
api_base_url = "https://api.openai.com/v1"  # Required
```

**Error:** `provider openai: api_base_url is required`

### API Key Configuration

Three modes are supported:

#### 1. Direct Key
```toml
[[providers]]
api_key = "sk-xxx"  # Must not be empty
```
**Error:** `provider openai: api_key cannot be empty`

#### 2. Environment Variable
```toml
[[providers]]
api_key = "env:OPENAI_API_KEY"  # Env var must exist and not be empty
```
**Errors:**
- `provider openai: env: mode requires an environment variable name`
- `provider openai: environment variable 'OPENAI_API_KEY' is not set or is empty`

#### 3. Bypass/Forward
```toml
[[providers]]
api_key = "bypass"  # or "forward"
```
Always valid.

### Models List
```toml
[[providers]]
models = ["gpt-4o", "gpt-4.1-mini"]  # Must not be empty
```
**Errors:**
- `provider openai: models list is required and must not be empty`
- `provider openai: model 0: model name cannot be empty`

### Vertex AI Configuration
If `use_vertex_auth = true`:
```toml
[[providers]]
use_vertex_auth = true
vertex_project = "your-project-id"  # Required
vertex_location = "us-central1"      # Required
```
**Errors:**
- `provider openai: vertex_project is required when use_vertex_auth is true`
- `provider openai: vertex_location is required when use_vertex_auth is true`

## Model Mappings Validation

### Mapping Format
```toml
[mappings]
"alias" = "provider/model"  # Must have "/" separator
```
**Errors:**
- `mapping: alias '' cannot be empty`
- `mapping: alias 'test' cannot map to empty string`
- `mapping: alias 'test' maps to invalid format 'provider' (expected 'provider/model')`
- `mapping: alias 'test' maps to invalid model name in 'provider/'`

### Provider Existence
```toml
[mappings]
"alias" = "nonexistent/model"  # Provider must exist
```
**Error:** `mapping: alias 'alias' references non-existent provider 'nonexistent'`

## Validation Examples

### Invalid Configuration 1: Missing Environment Variable
```toml
[[providers]]
name = "openai"
api_key = "env:NONEXISTENT_KEY"
```
**Error:** `provider openai: environment variable 'NONEXISTENT_KEY' is not set or is empty`

### Invalid Configuration 2: Empty Models List
```toml
[[providers]]
name = "openai"
models = []
```
**Error:** `provider openai: models list is required and must not be empty`

### Invalid Configuration 3: Duplicate Provider Name
```toml
[[providers]]
name = "openai"
[[providers]]
name = "openai"  # Duplicate
```
**Error:** `duplicate provider name: openai`

### Invalid Configuration 4: Invalid Mapping
```toml
[mappings]
"alias" = "nonexistent/model"  # Provider doesn't exist
```
**Error:** `mapping: alias 'alias' references non-existent provider 'nonexistent'`

### Valid Configuration
```toml
[server]
port = 8082

[[providers]]
name = "ollama"
type = "openai"
api_base_url = "http://localhost:11434/v1"
api_key = "bypass"
models = ["llama3.2:3b"]

[mappings]
"haiku" = "ollama/llama3.2:3b"
```
**Result:** Server starts successfully

## Running Validation Tests

Run the included test script to verify all validation rules:

```bash
./test_validation.sh
```

This will test:
- Missing environment variables
- Empty models lists
- Duplicate provider names
- Invalid mappings
- Valid configurations
