# Ephemeral Resources Security Guidelines

## Overview

Ephemeral resources handle sensitive data (API keys, tokens) that should never be persisted in Terraform state or exposed in logs. This document provides concise security guidelines for developing ephemeral resources safely.

## Key Implementation Insights

- **Framework APIs are sufficient** - Use `SetKey()`/`GetKey()` directly, avoid custom helpers
- **Diagnostic handling is critical** - Always append framework diagnostics to responses  
- **Security through simplicity** - Direct framework usage is safer than abstractions
- **Private data is ephemeral-specific** - Only ephemeral resources need inter-method data storage

## Core Security Principles

### 1. No State Persistence
- Ephemeral resources MUST NEVER store sensitive data in Terraform state
- All sensitive values are computed and returned directly without state storage

### 2. Secure Logging  
- Sensitive data MUST NEVER appear in logs, error messages, or debug output
- Use `LogEphemeralOperation()/LogEphemeralError()` utilities for audit trails

### 3. Private Data Protection
- Use framework `SetKey()`/`GetKey()` APIs directly - they're already secure
- Always handle diagnostics properly with `resp.Diagnostics.Append(diags...)`
- Use constants from `fwutils.PrivateDataKey*` to prevent typos

### 4. Framework APIs are Sufficient
- **Key Insight:** Framework provides clean, secure APIs - avoid custom helper abstractions
- Private data is automatically encrypted and isolated by the framework
- Direct API usage is simpler and more maintainable than wrappers

## Secure Coding Patterns

### Framework APIs Usage
```go
// ✅ Private Data - Always handle diagnostics
diags := resp.Private.SetKey(ctx, fwutils.PrivateDataKeyResourceID, []byte(keyID))
resp.Diagnostics.Append(diags...)

data, diags := req.Private.GetKey(ctx, fwutils.PrivateDataKeyResourceID)
if diags.HasError() {
    resp.Diagnostics.Append(diags...)
    return
}

// ✅ Secure Logging - No sensitive data 
fwutils.LogEphemeralOperation("open", "api_key", true)
fwutils.LogEphemeralError("close", "api_key", err)

// ✅ Safe Errors - Generic messages only
resp.Diagnostics.AddError("API Key Retrieval Failed", "Unable to fetch API key data")
```

### Critical Anti-Patterns
```go
// ❌ NEVER: Log sensitive data
log.Printf("API key: %s", apiKey)
log.Printf("Config: %+v", config)

// ❌ NEVER: Ignore diagnostics  
resp.Private.SetKey(ctx, "key", data)  // Missing error handling

// ❌ NEVER: Expose data in error messages
resp.Diagnostics.AddError("Error", fmt.Sprintf("Failed with key: %s", keyValue))
```

## Testing Security Requirements

### State Isolation Testing
Every ephemeral resource MUST include tests that verify sensitive data never appears in state:

```go
func TestEphemeralResource_NoStateStorage(t *testing.T) {
    // Create ephemeral resource
    // Verify sensitive values are accessible during plan/apply
    // Verify sensitive values do not appear in state after completion
    
    // Example verification:
    stateContent := getStateFileContent(t)
    assert.NotContains(t, stateContent, "sensitive_api_key_value")
}
```

### Logging Security Testing
Every ephemeral resource MUST include tests that verify no sensitive data is logged:

```go
func TestEphemeralResource_SecureLogging(t *testing.T) {
    // Capture log output
    logBuffer := &bytes.Buffer{}
    log.SetOutput(logBuffer)
    
    // Perform ephemeral resource operations
    resource.Open(ctx, req, resp)
    
    // Verify logs don't contain sensitive data
    logContent := logBuffer.String()
    assert.NotContains(t, logContent, "api_key_value")
    assert.NotContains(t, logContent, "secret_token")
}
```

## Code Review Checklist

When reviewing ephemeral resource code, verify:

### Security Checklist
- [ ] No sensitive data is logged at any level (DEBUG, INFO, WARN, ERROR)
- [ ] All error messages are generic and don't expose internal details
- [ ] Private data uses helper functions with proper error handling
- [ ] No sensitive data appears in panic messages or stack traces
- [ ] Resource implements secure cleanup in Close() method
- [ ] Tests verify state isolation and logging security

### Implementation Checklist
- [ ] All required interfaces are implemented (Metadata, Schema, Open)
- [ ] Optional interfaces use proper interface detection
- [ ] Error handling is comprehensive and secure
- [ ] Resource follows existing patterns and conventions
- [ ] Tests achieve >95% code coverage
- [ ] Documentation is complete and accurate

## Security Anti-Patterns to Avoid

### ❌ Logging Sensitive Data
```go
// WRONG - Exposes sensitive data in logs
log.Printf("API key: %s", apiKey.GetKey())
log.Printf("Config: %+v", config)

// CORRECT - Use secure logging
fwutils.LogEphemeralOperation("open", "api_key", true)
```

### ❌ Ignoring Diagnostics  
```go
// WRONG - Ignores error handling
resp.Private.SetKey(ctx, "key", data)

// CORRECT - Proper diagnostic handling
diags := resp.Private.SetKey(ctx, "key", data)
resp.Diagnostics.Append(diags...)
```

### ❌ Exposing Data in Error Messages
```go
// WRONG - Leaks internal details
return fmt.Errorf("API error: %v with key: %s", err, keyValue)

// CORRECT - Generic, safe errors
resp.Diagnostics.AddError("API Key Retrieval Failed", "Unable to fetch API key data")
```

## Security Testing Requirements

### Essential Tests
- **State Isolation:** Verify sensitive data never appears in `.tfstate` files
- **Logging Security:** Ensure no sensitive data in log output
- **Error Handling:** Test that error messages don't expose sensitive details
- **Private Data:** Test framework `SetKey()`/`GetKey()` patterns work correctly

### Test Data Rules
- Use mock/fake data only - never real credentials
- Clean up temporary resources immediately after tests

## Incident Response

### If Sensitive Data is Exposed
1. **Immediate Actions:**
   - Rotate any exposed keys/tokens immediately
   - Identify scope of exposure (logs, state files, error messages)
   - Document the incident for security review

2. **Code Fixes:**
   - Remove sensitive data from logs/errors
   - Update affected resources to use secure patterns
   - Add tests to prevent similar issues

3. **Prevention:**
   - Review all ephemeral resources for similar patterns
   - Update security guidelines if necessary
   - Consider additional automated security checks

## Compliance Considerations

### Data Protection Requirements
- Ephemeral resources help meet data protection requirements by avoiding persistent storage
- Sensitive data lifecycle is limited to Terraform execution time
- No sensitive data should cross organizational boundaries through state files

### Audit Requirements
- All ephemeral resource operations should be auditable through secure logs
- Access patterns should be traceable without exposing sensitive values
- Resource lifecycle should be documentable for compliance purposes

## Key Takeaways

- **Use framework APIs directly** - `SetKey()`/`GetKey()` are already secure
- **Never log sensitive data** - Use `LogEphemeralOperation()` for audit trails  
- **Handle diagnostics properly** - Always append framework diagnostics to responses
- **Test security properties** - Verify no data leakage in state, logs, or errors

**When in doubt: don't log it, don't store it, don't expose it.**