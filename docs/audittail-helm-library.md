# audittail-helm-library

A library chart is a type of Helm chart that defines chart primitives or definitions
which can be shared by Helm templates in other charts.
This allows users to share snippets of code that can be re-used across charts, avoiding repetition and keeping charts DRY.

You can use audittail helm library to include audittail initcontainer, sidecarcontainer, volume configuration in application helm charts, like below:

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
  labels:
    app: k8s-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: k8s
  template:
    metadata:
      labels:
        app: k8s
    spec:
      initContainers:
        {{- include "audittail.initContainer" .| nindent 8 }}
      containers:
        - image: nging
          name: nginx
          args:
            - --audit-log-path={{ template "audittail.auditLogPath" }}
          volumeMounts:
            {{- include "audittail.volumeMount" . | nindent 12}}
        {{- include "audittail.sidecarContainer" .| nindent 8 }}
      volumes:
        {{- include "audittail.volume" . | nindent 8}}
```

Just include the audittail helm library chart in their dependencies as they can any other helm dependency.
Example:

```yaml
---
dependencies:
  - name: audittail
    repository: https://github.com/metal-toolbox/auditevent/tree/main/audittail-helm-library
    version: 1.0.0
