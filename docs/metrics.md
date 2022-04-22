# Metrics

## Contents

This library provides support for [Prometheus metrics](https://prometheus.io/). The
following metrics are registered:

* `audit_events_total`: a simple counter that represents the events persisted by a
  specific writer.

* `audit_errors_total`: a simple counter that represents the errors writing
  audit event that a writer has encountered.

These metrics are useful not only to monitor the functionality of the audit event
generator, but also to be able to react in case there are errors writing audit logs.

The metrics also have a simple label defined, called `component`
which defines the software component that's emitting these
metrics. Normally, this will match the name of the service;
however, a service's sub-component or micro-service name may be used.

## Context

Different compliance standards will have different recommendations on this, but
it is relevant for system administrators to have procedures in place to deal with
audit event write errors.

NIST SP-800-53 Revision 5.1:: Control AU-5 provides the following guidance:

> Audit logging process failures include software and hardware errors,
> failures in audit log capturing mechanisms, and reaching or exceeding
> audit log storage capacity. Organization-defined actions include overwriting
> oldest audit records, shutting down the system, and stopping the generation
> of audit records. Organizations may choose to define additional actions for
> audit logging process failures based on the type of failure, the location of
> the failure, the severity of the failure, or a combination of such factors.
> When the audit logging process failure is related to storage, the response
> is carried out for the audit log storage repository (i.e., the distinct system
> component where the audit logs are stored), the system on which the audit logs
> reside, the total audit log storage capacity of the organization (i.e., all
> audit log storage repositories combined), or all three. Organizations may
> decide to take no additional actions after alerting designated roles or personnel.

By having metrics in place, administrators are able to configure relevant alerts in
order to fulfil such requirements.

## Usage

###  In `auditevent.Writer`

An [`auditevent.EventWriter`](auditevent.md) instance may generate
metrics and use a pre-defined `prometheus.Registerer` as follows:

```golang
aew := auditevent.NewAuditEventWriter(writer, encoder)

// takes component name and registerer
aew.WithPrometheusMetricsForRegisterer("web-server", registerer)
```

This will register the metrics in the `prometheus.Registerer` instance
and set up metrics when writing audit events.

If the code-base is using the default registerer (which is normally the case), the following function may be used:

```golang
// takes component name
aew.WithPrometheusMetrics("service-proxy")
```

**NOTE**: one may only register these metrics once per registerer. This is particularly important with the default
Prometheus registerer. Failing to do so will result in
a `panic`. This decision was taken to ensure we don't
loose information about audit event being generated.

### In Gin Middleware

A `ginaudit.Middleware` instance may generate metrics and use
a pre-defined `prometheus.Registerer` instance as follows:

```golang
mdw := ginaudit.NewMiddleware("my-test-component", eventwriter)

mdw.WithPrometheusMetricsForRegisterer(registerer)
```

This will register the metrics in the `prometheus.Registerer` instance
and set up metrics when writing audit events.
The component name defined in the `Middleware` will
be used as a label in the metrics.

If the code-base is using the default registerer (which is normally the case), the following function may be used:

```golang
mdw.WithPrometheusMetrics()
```

**NOTE**: one may only register these metrics once per registerer. This is particularly important with the default
Prometheus registerer. Failing to do so will result in
a `panic`. This decision was taken to ensure we don't
loose information about audit event being generated.