MongoDB
==================

To collect Mondo metrics and logs, we need to create a read-only user for the Datadog Agent in the admin database. To achieve it, we can run a shell script when the container starts.

##### **`kubernetes-manifests/mongo.yaml`**
```yaml
lifecycle:
    postStart:
        exec:
        command:
        - sh
        - -c
        # sleep is for Azure
        - "sleep 30; mongo --eval 'db = db.getSiblingDB(\"admin\"); db.createUser({ \"user\": \"datadog\", \"pwd\": \"ddpass\", \"roles\": [{role: \"read\", db: \"admin\"},{role: \"clusterMonitor\", db: \"admin\"},{role: \"read\", db: \"local\"}, {role: \"read\", db: \"store\"}]});'"
```

Finally, we need to annotate the pod for the Datadog agent to pick it up.

##### **`kubernetes-manifests/mongo.yaml`**
```yaml
annotations:
    ad.datadoghq.com/mongo.logs: '[{"source": "mongodb", "service": "mongo"}]'
    ad.datadoghq.com/mongo.check_names: '["mongo"]'
    ad.datadoghq.com/mongo.init_configs: '[{}]'
    ad.datadoghq.com/mongo.instances: |
        [
        {
            "server": "mongodb://datadog:ddpass@%%host%%:27017/admin",
            "additional_metrics":
            [
                "collection",
                "metrics.commands",
                "tcmalloc",
                "top"
            ]
        }
        ]
```