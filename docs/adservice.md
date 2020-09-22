Ad Service
=================

## Logs
To start collecting logs, we need to annotate the pod definition to indicate to the Datadog agent to collect logs and their nature

##### **`kubernetes-manifests/adservice.yaml`**
```yaml
annotations:
    ad.datadoghq.com/server.logs: '[{"source":"java","service":"adservice", "sourcecategory":"sourcecode"}]'
```

Because the service is flagged as Java type, and in JSON format, it will be, almost entirely, parsed automatically by Datadog.

But the log message contains context words used to propose ads. This is a relevant information that we can easily extract with a gork parser.

```regex
rule received\s+ad\s+request\s+\(context_words\=%{data:context:array("[]",", ")}\)
```

This will extract context words in an array that we will be able to filter or analyze logs.

## APM

To collect traces, we will add the java agent that will automatically instrument our application as GRPC framework is natively supported in Java.

In the Ad Service Dockerfile, we need to download the java agent:

##### **`src/adservice/Dockerfile`**
```docker
RUN mkdir -p /opt/datadog && \
    wget -qO /opt/datadog/dd-java-agent.jar 'https://repository.sonatype.org/service/local/artifact/maven/redirect?r=central-proxy&g=com.datadoghq&a=dd-java-agent&v=LATEST'
```

We can then add the java agent property in `build.gradle`:

##### **`src/adservice/build.gradle`**
```java
task adService(type: CreateStartScripts) {
    mainClassName = 'hipstershop.AdService'
    applicationName = 'AdService'
    outputDir = new File(project.buildDir, 'tmp')
    classpath = startScripts.classpath
    defaultJvmOpts =
            ["-javaagent:/opt/datadog/dd-java-agent.jar"]
}
```

Finally, some environment variables needs to b added to the deployment for the java agent to work, and activate profiling, log injection, and trace analytics.

##### **`kubernetes-manifests/adservice.yaml`**
```yaml
apiVersion: apps/v1
kind: Deployment
...
labels:
    tags.datadoghq.com/env: "prod"
    tags.datadoghq.com/service: "adservice"
    tags.datadoghq.com/version: "1.0"
...
  spec:
    containers:
    - name: <CONTAINER_NAME>
      image: <CONTAINER_IMAGE>/<TAG>
      env:
        - name: DD_AGENT_HOST
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        - name: DD_ENV
          valueFrom:
            fieldRef:
              fieldPath: metadata.labels['tags.datadoghq.com/env']
        - name: DD_SERVICE
          valueFrom:
            fieldRef:
              fieldPath: metadata.labels['tags.datadoghq.com/service']
        - name: DD_VERSION
          valueFrom:
            fieldRef:
              fieldPath: metadata.labels['tags.datadoghq.com/version']
        
        - name: DD_PROFILING_ENABLED
          value: "true"
        - name: DD_LOGS_INJECTION
          value: "true"
        - name: DD_TRACE_ANALYTICS_ENABLED
          value: "true"
```