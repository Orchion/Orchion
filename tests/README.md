# Integration Tests

Scripts for testing the Orchion system end-to-end.

## Available Tests

### `test-api.ps1`

Tests the Orchion REST API by fetching nodes from the running orchestrator.

**Usage:**
```powershell
.\tests\test-api.ps1
```

**Requirements:**
- Orchestrator must be running (`.\shared\scripts\run-all.ps1`)
- Tests the `/api/nodes` endpoint

### `test-job.ps1`

Tests job submission and execution through the Orchion API.

**Usage:**
```powershell
.\tests\test-job.ps1
```

**Requirements:**
- Both orchestrator and node-agent must be running (`.\shared\scripts\run-all.ps1`)
- Tests the `/v1/chat/completions` endpoint with a sample chat completion request

## Running Tests

1. Start the system:
   ```powershell
   .\shared\scripts\run-all.ps1
   ```

2. Run individual tests:
   ```powershell
   .\tests\test-api.ps1
   .\tests\test-job.ps1
   ```

## Notes

- These are integration tests that require a running Orchion system
- Use the scripts in `shared/scripts/` to manage the system lifecycle
- Tests will show helpful error messages if components aren't running