// Package modules is the root for all business modules in the Modular Monolith.
//
// Per project architecture, each business module follows this internal structure:
//
//   internal/modules/<module>/
//       api/           - HTTP handlers, request/response DTOs, route registration
//       service/       - Business logic, use case orchestration
//       repository/    - Data persistence (interface + implementation)
//       gateway/       - Device capability interface definitions
//       adapter/       - Protocol/transport implementations (Simulator, real devices)
//       model/         - Business entities, enums, states, value objects
//
// Dependency direction:
//
//   api -> service -> model
//   api -> service -> repository (interface defined in service)
//   api -> service -> gateway (interface defined in service)
//   adapter implements gateway interface
//   repository implementation depends on platform/database
//   model does not depend on any outer layer
//
// Module communication rules:
//   - Modules interact through Service layer only
//   - No direct access to another module's Repository or Gateway
//   - No circular dependencies
//
// Future modules (placeholders, no business code in skeleton phase):
//   - device/    - Device management, Gateway, Adapters (ADR-002)
//   - operation/ - Operation/Execution management (ADR-005)
//   - agent/     - Agent Tool boundary (ADR-004)
package modules