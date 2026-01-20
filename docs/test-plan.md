# Orchion Test Plan

## Overview

This document outlines a comprehensive testing strategy for the Orchion distributed AI orchestration platform. The testing approach covers unit tests, integration tests, and end-to-end tests to ensure system reliability, performance, and correctness across all components.

## Testing Strategy

### Test Pyramid
```
End-to-End Tests (Playwright)
    ├── Integration Tests (API/Contract)
        └── Unit Tests (Component Logic)
```

### Test Categories

1. **Unit Tests**: Test individual functions, methods, and components in isolation
2. **Integration Tests**: Test component interactions, API contracts, and data flow
3. **End-to-End Tests**: Test complete user workflows through the full system

---

## 1. Unit Tests

### 1.1 Go Backend Components

#### Testing Framework
- **Primary**: Go's built-in `testing` package
- **Assertions**: `testify/assert` and `testify/require`
- **Mocking**: `testify/mock` for interfaces, `httptest` for HTTP
- **Coverage**: `go test -cover` with minimum 80% coverage target

#### Orchestrator Unit Tests (`orchestrator/`)

**File Structure:**
```
orchestrator/
├── internal/
│   ├── node/
│   │   ├── registry_test.go
│   │   └── registry_benchmark_test.go
│   ├── orchestrator/
│   │   └── service_test.go
│   ├── scheduler/
│   │   └── scheduler_test.go
│   ├── queue/
│   │   └── queue_test.go
│   ├── logging/
│   │   └── service_test.go
│   └── gateway/
│       └── gateway_test.go
```

**Test Coverage Requirements:**

**`internal/node/registry_test.go`**
- Node registration and validation
- Capability updates
- Heartbeat tracking and staleness detection
- Concurrent access (race conditions)
- Node removal and cleanup
- Registry listing and filtering
- Error handling (invalid nodes, missing IDs)

**`internal/orchestrator/service_test.go`**
- gRPC method implementations:
  - `RegisterNode` - validation, error cases, success paths
  - `Heartbeat` - node updates, missing nodes
  - `ListNodes` - data transformation, filtering
  - `UpdateNode` - capability updates, validation
  - `ChatCompletion` - job creation, routing, streaming
- Input validation and error responses
- gRPC status code correctness
- Context handling and timeouts

**`internal/scheduler/scheduler_test.go`**
- Round-robin node selection
- Node availability checking
- Load balancing algorithms
- Scheduler state management
- Error handling (no available nodes)

**`internal/queue/queue_test.go`**
- Job enqueue/dequeue operations
- Queue persistence (SQLite operations)
- Job status updates
- Concurrent queue access
- Queue capacity limits
- Job prioritization

**`internal/logging/service_test.go`**
- Log entry creation and validation
- gRPC streaming implementation
- Log filtering and querying
- Structured logging fields
- Log retention policies

**`internal/gateway/gateway_test.go`**
- HTTP to gRPC translation
- Request/response transformation
- CORS handling
- Error propagation
- Middleware functionality

#### Node Agent Unit Tests (`node-agent/`)

**File Structure:**
```
node-agent/
├── internal/
│   ├── capabilities/
│   │   └── capabilities_test.go
│   ├── heartbeat/
│   │   └── heartbeat_test.go
│   ├── containers/
│   │   ├── manager_test.go
│   │   ├── ollama_test.go
│   │   └── vllm_test.go
│   ├── executor/
│   │   └── executor_test.go
│   └── inference/
│       ├── ollama_test.go
│       ├── vllm_test.go
│       └── service_test.go
```

**Test Coverage Requirements:**

**`internal/capabilities/capabilities_test.go`**
- CPU core detection accuracy
- Memory detection (system RAM vs available)
- OS information parsing
- GPU detection and VRAM reporting
- Power usage monitoring
- Hardware capability edge cases
- Cross-platform compatibility

**`internal/heartbeat/heartbeat_test.go`**
- gRPC client connection management
- Auto-reconnection logic
- Heartbeat interval timing
- Connection failure handling
- Authentication/authorization
- Context cancellation

**`internal/containers/manager_test.go`**
- Container lifecycle management
- Docker/Podman client interactions
- Image pulling and caching
- Port mapping and networking
- Resource limits (CPU, memory, GPU)
- Container health monitoring
- Cleanup and error handling

**`internal/executor/executor_test.go`**
- Job execution workflows
- Model loading and inference
- Result streaming
- Error handling and recovery
- Resource cleanup
- Timeout management
- Concurrent job execution

#### Shared Library Tests (`shared/`)

**File Structure:**
```
shared/
├── logging/
│   └── logger_test.go
└── proto/v1/
    └── validation_test.go
```

**Test Coverage Requirements:**

**`shared/logging/logger_test.go`**
- Log level filtering
- Structured field handling
- Output formatting (JSON, text)
- Concurrent logging safety
- Log rotation and file management

### 1.2 Frontend Tests (Dashboard)

#### Testing Framework
- **Unit Tests**: Vitest (already configured)
- **Component Tests**: Vitest with `@testing-library/svelte`
- **Coverage**: `vitest --coverage`

**File Structure:**
```
dashboard/
├── src/
│   ├── lib/
│   │   ├── orchion.test.ts
│   │   └── components/
│   │       └── *.test.ts
│   ├── routes/
│   │   ├── +page.test.ts
│   │   ├── logs/
│   │   │   └── +page.test.ts
│   │   └── api/
│   │       └── nodes.test.ts
│   └── app.d.ts (types)
```

**Test Coverage Requirements:**

**API Client Tests (`lib/orchion.test.ts`)**
- HTTP request/response handling
- Error parsing and propagation
- Type safety and validation
- Network failure scenarios
- CORS and authentication headers

**Component Tests**
- Node list rendering and updates
- Error state handling
- Loading states
- Real-time data updates
- User interaction handling

### 1.3 VS Code Extension Tests

#### Testing Framework
- **Framework**: `@vscode/test-cli` (already configured)
- **Assertions**: Node.js assert or Chai
- **Mocking**: sinon for VS Code APIs

**File Structure:**
```
vscode-extension/orchion-tools/
├── src/
│   └── test/
│       ├── extension.test.ts
│       ├── commands.test.ts
│       ├── views.test.ts
│       └── api.test.ts
```

**Test Coverage Requirements:**

**Extension Activation (`extension.test.ts`)**
- Extension activation and deactivation
- Command registration
- Tree view provider setup
- Configuration loading

**Command Tests (`commands.test.ts`)**
- Chat opening functionality
- Agent prompting
- Node/agent refresh operations
- Error handling

**View Tests (`views.test.ts`)**
- Tree data provider implementations
- Node list population
- Log streaming display
- Real-time updates

---

## 2. Integration Tests

### 2.1 API Integration Tests

#### Testing Framework
- **Go**: Custom test helpers with `testify`
- **HTTP**: `httptest` server for orchestrator
- **gRPC**: In-memory gRPC server/client
- **Database**: SQLite in-memory for persistence tests

**File Structure:**
```
tests/
├── integration/
│   ├── orchestrator_test.go
│   ├── node_agent_test.go
│   ├── api_contracts_test.go
│   └── database_test.go
```

#### Test Scenarios

**Orchestrator API Tests (`orchestrator_test.go`)**
- Full node lifecycle: register → heartbeat → update → list → remove
- Concurrent node operations
- Heartbeat timeout and cleanup
- Job submission and routing
- Log streaming functionality
- Error scenarios (network failures, invalid data)

**Node Agent Integration (`node_agent_test.go`)**
- Registration with orchestrator
- Heartbeat loop functionality
- Capability reporting accuracy
- Job execution workflows
- Log streaming to orchestrator

**API Contract Tests (`api_contracts_test.go`)**
- HTTP REST API compliance
- gRPC protobuf contract validation
- Request/response schema validation
- Backward compatibility checks
- API versioning

**Database Integration (`database_test.go`)**
- Job queue persistence
- Log storage and retrieval
- Migration testing
- Concurrent database access
- Data integrity checks

### 2.2 Component Integration Tests

**Container Integration**
- Docker/Podman daemon connectivity
- Image management workflows
- Network configuration
- Volume mounting
- Resource allocation

**Logging Integration**
- Log streaming from node-agent to orchestrator
- HTTP SSE broadcasting
- Log aggregation and filtering
- Performance under load

**Scheduler Integration**
- Multi-node job distribution
- Load balancing verification
- Node failure handling
- Job queue integration

---

## 3. End-to-End Tests

### 3.1 Playwright E2E Tests

#### Testing Framework
- **Framework**: Playwright (already configured in dashboard)
- **Browser**: Chromium, Firefox, WebKit
- **CI/CD**: GitHub Actions with screenshots/videos on failure

**File Structure:**
```
tests/
├── e2e/
│   ├── dashboard/
│   │   ├── nodes.spec.ts
│   │   ├── logs.spec.ts
│   │   ├── jobs.spec.ts
│   │   └── error-handling.spec.ts
│   ├── vscode-extension/
│   │   ├── cluster-view.spec.ts
│   │   ├── chat.spec.ts
│   │   └── logs.spec.ts
│   └── full-system/
│       └── complete-workflow.spec.ts
```

#### Test Scenarios

**Dashboard E2E Tests**
- **Node Management**: Register node → verify display → check capabilities
- **Real-time Updates**: Node heartbeat → UI updates automatically
- **Log Streaming**: View real-time logs from orchestrator
- **Error Handling**: Network failures, invalid responses, recovery
- **Job Management**: Submit job → monitor progress → view results

**VS Code Extension E2E Tests**
- **Cluster View**: Node discovery and status display
- **Chat Interface**: Open chat → send message → receive response
- **Log Monitoring**: Real-time log streaming in VS Code
- **Agent Interaction**: Prompt agents → monitor execution

**Full System Workflow**
- Complete AI inference workflow: submit prompt → route to node → execute model → return result
- Multi-node scenarios: Load balancing, failover, recovery
- System monitoring: Dashboard + VS Code extension synchronization

---

## 4. Test Infrastructure and Tools

### 4.1 Test Environments

#### Local Development
- **Unit Tests**: Run with `go test ./...`
- **Integration Tests**: Use test databases and mocked external services
- **E2E Tests**: Local system with `run-all.ps1` script

#### CI/CD Environment
- **GitHub Actions**: Automated test execution
- **Test Databases**: SQLite in-memory or temporary files
- **Container Registry**: Pre-built test images
- **Browser Testing**: Headless Playwright instances

### 4.2 Test Data and Fixtures

#### Test Data Strategy
- **Factories**: Go factories for creating test nodes, jobs, logs
- **Fixtures**: JSON files for complex test scenarios
- **Mock Data**: Realistic but deterministic test data
- **Edge Cases**: Invalid inputs, boundary conditions, error states

#### Test Utilities
```go
// Example test helper
func createTestNode(id string) *pb.Node {
    return &pb.Node{
        Id:       id,
        Hostname: "test-host-" + id,
        Capabilities: &pb.Capabilities{
            Cpu:    "4 cores",
            Memory: "8GB",
            Os:     "linux",
        },
        LastSeenUnix: time.Now().Unix(),
    }
}
```

### 4.3 Test Execution and Reporting

#### Test Commands
```bash
# Unit tests
go test ./orchestrator/...
go test ./node-agent/...
npm run test:unit  # Dashboard

# Integration tests
go test ./tests/integration/...

# E2E tests
npm run test:e2e   # Dashboard
npx playwright test # VS Code extension

# Coverage reports
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Coverage Requirements
- **Unit Tests**: 80% minimum coverage
- **Integration Tests**: 70% API coverage
- **E2E Tests**: Critical user paths only

#### Test Reporting
- **JUnit XML**: For CI/CD integration
- **Coverage Reports**: HTML and JSON formats
- **Test Results**: Structured output for dashboards
- **Performance Metrics**: Test execution times, resource usage

---

## 5. Test Automation and CI/CD

### 5.1 GitHub Actions Workflow

```yaml
name: CI/CD Pipeline
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
      
      - name: Install dependencies
        run: |
          go mod download
          npm ci
      
      - name: Run unit tests
        run: |
          go test -race -coverprofile=coverage.out ./...
          npm run test:unit
      
      - name: Run integration tests
        run: ./scripts/test-integration.sh
      
      - name: Run E2E tests
        run: npm run test:e2e
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
```

### 5.2 Test Scripts

#### PowerShell Scripts (Windows)
- `test-unit.ps1`: Run all unit tests
- `test-integration.ps1`: Run integration tests with full system
- `test-e2e.ps1`: Run end-to-end tests
- `test-coverage.ps1`: Generate coverage reports

#### Bash Scripts (Cross-platform)
- `scripts/test-unit.sh`: Go unit tests
- `scripts/test-integration.sh`: Full system integration
- `scripts/test-e2e.sh`: Playwright E2E tests

---

## 6. Test Maintenance and Best Practices

### 6.1 Test Organization
- **Test Files**: Co-locate with source code (`*_test.go`)
- **Test Helpers**: Shared utilities in `testutils/` packages
- **Test Data**: Centralized fixtures and factories
- **Mocks**: Interface mocks for external dependencies

### 6.2 Test Quality Guidelines
- **Descriptive Names**: `TestRegisterNode_WithValidInput_ReturnsSuccess`
- **Single Responsibility**: One scenario per test function
- **Independent Tests**: No shared state between tests
- **Fast Execution**: Unit tests < 100ms, integration < 5s
- **Flaky Test Prevention**: Avoid timing dependencies, use proper synchronization

### 6.3 Continuous Testing
- **Pre-commit Hooks**: Run unit tests before commits
- **PR Gates**: Require passing tests for merge
- **Nightly Runs**: Full test suite on main branch
- **Performance Regression**: Track test execution times

### 6.4 Debugging and Troubleshooting
- **Test Debugging**: Use `dlv test` for Go tests
- **Log Analysis**: Structured logging in tests
- **Screenshot/Video**: Playwright artifacts on failure
- **Test Isolation**: Ensure tests don't interfere with each other

---

## 7. Implementation Roadmap

### Phase 1: Foundation ✅ **COMPLETED**
- [x] Setup test frameworks and dependencies (testify, mocks)
- [x] Create test directory structure
- [x] Implement basic test helpers and fixtures
- [x] Add unit tests for core components (registry, service)

### Phase 2: Unit Test Coverage ✅ **MOSTLY COMPLETED**
- [x] Complete unit tests for orchestrator components (100% coverage achieved)
- [x] Complete unit tests for shared components (98%+ coverage achieved)
- [x] Achieve 95%+ unit test coverage on core components
- [ ] Complete unit tests for node-agent components
- [ ] Add frontend unit tests

### Phase 3: Integration Testing 🔄 **IN PROGRESS**
- [x] Implement basic API integration tests
- [ ] Add comprehensive component integration tests
- [ ] Test database operations
- [ ] Validate gRPC/HTTP contracts

### Phase 4: End-to-End Testing ⏳ **PLANNED**
- [ ] Setup Playwright infrastructure
- [ ] Implement dashboard E2E tests
- [ ] Add VS Code extension E2E tests
- [ ] Create full system workflow tests

### Phase 5: CI/CD and Automation ✅ **COMPLETED**
- [x] Configure GitHub Actions with coverage enforcement
- [x] Add test reporting and coverage analysis
- [x] Implement automated test scripts (test-all.ps1, check-coverage.ps1)
- [x] Setup performance regression monitoring

---

## 8. Success Metrics

### Coverage Targets
- **Unit Tests**: 95% code coverage for all Go components
- **Integration Tests**: 70% API endpoint coverage
- **E2E Tests**: 100% critical user paths

### Quality Metrics
- **Test Execution Time**: < 5 minutes for full suite
- **Flaky Tests**: < 1% failure rate
- **Test Maintenance**: < 2 hours/month
- **CI/CD Reliability**: 99% pipeline success rate

### Performance Benchmarks
- **API Response Time**: < 100ms for node operations
- **Test Startup Time**: < 30 seconds for integration tests
- **Memory Usage**: < 512MB per test run
- **Concurrent Users**: Support 100+ simultaneous operations

This comprehensive test plan ensures Orchion's reliability, performance, and maintainability as it scales from a minimal working system to a production-ready distributed AI platform.