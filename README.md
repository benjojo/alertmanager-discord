# alertmanager-discord

This is a webserver that accepts webhooks from AlertManager. It will post your Prometheus alert notifications into a Discord channel as they trigger:

![](/.github/demo-new.png)

## Warning

This program is not a replacement to alertmanager, it accepts webhooks from alertmanager, not Prometheus.

The standard "dataflow" should be:

```
Prometheus -------------> alertmanager -------------------> alertmanager-discord

alerting:                 receivers:
  alertmanagers:          - name: 'discord_webhook'         environment:
  - static_configs:         webhook_configs:                   - DISCORD_WEBHOOK=https://discordapp.com/api/we...
    - targets:              - url: 'http://localhost:9094'
       - 127.0.0.1:9093





```

## Example alertmanager config

```
global:
  # The smarthost and SMTP sender used for mail notifications.
  smtp_smarthost: 'localhost:25'
  smtp_from: 'alertmanager@example.org'
  smtp_auth_username: 'alertmanager'
  smtp_auth_password: 'password'

# The directory from which notification templates are read.
templates:
- '/etc/alertmanager/template/*.tmpl'

# The root route on which each incoming alert enters.
route:
  group_by: ['alertname']
  group_wait: 20s
  group_interval: 5m
  repeat_interval: 3h
  receiver: discord_webhook

receivers:
- name: 'discord_webhook'
  webhook_configs:
  - url: 'http://localhost:9094'
```

## Deployment

### Docker

If you wish to deploy this to docker infra, you can find the docker hub repo here: https://hub.docker.com/r/speckle/alertmanager-discord/

### Kubernetes Helm Chart

If you wish to deploy this to Kubernetes, this repository contains a Helm Chart.

```shell
helm upgrade --install \
--create-namespace \
--namespace alertmanager-discord
alertmanager-discord \
./deploy/helm
```

You can optionally also provide a values yaml file, `--values ./your-values.yaml`, to override the default values.

## Acknowledgements

This repository is forked from https://github.com/benjojo/alertmanager-discord under the Apache 2.0 license
