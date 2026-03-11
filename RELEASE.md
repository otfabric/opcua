# Release v0.1.6

**Date:** 2026-03-11
**Previous release:** v0.1.5

## Summary

Improves error messages when the server closes the connection (EOF) during
subscription or monitored-item creation. Callers (e.g. `monitor event` /
`monitor alarm` against servers that do not support event subscriptions) now
see a clear hint instead of a raw "EOF".

## Client: EOF handling in subscription path

When the server closes the connection instead of returning a service fault
(e.g. WAGO PLC not supporting OPC UA event or alarm subscriptions), the SDK
previously surfaced **io.EOF** with no context.

- **`Subscription.Monitor()`** — If the request returns `io.EOF`, the
  returned error now wraps it with: "connection closed while creating
  monitored items (server may not support event or alarm subscriptions)".
  Callers can still use `errors.Is(err, io.EOF)`.
- **`Client.Subscribe()`** — If `CreateSubscription` returns `io.EOF`, the
  returned error now wraps it with: "connection closed while creating
  subscription (server may not support subscriptions)".

Documentation and doc comments for `Monitor` and `SubscriptionBuilder.Start`
note that connection-close errors may wrap `io.EOF` with this hint.

---

# Release v0.1.5

**Date:** 2026-03-11
**Previous release:** v0.1.4

## Summary

Adds the depth-limited `WalkLimit` API for browsing the address space and fixes a
nil pointer dereference when encoding `HistoryReadRequest` with optional
`DataEncoding` (e.g. `history value` / `history event` commands).

## Client: WalkLimit (depth-limited walk)

- **`Node.WalkLimit(ctx, maxDepth)`** — Same as `Walk` but stops recursing when
  depth reaches `maxDepth`. The node at `maxDepth` is still yielded. Use for
  "find node", "find type", or "browse tree" style tools to avoid unbounded
  traversal (e.g. a `-depth` flag on the CLI). If `maxDepth < 0`, depth is
  unlimited (equivalent to `Walk`).
- **`Node.Walk(ctx)`** — Unchanged; now implemented via `WalkLimit(ctx, -1)`.

## Bug fix: HistoryReadRequest encoding with nil DataEncoding

Encoding a `HistoryReadRequest` whose `HistoryReadValueID` entries had
`DataEncoding == nil` caused a panic in `QualifiedName.Encode()` (nil pointer
dereference). This affected `HistoryReadRawModified`, `HistoryReadEvent`, and
other history read calls when the optional `DataEncoding` field was omitted.

- **`ua/encode.go`** — Nil pointer fields that implement `BinaryEncoder` are
  now encoded as no bytes instead of calling `Encode()` on a nil receiver.
- **`ua/qualified_name.go`** — `QualifiedName.Encode()` guards against a nil
  receiver and returns `(nil, nil)`.

---

# Release v0.1.4

**Date:** 2026-03-11
**Previous release:** v0.1.3

## Summary

Adds server certificate validation infrastructure and two new client options:
`InsecureSkipVerify()` and `TrustedCertificates()`. When `SecurityMode` is
`Sign` or `SignAndEncrypt`, the client now validates the server certificate by
default. Use `TrustedCertificates()` to trust self-signed servers or private
CAs, or `InsecureSkipVerify()` to disable validation for development.

## Client: Server Certificate Validation

The SDK previously performed no X.509 trust-chain validation of the server's
certificate — it parsed the certificate only to extract the RSA public key for
signing and encryption. This release adds opt-in validation and a deprecation
path toward secure-by-default behavior.

### New Options

| Option | Description |
|--------|-------------|
| `TrustedCertificates(certs ...*x509.Certificate)` | Add CA or self-signed certificates to the trust pool. Merged with the system CA pool. Enables full validation (chain, expiry, key usage). |
| `InsecureSkipVerify()` | Disable all server certificate validation. Certificate is still parsed for its public key. **INSECURE — development only.** |

### Validation Checks (when `TrustedCertificates` is configured)

| Check | Description |
|-------|-------------|
| **Trust chain** | Verifies the certificate chains to a trusted root CA (system pool + user-supplied certs) |
| **Expiration** | Rejects expired or not-yet-valid certificates |
| **Key usage** | Warns if `DigitalSignature` / `KeyEncipherment` bits are missing |

### Validation Points

Server certificate validation runs at two points in the connection flow:

- **`Dial()`** — validates `RemoteCertificate` (set via `SecurityFromEndpoint`
  or `RemoteCertificate` option) after `OpenSecureChannel`
- **`CreateSession()`** — validates `ServerCertificate` from the
  `CreateSessionResponse` after verifying the session signature

### Behavioral Summary

| Scenario | Certificate check | How to configure |
|----------|------------------|------------------|
| `SecurityMode == None` | No certificate exchanged, nothing to validate | Default |
| `Sign` or `SignAndEncrypt` (default) | Full validation: chain, expiry, key usage | Default |
| `Sign` or `SignAndEncrypt` + self-signed server | Fails unless cert added to trust pool | `TrustedCertificates(serverCACert)` |
| `Sign` or `SignAndEncrypt` + skip verify | No validation, just parse for public key | `InsecureSkipVerify()` |

### Config Changes

Added `serverCertValidator` to the internal `Config` struct:

```go
type serverCertValidator struct {
    insecureSkipVerify bool
    trustedCerts       *x509.CertPool
    trustedCertsList   []*x509.Certificate
}
```

## Documentation

- **API.md** — added `InsecureSkipVerify()` and `TrustedCertificates()` to the
  options table
- **docs/security.md** — new "Server Certificate Validation" section with
  usage examples, trust configuration, and dev-mode skip
- **docs/client-guide.md** — added new options to the client options table
- **README.md** — updated security feature description

## Files Changed

6 files changed. Hand-written Go: 3 files (config.go, client.go, config_test.go).

---

# Release v0.1.3

**Date:** 2026-03-11
**Previous release:** v0.1.2

## Summary

Patch release with a small improvement for anonymous authentication when using
the client (e.g. `--auth anonymous` in example CLIs).

## Client: anonymous auth without pre-set PolicyID

When `AuthAnonymous()` is applied before the server's endpoints are known (for
example when using `--auth anonymous` on the command line), the
`AnonymousIdentityToken` is created without a policy ID. The client now
resolves the correct anonymous user token policy from the server's advertised
endpoints after `CreateSession` and sets it on the token, so anonymous
authentication works correctly without requiring endpoint or security options
to be applied first.

---

# Release v0.1.2

**Date:** 2026-03-11
**Previous release:** v0.1.1

## Summary

This release brings the OPC UA schema files up to the latest OPC Foundation
specification and adds a security layer to the server: access restrictions,
role-based access control (RBAC), and session identity-to-role mapping. Three
new code generators were added and the existing service generator was hardened
to skip JSON-only types.

## Schema Update

Updated all schema files from the OPC Foundation UA-Nodeset repository:

- **NodeIds.csv** — refreshed; thousands of new/renamed node identifiers
- **StatusCode.csv** — new status codes added
- **Opc.Ua.Types.bsd** — new structured types and enumerations
- **Opc.Ua.NodeSet2.xml** — expanded node set with role permissions
- **Opc.Ua.PredefinedNodes.xml** — added (new file, used by .NET tooling)
- **AttributeIds.csv** — added (new file, 27 attribute IDs)
- **ServerCapabilities.csv** — added (new file, 39 capability identifiers)
- **Opc.Ua.NodeIds.permissions.csv** — added (new file, 557 default permission entries)

## New Code Generators

| Generator | Input | Output | Description |
|-----------|-------|--------|-------------|
| `cmd/attrid` | `AttributeIds.csv` | `ua/enums_attribute_id_gen.go` | AttributeID enum constants (replaces hand-maintained block) |
| `cmd/capability` | `ServerCapabilities.csv` | `ua/server_capabilities_gen.go` | 39 `ServerCapability*` constants, `KnownCapabilities` map, `ValidateCapability()` |
| `cmd/permissions` | `Opc.Ua.NodeIds.permissions.csv` | `server/default_permissions_gen.go` | 557 default node permission entries for RBAC |

## Code Generation Fixes

- **Service generator** (`cmd/service`): added `-nodeids` flag and
  `filterByBinaryEncoding()` to skip types that only have a JSON encoding in
  the spec — prevents generating codec registrations for types that cannot be
  serialized over OPC UA Binary.
- **`generate.sh`**: updated to run the three new generators; added descriptive
  header comment; fixed shellcheck warnings (SC2035, SC2086).

## Server Security & RBAC

### Access Restrictions (OPC UA Part 3 §5.2.11)

- `checkAccessRestrictions()` enforces `SigningRequired` and
  `EncryptionRequired` bits against the secure channel's security mode.
- `checkAccessRestrictionsForBrowse()` only enforces restrictions when the
  `ApplyRestrictionsToBrowse` bit is set.
- Wired into Read, Write, Browse, and Call service handlers.
- Added `SecurityMode()` getter on `SecureChannel`.

### Role-Based Access Control

- **`RBACAccessController`** — checks node `rolePermissions` against the
  session's assigned roles for Read, Write, Browse, and Call operations.
  Nodes without role permissions are unrestricted.
- **`RoleMapper`** function type and `DefaultRoleMapper` — maps identity tokens
  to well-known role NodeIDs (anonymous → `Anonymous`, others →
  `AuthenticatedUser`). Configurable via `WithRoleMapper()` server option.
- **Session identity tracking** — `ActivateSession` now extracts the
  `UserIdentityToken` and resolves roles through the configured `RoleMapper`.

### Well-Known Roles

- New `ua/well_known_role.go`: 12 well-known roles from the spec (Anonymous,
  AuthenticatedUser, Observer, Operator, Engineer, Supervisor, ConfigureAdmin,
  SecurityAdmin, SecurityKeyServer, SecurityKeyServerAdmin,
  SecurityKeyServerAccess, SecurityKeyServerPush).
- Each role has `String()`, `NodeID()` methods and lookup via `RoleByName` map.

### Node RolePermissions

- Server `Node` stores `[]*ua.RolePermissionType` resolved from the generated
  default permissions at import time via `resolveRolePermissions()`.
- `AttributeIDRolePermissions` and `AttributeIDUserRolePermissions` are served
  from the node as `[]*ua.ExtensionObject`.

## Server Capabilities Expansion

- `OperationalLimits` expanded from 1 field to 12 (all defaulting to 32):
  `MaxNodesPerRead`, `MaxNodesPerWrite`, `MaxNodesPerBrowse`,
  `MaxNodesPerMethodCall`, `MaxNodesPerRegisterNodes`,
  `MaxNodesPerTranslateBrowsePathsToNodeIDs`, `MaxNodesPerNodeManagement`,
  `MaxMonitoredItemsPerCall`, `MaxNodesPerHistoryReadData`,
  `MaxNodesPerHistoryReadEvents`, `MaxNodesPerHistoryUpdateData`,
  `MaxNodesPerHistoryUpdateEvents`.
- Server capability nodes generated dynamically from the struct.

## Code Quality

- Enabled `unparam` linter in `.golangci.yml` — `make check` now catches
  unused parameters.
- Fixed 4 genuine unused-parameter issues in `secure_channel.go` and
  `race_test.go`:
  - `newSecureChannel`: removed 3 dead parameters (`secureChannelID`,
    `sequenceNumber`, `securityTokenID`) — values were only used by
    `NewServerSecureChannel` which sets them on `openingInstance` directly.
  - `sendResponseWithContext`: wired up `ctx` for cancellation checks and
    write deadlines.
  - `mergeChunks`: removed always-nil error return.
  - `race_test.go`: removed unused goroutine parameter.

## Generated Code Changes

All generated files were regenerated from the updated schema:

- `id/` — NodeID constants (DataType, Method, Object, ObjectType, ReferenceType,
  Variable, VariableType)
- `ua/enums_gen.go` — updated/new enum types
- `ua/enums_strings_gen.go` — stringer output for all enums
- `ua/extobjs_gen.go` — extension object codecs
- `ua/register_extobjs_gen.go` — filtered to binary-encoded types only
- `ua/status_gen.go` — new status codes
- `connstate_strings_gen.go` — regenerated

## Files Changed

52 files changed, ~232k insertions, ~53k deletions (bulk is schema XML and
generated code). Hand-written Go: 19 files, +866 / -73 lines.
