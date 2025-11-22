# # **RGS Platform – Architecture Overview**

This document summarizes the architecture of the RGS (Remote Gaming Server) project, including the services, workflows, domain logic, and supporting infrastructure.

## **1. System Components**

### **RGS (Core Service)**

Primary backend service that handles:

-   Session management (launch, verify, revoke)
-   Bet processing & settlement
-   Round generation (provably fair)
-   Compliance enforcement (jurisdictions, limits, audit)
-   Webhook event creation
-   SSE real-time event streaming
-   Outbox deferred settlement processing

### **WalletMock**

A lightweight mock external wallet API:

-   /wallet/debit
-   /wallet/credit
-   In-memory balances
-   HMAC signature verification
-   /health endpoint

### **PostgreSQL**

Relational database storing:

-   Operators, players, sessions
-   Bets & rounds
-   Outbox deferred settlements
-   Webhook queue
-   Audit logs
-   Operator limits

### **Adminer**

Optional UI for inspecting the database.

### **Observability Stack**

-   **Prometheus metrics**
-   **OpenTelemetry tracing** (OTLP)
-   **Structured logging** via Zap

## **2. Core Request Flow**

### **A) Session Launch**

1.  Operator requests /sessions/launch
2.  Player auto-created if not existing
3.  Session stored with UUID token
4.  SSE: session.launched
5.  Audit log added

### **B) Bet Placement**

1.  Idempotency check
2.  Compliance checks
3.  Provably-fair RNG
4.  Round created
5.  Bet inserted (processing)
6.  Wallet debit (external)
7.  Win/loss resolution:
    -   Win → try wallet credit
    -   On credit fail → Outbox entry
8.  Bet updated with final status
9.  SSE event
10.  Webhook event persisted
11.  Audit log generated

### **C) Outbox Worker**

Retries failed wallet credits:

-   Marks bet as won on success
-   Schedules next retry on failure
-   Emits SSE & webhook events

### **D) Webhook Worker**

Fetches queued webhook events:

-   Sends POST to operator webhook URL
-   Implements retry/backoff
-   Marks events completed or failed

## **3. SSE Event Streaming**

### **Endpoint:**
```
GET /stream
X-Operator-Key: <api-key>
Last-Event-ID: <optional>
```

Supports:
-   Real-time streaming
-   Resume using Last-Event-ID
-   Emits everything: wallet, rounds, bets, settlements, sessions

EventBus keeps recent events in memory for recovery.

## **4. Compliance Layer**

Provides:

-   Jurisdiction blocking
-   Maximum bet enforcement
-   Daily win/loss limits
-   Writing audit logs for:
    -   sessions
    -   bets
    -   wallet (debit/credit)

Exposed through:
-   /audit endpoint (per operator)

## **5. Observability**

### **Metrics (Prometheus)**

-   rgs_http_requests_total
-   rgs_bets_placed_total
-   rgs_wallet_debit_calls_total
-   rgs_wallet_debit_failures_total
-   rgs_bet_settlement_seconds

### **Tracing**

-   Request tracing via Chi middleware
-   Custom spans inside BetAggregate
-   Exported to OTel Collector → Jaeger

### **Logging**

Structured JSON logs:
-   request metadata
-   worker results
-   retries
-   errors

# Folder structure
```
.
├── ARCHITECTURE.md
├── Dockerfile
├── README.md
├── app.go
├── config.go
├── db.go
├── docker-compose.yml
├── game
│   └── fair.go
├── go.mod
├── go.sum
├── handlers
│   ├── audit_handler.go
│   ├── bets_handler.go
│   ├── outbox_handlers.go
│   ├── rounds_handler.go
│   ├── sessions_handler.go
│   ├── sse_handler.go
│   └── webhook_handlers.go
├── main.go
├── middleware
│   ├── chi_tracing.go
│   ├── operator.go
│   └── rate_limit.go
├── migrations
│   ├── 0001_init.up.sql
│   └── 0002_seed_data.up.sql
├── observability
│   ├── logger.go
│   ├── metrics.go
│   ├── middleware.go
│   └── tracing.go
├── otel-collector-config.yaml
├── services
│   ├── bet_aggregate.go
│   ├── compliance.go
│   ├── eventbus.go
│   ├── outbox_service.go
│   ├── outbox_worker.go
│   ├── sessions.go
│   ├── wallet_client.go
│   ├── webhook_client.go
│   ├── webhook_service.go
│   └── webhook_worker.go
├── sqlc
│   ├── audit.sql.go
│   ├── bets.sql.go
│   ├── compliance.sql.go
│   ├── db.go
│   ├── models.go
│   ├── outbox.sql.go
│   ├── queries
│   │   ├── audit.sql
│   │   ├── bets.sql
│   │   ├── compliance.sql
│   │   ├── outbox.sql
│   │   ├── rounds.sql
│   │   ├── sessions.sql
│   │   └── webhooks.sql
│   ├── rounds.sql.go
│   ├── sessions.sql.go
│   └── webhooks.sql.go
├── sqlc.yaml
├── tests
│   └── testdata
└── walletmock
    ├── Dockerfile
    ├── handlers.go
    ├── main.go
    ├── models.go
    ├── security.go
    ├── server.go
    └── store.go
```


## **7. Non-Functional Requirements Implemented**

### **Security**

-   Operator authentication (X-Operator-Key)
-   Rate limiting middleware
-   Request validation in handlers

### **Stability**

-   Outbox guarantees settlement delivery
-   Webhook worker guarantees event delivery
-   Idempotency ensures safe retries

### **Performance**

-   Minimal DB roundtrips
-   Lightweight wallet integration

### **Scalability**

-   Stateless RGS service
-   Horizontal scalability supported


## **8. Summary**

This RGS implementation includes all essential components of a real gaming backend:

-   Complete bet lifecycle
-   External wallet integration
-   Provably fair RNG
-   Compliance enforcement & auditing
-   Webhooks & SSE streaming
-   Deferred settlement via Outbox pattern
-   Full observability stack
-   Containerized orchestration
-   Migrations + seed data

Designed for clarity, modularity, and production-style robustness.
