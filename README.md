# octo
Octo, toolbelt for your DevOps team.


## Pre-requisites

- `~/.octo/config` config file for octo.

ex.

```
---

staging:
  api:
    name: 'API'
    region: "ap-southeast-1"
    asgNames:
      - "ASG-APIWorker-caterspot-prod"
      - "ASG-API-caterspot-prod"
  web:
    name: 'Web App'
    region: "ap-southeast-1"
    asgNames:

production:
  api:
    Name: 'API'
    region: "ap-northeast-1"
    asgNames:
      - "ASG-APIWorker-caterspot-prod"
      - "ASG-API-caterspot-prod"
```

## Usage

```
AWS_PROFILE=code3 octo ssh -s api -e production
```

