# Topwatcher

`Topwatcher` is a useful and handy golang code which dedicated to `Kubernetes` clusters as a `Cronjob`.

## How does it work?

In simple word `Topwatcher` checks out all pods ram usage and get the metrics of pods from `metric-server`, if ram usage from `metrics-server` is more than the threshold in `config` file, it will restart the `deployment` of that pod. Also it can send notification to `Slack` channel.

## Configuratio file

```bash
kubernetes:
  kubeconfig: /path/to/kube/config
  namespaces: default
  podrestart: false
  threshold:
    ram: 5
  exceptions:
    deployments:
      - deployment-name
      - deployment-name
      - deployment-name
slack:
  notify: false
  webhookurl: ""
  channel: ""
  username: ""

logging:
  debug: false
```

For now, there are three categories inside the config file as you see above. Also, you can use switches to override the config file values.

* `kubernetes` is about configs that related to cluster. 
* `slack` is about configs that related to slack notify.
* `logging` is about logging of the code.

In `kubernetes` directive you will see:

* `kubeconfig` which is the address of `kubeconfig` file.( `-k` or `--kubeconfig` )
* `namespace` is the target namespace that `Topwatcher` is going to check it.( `-n` or `--namespace` )
* `podrestart` is flag, so if you want to restart the deployment you can change it to `true`.( `-R` or `--restart-pod` )
* `threshold.ram` is the threshold of ram usage. It is obvious if a pod has ram usage more than this value is going to be restart. ( `-r` or `--ram-threshold` )
* `exceptions.deployment` is a list of exceptions for those deployments that you never want to be restarted. ( `-e` or `--exceptions` )

In `slack` directive you will see:

* `notify` is a flag, so if you want to get notify on slack, change it to true.
* `webhookurl` is url of slack.
* `channel` is channel name of your slack.
* `username`, the notify is going to be sent by this user.

In `logging` directive you will see:

* `debug` is a flag, so if you want to see debug logs, change it to true.( `-d` or `--debug` )

## How to use Topwatcher?

1. At first you need to build `topwatcher` docker image:

```bash
docker image build -t topwatcher:latest .
```

2. if you already have a `Kubernetes` cluster skip this step, otherwise run the command below to creane one using `kind`:

```bash
kind create cluster
```

3. Modify `config.yaml` by your own values then create a `topwatcher-configmap`:

```bash
kubectl create -f kubernetes/configmap.yml
```

4. Add cluster `kubeconfig` as a secret to your cluster:

```bash
cp /path/to/kubeconfig .
```

**NOTE THAT YOU HAVE TO CHANGE SERVER ADDRESS FROM 127.0.0.1 TO IP ADDRESS OF YOUR MASTER NODE**

then run the command below:

```bash
kubectl create secret generic cluster-config --from-file=./config
```

5. The last step is to create a cronjob:

**NOTE THAT YOU CAN ALSO CHANGE CRON-JOB.YML BY YOUR VALUES**

```bash
kubectl create -f kubernetes/cron-job.yml
```

