# Client Development Guide

> Using the `github.com/otfabric/opcua` client to connect to OPC-UA servers.

---

## Connecting

### Basic Connection

```go
package main

import (
    "context"
    "log"

    "github.com/otfabric/opcua"
    "github.com/otfabric/opcua/ua"
)

func main() {
    ctx := context.Background()

    c, err := opcua.NewClient("opc.tcp://localhost:4840")
    if err != nil {
        log.Fatal(err)
    }
    if err := c.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer c.Close(ctx)

    // Use client...
}
```

### With Security

```go
c, _ := opcua.NewClient("opc.tcp://server:4840",
    opcua.SecurityPolicy("Basic256Sha256"),
    opcua.SecurityMode(ua.MessageSecurityModeSignAndEncrypt),
    opcua.Certificate(cert),
    opcua.PrivateKey(key),
    opcua.AuthUsername("user", "password"),
)
```

### With Auto-Reconnection

```go
c, _ := opcua.NewClient("opc.tcp://server:4840",
    opcua.AutoReconnect(true),
    opcua.ReconnectInterval(5 * time.Second),
)
```

### Endpoint Discovery

Discover available endpoints before connecting:

```go
endpoints, err := opcua.GetEndpoints(ctx, "opc.tcp://server:4840")
if err != nil {
    log.Fatal(err)
}

// Select the best endpoint
ep := opcua.SelectEndpoint(endpoints, "Basic256Sha256", ua.MessageSecurityModeSignAndEncrypt)

// Connect with auto-configured security
c, _ := opcua.NewClient(ep.EndpointURL,
    opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
)
```

---

## Client Options

| Option | Purpose |
|--------|---------|
| `SecurityPolicy(name)` | Security policy (`"None"`, `"Basic256Sha256"`, etc.) |
| `SecurityMode(mode)` | Message security mode |
| `Certificate(cert)` | Client certificate (DER bytes) |
| `PrivateKey(key)` | Client private key |
| `AuthAnonymous()` | Anonymous authentication |
| `AuthUsername(user, pass)` | Username/password authentication |
| `AuthCertificate(cert)` | Certificate-based authentication |
| `AutoReconnect(bool)` | Enable automatic reconnection |
| `ReconnectInterval(d)` | Time between reconnect attempts |
| `SessionTimeout(d)` | Session timeout duration |
| `ApplicationURI(uri)` | Application URI (auto-read from cert if provided) |
| `ApplicationName(name)` | Application name |
| `WithConnStateHandler(fn)` | Connection state change callback |
| `SecurityFromEndpoint(ep, auth)` | Auto-configure from discovered endpoint |

---

## Reading Values

### Single Value

```go
// High-level helper — reads the Value attribute
dv, err := c.ReadValue(ctx, ua.NewNumericNodeID(0, 2258))
if err != nil {
    log.Fatal(err)
}
fmt.Println("Server time:", dv.Value.Value())
```

### Multiple Values

```go
dvs, err := c.ReadValues(ctx,
    ua.NewNumericNodeID(2, 1001),
    ua.NewNumericNodeID(2, 1002),
    ua.NewNumericNodeID(2, 1003),
)
if err != nil {
    log.Fatal(err)
}
for _, dv := range dvs {
    fmt.Println(dv.Value.Value())
}
```

### Using the Node Helper

The `Node` type provides convenience methods for common attributes:

```go
node := c.Node(ua.NewNumericNodeID(2, 1001))

// Read the value
v, err := node.Value(ctx)

// Read other attributes
name, err := node.DisplayName(ctx)
class, err := node.NodeClass(ctx)
desc, err := node.Description(ctx)
browseName, err := node.BrowseName(ctx)
accessLevel, err := node.AccessLevel(ctx)
```

### Low-Level Read

For full control over the read request:

```go
req := &ua.ReadRequest{
    MaxAge:             2000,
    TimestampsToReturn: ua.TimestampsToReturnBoth,
    NodesToRead: []*ua.ReadValueID{
        {
            NodeID:      ua.NewNumericNodeID(2, 1001),
            AttributeID: ua.AttributeIDValue,
        },
    },
}

resp, err := c.Read(ctx, req)
if err != nil {
    log.Fatal(err)
}
for _, result := range resp.Results {
    fmt.Println(result.Value.Value())
}
```

---

## Writing Values

### Single Value

```go
status, err := c.WriteValue(ctx,
    ua.NewNumericNodeID(2, 1001),
    &ua.DataValue{
        EncodingMask: ua.DataValueValue,
        Value:        ua.MustVariant(42.0),
    },
)
if err != nil {
    log.Fatal(err)
}
if status != ua.StatusOK {
    log.Fatal("write failed:", status)
}
```

### Multiple Values

```go
statuses, err := c.WriteValues(ctx,
    &ua.WriteValue{
        NodeID:      ua.NewNumericNodeID(2, 1001),
        AttributeID: ua.AttributeIDValue,
        Value: &ua.DataValue{
            EncodingMask: ua.DataValueValue,
            Value:        ua.MustVariant(42.0),
        },
    },
    &ua.WriteValue{
        NodeID:      ua.NewNumericNodeID(2, 1002),
        AttributeID: ua.AttributeIDValue,
        Value: &ua.DataValue{
            EncodingMask: ua.DataValueValue,
            Value:        ua.MustVariant("hello"),
        },
    },
)
```

---

## Browsing

### Browse All References

```go
refs, err := c.BrowseAll(ctx, ua.NewNumericNodeID(0, id.ObjectsFolder))
if err != nil {
    log.Fatal(err)
}
for _, ref := range refs {
    fmt.Printf("%s: %s (%v)\n", ref.NodeID.NodeID, ref.BrowseName.Name, ref.NodeClass)
}
```

### Node Iterator

The `Node` type provides an `iter.Seq2` iterator that handles continuation points:

```go
node := c.Node(ua.NewNumericNodeID(0, id.ObjectsFolder))

for ref, err := range node.BrowseAll(ctx,
    id.HierarchicalReferences,
    ua.BrowseDirectionForward,
    ua.NodeClassAll,
    true, // Include subtypes
) {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s: %s\n", ref.NodeID.NodeID, ref.BrowseName.Name)
}
```

### Low-Level Browse

```go
req := &ua.BrowseRequest{
    View: &ua.ViewDescription{},
    RequestedMaxReferencesPerNode: 100,
    NodesToBrowse: []*ua.BrowseDescription{
        {
            NodeID:          ua.NewNumericNodeID(0, id.ObjectsFolder),
            BrowseDirection: ua.BrowseDirectionForward,
            ReferenceTypeID: ua.NewNumericNodeID(0, id.HierarchicalReferences),
            IncludeSubtypes: true,
            ResultMask:      uint32(ua.BrowseResultMaskAll),
        },
    },
}

resp, err := c.Browse(ctx, req)
```

---

## Subscriptions

### Using the Builder (Recommended)

The fluent builder API is the simplest way to create subscriptions:

```go
sub, notifyCh, err := c.NewSubscription().
    Interval(100 * time.Millisecond).
    Monitor(
        ua.NewNumericNodeID(2, 1001),
        ua.NewNumericNodeID(2, 1002),
    ).
    Start(ctx)
if err != nil {
    log.Fatal(err)
}
defer sub.Cancel(ctx)

// Process notifications
for msg := range notifyCh {
    if msg.Error != nil {
        log.Println("error:", msg.Error)
        continue
    }
    switch v := msg.Value.(type) {
    case *ua.DataChangeNotification:
        for _, item := range v.MonitoredItems {
            fmt.Printf("Item %d: %v\n", item.ClientHandle, item.Value.Value.Value())
        }
    case *ua.EventNotificationList:
        for _, event := range event.Events {
            fmt.Println("Event:", event.EventFields)
        }
    }
}
```

Builder options:
- `Interval(d)` — Publishing interval
- `LifetimeCount(n)` — Lifetime count
- `MaxKeepAliveCount(n)` — Max keep-alive count
- `MaxNotificationsPerPublish(n)` — Max notifications per publish
- `Priority(p)` — Subscription priority
- `Monitor(nodeIDs...)` — Add nodes for data change monitoring
- `MonitorItems(items...)` — Add custom monitored item requests
- `NotifyChannel(ch)` — Use a custom notification channel
- `Timestamps(ts)` — Timestamps to return

### Manual Subscription

```go
notifyCh := make(chan *opcua.PublishNotificationData, 100)

sub, err := c.Subscribe(ctx, &opcua.SubscriptionParameters{
    Interval: 100 * time.Millisecond,
}, notifyCh)
if err != nil {
    log.Fatal(err)
}
defer sub.Cancel(ctx)

// Add monitored items
miReq := opcua.NewMonitoredItemCreateRequestWithDefaults(
    ua.NewNumericNodeID(2, 1001),
    ua.AttributeIDValue,
    0, // client handle
)

res, err := sub.Monitor(ctx, ua.TimestampsToReturnBoth, miReq)
if err != nil {
    log.Fatal(err)
}

// Process notifications
for msg := range notifyCh {
    // ...
}
```

### Monitor Package (Callback-Based)

The `monitor` package provides a higher-level API with callbacks:

```go
import "github.com/otfabric/opcua/monitor"

m, err := monitor.NewNodeMonitor(c)
if err != nil {
    log.Fatal(err)
}

// Channel-based
ch := make(chan *monitor.DataChangeMessage, 16)
sub, err := m.ChanSubscribe(ctx, &opcua.SubscriptionParameters{
    Interval: 500 * time.Millisecond,
}, ch,
    "ns=2;s=Temperature",
    "ns=2;s=Pressure",
)

for msg := range ch {
    if msg.Error != nil {
        log.Println(msg.Error)
        continue
    }
    fmt.Printf("%s: %v\n", msg.NodeID, msg.Value)
}

// Or callback-based
sub, err := m.Subscribe(ctx, &opcua.SubscriptionParameters{
    Interval: 500 * time.Millisecond,
}, func(s *monitor.Subscription, msg *monitor.DataChangeMessage) {
    fmt.Printf("%s: %v\n", msg.NodeID, msg.Value)
},
    "ns=2;s=Temperature",
)
```

---

## Calling Methods

### High-Level Helper

```go
result, err := c.CallMethod(ctx,
    ua.NewNumericNodeID(2, 1001),  // object node ID
    ua.NewNumericNodeID(2, 2001),  // method node ID
    42.0,                           // arguments (auto-wrapped in Variants)
    "hello",
)
if err != nil {
    log.Fatal(err)
}
if result.StatusCode != ua.StatusOK {
    log.Fatal("method failed:", result.StatusCode)
}
for _, v := range result.OutputArguments {
    fmt.Println("output:", v.Value())
}
```

### Low-Level Call

```go
req := &ua.CallMethodRequest{
    ObjectID: ua.NewNumericNodeID(2, 1001),
    MethodID: ua.NewNumericNodeID(2, 2001),
    InputArguments: []*ua.Variant{
        ua.MustVariant(42.0),
        ua.MustVariant("hello"),
    },
}

result, err := c.Call(ctx, req)
```

---

## Connection State Monitoring

Track connection state changes for monitoring and alerting:

```go
c, _ := opcua.NewClient("opc.tcp://server:4840",
    opcua.AutoReconnect(true),
    opcua.ReconnectInterval(5 * time.Second),
    opcua.WithConnStateHandler(func(state opcua.ConnState) {
        switch state {
        case opcua.Connected:
            log.Println("Connected to server")
        case opcua.Disconnected:
            log.Println("Disconnected — will attempt reconnection")
        case opcua.Reconnecting:
            log.Println("Reconnecting...")
        case opcua.Closed:
            log.Println("Connection closed")
        }
    }),
)
```

### Connection States

| State | Meaning |
|-------|---------|
| `Closed` | Not connected (initial state or permanently closed) |
| `Connecting` | First connection attempt in progress |
| `Connected` | Active session, ready for operations |
| `Disconnected` | Lost connection (auto-reconnect will trigger) |
| `Reconnecting` | Recovery in progress |

---

## Session Management

### Session Transfer

Detach a session from one channel and reattach it to another:

```go
// Detach from current client
session, err := c1.DetachSession(ctx)
if err != nil {
    log.Fatal(err)
}

// Create new client with channel only
c2, _ := opcua.NewClient(endpoint, opts...)
c2.Dial(ctx)

// Reattach session
err = c2.ActivateSession(ctx, session)
```

This is useful for failover scenarios or migrating between connections.

---

## Error Handling

The library provides sentinel errors for common failure modes:

```go
import "github.com/otfabric/opcua/errors"

// Check for specific errors
if errors.Is(err, errors.ErrTimeout) {
    // Handle timeout
}
if errors.Is(err, errors.ErrNotConnected) {
    // Handle disconnection
}

// Status codes from the server
if dv.Status != ua.StatusOK {
    fmt.Println("Server returned:", dv.Status)
}
```

### Common Status Codes

| Code | Meaning |
|------|---------|
| `StatusOK` | Success |
| `StatusBadNodeIDUnknown` | Node does not exist |
| `StatusBadAttributeIDInvalid` | Attribute not supported |
| `StatusBadUserAccessDenied` | Insufficient permissions |
| `StatusBadTimeout` | Operation timed out |
| `StatusBadSessionIDInvalid` | Session expired or invalid |
| `StatusBadSecureChannelIDInvalid` | Channel needs renewal |

---

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/otfabric/opcua"
    "github.com/otfabric/opcua/ua"
)

func main() {
    ctx := context.Background()

    c, err := opcua.NewClient("opc.tcp://localhost:4840",
        opcua.AutoReconnect(true),
        opcua.ReconnectInterval(5 * time.Second),
        opcua.WithConnStateHandler(func(s opcua.ConnState) {
            log.Printf("State: %s", s)
        }),
    )
    if err != nil {
        log.Fatal(err)
    }
    if err := c.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer c.Close(ctx)

    // Read server time
    dv, err := c.ReadValue(ctx, ua.NewNumericNodeID(0, 2258))
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Server time:", dv.Value.Value())

    // Subscribe to value changes
    sub, notifyCh, err := c.NewSubscription().
        Interval(500 * time.Millisecond).
        Monitor(ua.NewNumericNodeID(2, 1001)).
        Start(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer sub.Cancel(ctx)

    for msg := range notifyCh {
        if msg.Error != nil {
            log.Println("error:", msg.Error)
            continue
        }
        if dcn, ok := msg.Value.(*ua.DataChangeNotification); ok {
            for _, item := range dcn.MonitoredItems {
                fmt.Printf("Value changed: %v\n", item.Value.Value.Value())
            }
        }
    }
}
```
