Email Service
==================

## Logs

To start collecting logs, we need to annotate the pod definition to indicate to the Datadog agent to collect logs and their nature

##### **`kubernetes-manifests/emailservice.yaml`**
```yaml
annotations:
    ad.datadoghq.com/server.logs: '[{"source":"python","service":"emailservice", "sourcecategory":"sourcecode", "file":"file"}]'
```

## APM

To collect traces and enable profiling we will install the Python agent

##### **`src/emailservice/Dockerfile`**
```docker
# Add the application
COPY . .

RUN pip install ddtrace

EXPOSE 8080
ENTRYPOINT [ "/usr/local/bin/ddtrace-run", "python", "email_server.py" ]
```

Finally we add a few environment variables in our deployment

##### **`kubernetes-manifests/emailservice.yaml`**
```yaml
- name: DD_AGENT_HOST
    valueFrom:
    fieldRef:
        fieldPath: status.hostIP
- name: DD_ENV
    value: "prod"
- name: DD_SERVICE
    value: "emailservice"
- name: DD_VERSION
    value: "1.0.0"
- name: DD_TRACE_ENABLED
    value: "true"
- name: DD_LOGS_INJECTION
    value: "true"
- name: DD_TRACE_ANALYTICS_ENABLED
    value: "true"
- name: DD_PROFILING_ENABLED
    value: "true"
```