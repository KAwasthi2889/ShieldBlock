# ShieldBlock DNS Server: Engineering Evolution Roadmap

## Table of Contents

- [Engineering Principles](#engineering-principles)
- [Phase 1: Minimal DNS-over-TLS Forwarder](#phase-1--minimal-dns-over-tls-forwarder)
- [Phase 2: Basic DNS Filtering](#phase-2--basic-dns-filtering)
- [Phase 3: User Authentication & Analytics Writes](#phase-3--user-authentication--direct-analytics-writes)

## Engineering Principles

1. Start with the simplest architecture possible.
2. Add complexity only when bottlenecks are measurable.
3. Optimize only after profiling.
4. Keep latency-critical paths minimal.
5. Prefer lifecycle-aware optimizations.
6. Separate critical and non-critical workloads.
7. Favor observability before scalability.
8. Prefer evidence-driven optimization over intuition.
9. Design around failure, not only success.
10. Document every architectural tradeoff.
11. Every operational dependency increases:
  - maintenance burden
  - operational state
  - cognitive complexity
  - failure surface area
12. Prefer graceful degradation over hard failure.
13. Unbounded queues are hidden outages.
14. Tail latency matters more than average latency.
15. Reject overload early instead of failing unpredictably.
16. Invalid configuration should fail fast.
17. External dependencies must be treated as unreliable.

---

## Phase 1: Minimal DNS-over-TLS Forwarder

### Goal

- Build the smallest possible working DoT resolver.

### Scope

Client → ShieldBlock → Cloudflare

### Features

- DoT server on :853
- valid TLS cert
- accept encrypted DNS queries
- forward upstream
- return responses

### Lightweight Operational Visibility

Very early visibility:

- structured logs
- request timing logs
- connection open/close logs

### Learn

- TLS sockets
- DNS forwarding
- persistent TCP connections
- DNS request lifecycle
- structured logging basics

### Immediate Learning

- You now understand the complete encrypted DNS request lifecycle.

### Bottleneck

- No filtering. No user auth. No analytics.
- Everything depends on upstream.

---

## Phase 2: Basic DNS Filtering

### Goal

- Turn the resolver into an actual ad blocker.

### Scope

- Exact-match blocking only.

### Features

- in-memory hashmap blocklists
- return 0.0.0.0
- basic logging

### Lightweight Testing

Simple correctness tests:

- blocked domain tests
- allowed domain tests
- malformed request tests

### Learn

- DNS response crafting
- hot-path operations
- in-memory lookups
- table-driven testing

### Immediate Improvement

- The project becomes a functioning DNS ad blocker.

### Bottleneck

Problems:

- exact matching only
- no categories
- no user-specific policies
- no scalability concerns yet

---

## Phase 3: User Authentication & Direct Analytics Writes

### Goal

- Introduce personalized filtering and observability.

### Scope

Users authenticate via:

- {config hash}.dns.shieldblock.in

### Features

- extract SNI
- parse user hash
- user bitmask policies
- direct analytical DB writes
- internal metrics
- basic /healthz
- basic /readyz

### Analytics Stored

- user hash
- queried domain
- blocked/allowed
- response latency
- timestamp

### Lightweight Operational Visibility

Minimal metrics endpoint:

- request counters
- blocked counters
- auth lookup counters
- basic latency histograms

### Lightweight Failure Testing

Basic fault simulation:

- DB unavailable
- malformed SNI
- auth lookup failure

### Learn

- SNI-based authentication
- connection identity
- analytics modeling
- write-heavy systems
- operational visibility basics
- failure-oriented thinking
- health endpoint semantics

### Immediate Improvement

Users now get:

- personalized filtering
- policy-based blocking
- query analytics
- measurable resolver behavior

### Failure Discovered

- System behavior became difficult to reason about without visibility.

### Why It Happened

- Architectural complexity increased beyond intuitive debugging.

### Architectural Fix

- Introduce lightweight internal metrics.

### Tradeoff Introduced

- Minor instrumentation overhead.

### Bottleneck

Every DNS query performs:

- authentication lookup
- analytics DB write

Problems:

- repeated auth work
- increased latency
- DB dependency inside hot path

---

