Payment Service
===================

## Logs
To start collecting logs, we need to annotate the pod definition to indicate to the Datadog agent to collect logs and their nature

##### **`kubernetes-manifests/paymentservice.yaml`**
```yaml
annotations:
    ad.datadoghq.com/server.logs: '[{"source":"nodejs","service":"paymentservice", "sourcecategory":"sourcecode", "file":"file"}]'
```
If you have a look at the logs in the Datadog Log Explorer, you will notice that the hostname is the pod name instead of the cluster name. This is because the service uses the Pino library that by default injects a hostname tag in the logs. For the logger, the host of the service is the pod name. We can remove this injection by setting `base: null` when setting it.

##### **`src/paymentservice/server.js`**
```js
const logger = pino({
  name: 'paymentservice-server',
  base: null,
```

## APM

The first step in instrumenting a JS application is to install the Javascript client.

Inside the `paymentservice` directory run:

```sh
npm install --save dd-trace
```

Now we can instrument our application by initiating a tracer

##### **`src/paymentservice/server.js`**
```js
const tracer = require('dd-trace').init()
const path = require('path');
```

Next we need to add a few environment variables to the deployment. I choose to disable `fs` plugin as it is not really relevant in our context

##### **`kubernetes-manifests/paymentservice.yaml`**
```yaml
- name: DD_AGENT_HOST
  valueFrom:
    fieldRef:
      fieldPath: status.hostIP
- name: DD_ENV
  value: "prod"
- name: DD_SERVICE
  value: "paymentservice"
- name: DD_VERSION
  value: "1.0.0"
- name: DD_TRACE_ENABLED
  value: "true"
- name: DD_LOGS_INJECTION
  value: "true"
- name: DD_TRACE_ANALYTICS_ENABLED
  value: "true"
- name: DD_TRACE_DISABLED_PLUGINS
  value: "fs"
```