## Gin Middleware

### Setup

As a useful utility to use this audit event structure, a gin-based middleware structure is available:
`ginaudit.Middleware`. This structure allows one to set gin routes to log audit events to a
specified `io.Writer` via the aforementioned `auditevent.EventWriter` structure.

One would create a `ginaudit.Middleware` instance as follows:

```golang
mdw := ginaudit.NewMiddleware("my-test-component", eventwriter)
```

Given that JSON is a reasonable default, a utility function that defaults to using
a JSON writer was implemented:

```golang
mdw := ginaudit.NewJSONMiddleware("my-test-component", writer)
```

Here, `writer` is an instance of an structure that implements the `io.Writer` interface.

It is often the case that one must not start to process events until the audit logging
capabilities are set up. For this, the following pattern is suggested:

```golang
fd, err := ginaudit.OpenAuditLogFileUntilSuccess(auditLogPath)
if err != nil {
    panic(err)
}
// The file descriptor shall be closed only if the gin server is shut down
defer fd.Close()

// Set up middleware with the file descriptor
mdw := ginaudit.NewJSONMiddleware("my-test-component", fd)
```

The function `ginaudit.OpenAuditLogFileUntilSuccess` attempts to open the audit log
file, and will block until it's available. This file may be created beforehand or it
may be created by another process e.g. a sidecar container. It opens the file with
`O_APPEND` which enables atomic writes as long as the audit events are less than 4096 bytes.

### Usage

Now that we have a middleware instance available, it's a matter of taking it into
use in our gin `Router`:

```golang
// Get router instance
r := gin.New(...)

// Add middleware
r.Use(mdw.Audit())

// ... All paths after the middleware addition will issue
// audit events
r.GET("/", myGetHandler)
```

Since it's standard gin middleware, it's also possible to set it up per handler:

```golang
r.GET("/foo", mdw.Audit() myGetFooHandler)
```

### Audit event types

Audit event types identify the action that happened on a given request.
By default, the event type will take the following form: `<HTTP Method>:<Path>`.

It is often a best practice to have human readable names, and to have an exhaustive
list of event types that your application may produce. So, in order to
register a type and tie it to a handler, the `RegisterEventType` function is available.
It may be used as follows:

```golang
// Add middleware
r.Use(mdw.Audit())

// ... All paths after the middleware addition will issue
// audit events

mdw.RegisterEventType("ListFoos", http.MethodGet, "/foo")
r.GET("/foo", myGetHandler)

mdw.RegisterEventType("CreateFoo", http.MethodPost, "/foo")
r.POST("/foo", myGetHandler)
```

It's also possible to both set the audit middleware for a specific path and
set a specific audit event type for the path:

```golang
router.GET("/user/:name", mdw.AuditWithType("GetUserInfo"), userInfoHandler)
```

**NOTE**: It is not recommended to assign a default or shared event type
to all events as audit events need to be uniquely identifiable
actions.