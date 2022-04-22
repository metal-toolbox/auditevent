# `auditevent`

[![test](https://github.com/metal-toolbox/auditevent/workflows/test/badge.svg)](https://github.com/metal-toolbox/auditevent/actions/workflows/test.yml)
[![coverage](https://codecov.io/gh/metal-toolbox/auditevent/branch/main/graph/badge.svg?token=GXV4UZ2JF6)](https://codecov.io/gh/metal-toolbox/auditevent)
[![Release](https://github.com/metal-toolbox/auditevent/workflows/Release/badge.svg)](https://github.com/metal-toolbox/auditevent/actions/workflows/release.yml)

A small and flexible library to help you create audit events.

## Context

While audit logging may seem like a very simple thing to add to an application, doing it right is
full of caveats. This project aims to provide a simple, general, intuitive and standardized
representation for an audit event, as well as tools to take this into use. This will help us
have uniform logs and and meet regulatory compliance requirements.

Correct generation of audit events aids us in determining what's happening in our systems,
doing forensic analysis on security incidents, as well as serving as evidence in court in
case of a breach. Hence, why it's important for us to generate correct and accurate
audit events.

As a guide to create this project and gather requirements for it, the NIST SP 800-53 Audit-related
controls were used.

The project provides the following:

### `auditevent`

An library to generate and write audit events.

[Read more.](docs/auditevent.md)

### Gin middleware

Middleware for the [Gin HTTP framework](https://gin-gonic.com/)
which allows us to write audit events.

[Read more.](docs/middleware.md)

### Metrics

The reference `auditevent` writer and the aforementioned Gin Middleware
both have prometheus metric support baked in. 

[Read more.](docs/metrics.md)

### `audittail`

A simple utility to read audit logs and reliably output them.
e.g. in a sidecar container.

[Read more.](docs/audittail.md)