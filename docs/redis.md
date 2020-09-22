Redis
==========

To collect Redis metrics and logs, all we need to do is annotate the pod for the agent to pick it up:


##### **`kubernetes-manifests/redis.yaml`**
```yaml
annotations:
    ad.datadoghq.com/redis.logs: '[{"source":"redis","service":"redis"}]'
    ad.datadoghq.com/redis.check_names: '["redisdb"]'
    ad.datadoghq.com/redis.init_configs: '[{}]'
    ad.datadoghq.com/redis.instances: |
        [
        {
            "host": "%%host%%",
            "port":"6379"
        }
        ]
```