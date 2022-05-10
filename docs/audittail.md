# audittail

While one would be tempted to get a small container that would simply do `tail -f`,
the `audittail` utility intends to replace that while enforcing best-praactices.

In a nutshell, it's a container that will output the audit logs while following
some opinionated approaches on the log files. e.g. It initializes the audit log
file as a named pipe, so logs aren't written to disk and are written to the
container as fast as possible.

All of this in a minimal container that only has this binary (no need for bash).

It may be used on your Deployment as follows:

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  ...
spec:
  ...
  template:
    ...
    spec:
      initContainers:
        # Optional: Pre-creates the `/app-audit/audit.log` named pipe.
        - image: ghcr.io/metal-toolbox/audittail:v0.1.7
          args:
            - 'init'
            - '-f'
            - '/app-audit/audit.log'
          name: init-audit-logs
          volumeMounts:
            - mountPath: /app-audit
              name: audit-logs
      containers:
        - image: MY_APP_IMAGE
          args:
            - --audit-log-path=/app-audit/audit.log
          name: my-app
          volumeMounts:
            - mountPath: /app-audit
              name: audit-logs
        # Mandatory: tails audit logs and outputs them to stdout
        # for the Splunk forwarder to pick up
        - image: ghcr.io/metal-toolbox/audittail:v0.1.7
          args:
            - '-f'
            - '/app-audit/audit.log'
          name: audit-logger
          volumeMounts:
            - mountPath: /app-audit
              name: audit-logs
              readOnly: true
      volumes:
        - name: audit-logs
          emptyDir: {}
```

While the example above took an `initContainer` into use, it is not
strictly necessary, the base `audittail` container (with it's base
command) will do this as well.