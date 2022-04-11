# `auditevent`

A small and flexible library to help you create audit events.

## Context

While audit logging may seem like a very simple thing to add to an application, doing it right is
full of caveats. This library aims to provide a simple, general, intuitive and standardize
representation for an audit event. This will help us having uniform logs and and meet
regulatory compliance requirements.

Correct generation of audit events aids us in determining what's happening in our systems,
doing forensic analysis on security incidents, as well as serving as evidence in court in
case of a breach. Hence, why it's important for us to generate correct and accurate
audit events.

As a guide to create this library and gather requirements for it, the NIST SP 800-53 Audit-related
controls were used.

# `AuditEvent` structure

This is the main structure with which to generate events.

per NIST SP-800-53 Revision 5.1:: Control AU-3

It aims to fulfil the following requirement:

> Ensure that audit records contain information that establishes the following:
>
> a. What type of event occurred;
>
> b. When the event occurred;
>
> c. Where the event occurred;
>
> d. Source of the event;
>
> e. Outcome of the event; and
>
> f. Identity of any individuals, subjects, or objects/entities associated with the event.

As a utility function (and the recommended way to create an event), the function
`NewAuditEvent` was created. It shall be called as follows:

```golang
e := auditevent.NewAuditEvent(
    "UserCreate",
    auditevent.EventSource{
        Type:  "IP",
        Value: "127.0.0.1",
    },
    "Success",
    map[string]string{
        "username": "test",
    },
    "test-component",
).WithTarget(map[string]string{
    "path":    "/user",
    "newUser": "foobar",
})
```

Calling this function generates an appropriate and unique Audit ID which is stored in the
`Metadata` section of the event structure. It also will automatically set the `LoggedAt` time,
which indicates when the message was logged. The `LoggedAt` value will already have the `UTC`
location set, which is recommended per NIST SP 800-53 control AU-8 section b:

> The information system:
>
> b. Records time stamps for audit records that can be mapped to Coordinated Universal Time (UTC)
> or Greenwich Mean Time (GMT) and meets [Assignment: organization-defined granularity of
> time measurement].

Note that this depends on the cluster having appropriate NTP configuration coming from an
authoritative source.

Whenever extra information is needed, it shall be placed in the `Data` section of the structure
as an appropriately formatted JSON string. e.g.

```golang
extraData := map[string]string {
    "httpMethod": "GET",
    "httpHeaders": headersMapStr,
}
jsonData, err := json.Marshal(extraData)
if err != nil {
    panic(err)
}
e.WithData(jsonData)
```