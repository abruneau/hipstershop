The first step in instrumenting a Kubernetes cluster with Datadog is to deploy the agent.

Multiple options to deploy the agent are available: Helm, Deamonset, Operator. We can find step by step instruction for each methods in the [documentation](https://docs.datadoghq.com/fr/agent/kubernetes/?tab=helm#pagetitle)

In our case, we will leverage Helm

# Helm values

In order for the Datadog agent to collect data, we need to activate functionalities and configure them. To do so, we can pass to the Helm deployment a Yaml file with values.
The easiest way to start is to download the template from Datadog website. We will store it in a `helm` directory

```sh
mkdir helm
wget -O ./helm/values.yaml https://raw.githubusercontent.com/DataDog/helm-charts/master/charts/datadog/values.yaml
```

## API and APP keys

You will need to provide an API key and an APP key to enable communication between the agent and Datadog platform. To prevent them from leaking in git repositories, it is not recommended to put them in `values.yaml`. A solution is to create a Kubernetes secret and reference it in `values.yaml`.

```sh
DATADOG_SECRET_NAME=datadog-secret
kubectl create secret generic $DATADOG_SECRET_NAME --from-literal api-key="<DATADOG_API_KEY>" --namespace="default"
```

The secret can also be put in a manifest and excluded of the git repo. This has the benefit of managing it with Skaffold

##### **`secrets.yaml`**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: datadog-secret
type: generic
stringData:
  api-key: <DATADOG_API_KEY>
  app-key: <DATADOG_APP_KEY>
```

##### **`values.yaml`**
```yaml
datadog:
  apiKey: # <DATADOG_API_KEY>
  apiKeyExistingSecret: datadog-secret
  appKey:  # <DATADOG_APP_KEY>
  appKeyExistingSecret: datadog-secret
```

## Logs

To collect logs, you will need to set `logs.enabled` to `true`. If you want to monitor your entire cluster, you can leave `logs.containerCollectAll` to true. But if you work pods after pods, you can set it to `false` to only collect logs from the containers you enabled. I also recommend to read logs from files, as it seems to be more stable across k8s distributions: `logs.containerCollectUsingFiles: true`

## APM

To collect traces, you will need to set `apm.enabled` to `true`.

# Apply Helm chart

You can deploy the standard way with

```sh
helm install -f ./helm/values.yaml datadog datadog/datadog
```

Or you can integrate it to your Skaffold file. This second option has the benefit to clean Datadog agent from your cluster with the rest of the exercise.

```yaml
deploy:
  kubectl:
    manifests:
    - ./kubernetes-manifests/*.yaml
  helm:
    releases:
      - name: datadog-agent
        chartPath: datadog/datadog
        valuesFiles: 
        - helm/values.yaml
        skipBuildDependencies: true
        remote: true
```