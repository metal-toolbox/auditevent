## `AuditEvent` structure

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

### Writing audit logs

The base package comes with a utility structure called `auditevent.EventWriter`. The `EventWriter`'s
purpose is to encode an audit event to whatever representation is needed. This could writing directly
to a file, a UNIX socket, or even an HTTP server. The requirement is that the writer that's passed
to the `EventWriter` structure **must** implement the `io.Writer` interface.

Audit events also need to be encoded somehow, so an encoder must be passed to the `EventWriter`. An
encoder **must** implement the `EventEncoder` interface that's made available in this package.

The creation of an event writer would look as follows:

```golang
aew := auditevent.NewAuditEventWriter(writer, encoder)
```

Since JSON encoding is common and expected, there is a default implementation that assumes
JSON encoding. It's may be used as follows:

```golang
aew := auditevent.NewDefaultAuditEventWriter(writer)
```

To write events to the `EventWriter` one can do so as follows:

```golang
err := aew.Write(eventToWrite)
```

#### Audit event metrics from writer

`auditevent.EventWriter` instances may generate metrics for events and errors.
For more information, [see the metrics documentation.](metrics.md)