#!/bin/bash
# Test configuration validation

echo "=== Testing Configuration Validation ==="
echo ""

# Test 1: Missing environment variable
echo "Test 1: Missing environment variable"
cat > /tmp/config_test1.toml << 'EOF'
[server]
host = "0.0.0.0"
port = 8082

[[providers]]
name = "openai"
type = "openai"
api_base_url = "https://api.openai.com/v1"
api_key = "env:NONEXISTENT_API_KEY"
models = ["gpt-4o"]
EOF

CONFIG_PATH=/tmp/config_test1.toml ./llm-to-anthropic serve 2>&1 | grep -E "(Failed|error)" || echo "❌ Should have failed"
echo ""

# Test 2: Empty models list
echo "Test 2: Empty models list"
cat > /tmp/config_test2.toml << 'EOF'
[server]
host = "0.0.0.0"
port = 8082

[[providers]]
name = "openai"
type = "openai"
api_base_url = "https://api.openai.com/v1"
api_key = "bypass"
models = []
EOF

CONFIG_PATH=/tmp/config_test2.toml ./llm-to-anthropic serve 2>&1 | grep -E "(Failed|error)" || echo "❌ Should have failed"
echo ""

# Test 3: Duplicate provider name
echo "Test 3: Duplicate provider name"
cat > /tmp/config_test3.toml << 'EOF'
[server]
host = "0.0.0.0"
port = 8082

[[providers]]
name = "openai"
type = "openai"
api_base_url = "https://api.openai.com/v1"
api_key = "bypass"
models = ["gpt-4o"]

[[providers]]
name = "openai"
type = "anthropic"
api_base_url = "https://api.anthropic.com"
api_key = "bypass"
models = ["claude-haiku-4-20250514"]
EOF

CONFIG_PATH=/tmp/config_test3.toml ./llm-to-anthropic serve 2>&1 | grep -E "(Failed|error)" || echo "❌ Should have failed"
echo ""

# Test 4: Mapping with nonexistent provider
echo "Test 4: Mapping with nonexistent provider"
cat > /tmp/config_test4.toml << 'EOF'
[server]
host = "0.0.0.0"
port = 8082

[[providers]]
name = "openai"
type = "openai"
api_base_url = "https://api.openai.com/v1"
api_key = "bypass"
models = ["gpt-4o"]

[mappings]
"test" = "nonexistent/model"
EOF

CONFIG_PATH=/tmp/config_test4.toml ./llm-to-anthropic serve 2>&1 | grep -E "(Failed|error)" || echo "❌ Should have failed"
echo ""

# Test 5: Valid configuration
echo "Test 5: Valid configuration"
cat > /tmp/config_test5.toml << 'EOF'
[server]
host = "0.0.0.0"
port = 8082

[[providers]]
name = "ollama"
type = "openai"
api_base_url = "http://localhost:11434/v1"
api_key = "bypass"
models = ["llama3.2:3b"]

[mappings]
"haiku" = "ollama/llama3.2:3b"
EOF

timeout 3 CONFIG_PATH=/tmp/config_test5.toml ./llm-to-anthropic serve 2>&1 | grep -E "(Starting|error)" && echo "✅ Should have started"
echo ""

echo "=== Validation Tests Complete ==="
