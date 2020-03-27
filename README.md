# octo
Octo, toolbelt for your DevOps team.

[![asciicast](https://asciinema.org/a/HKlKZ1SfdX4IcS25RGMEuT1R5.svg)](https://asciinema.org/a/HKlKZ1SfdX4IcS25RGMEuT1R5)


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

## To Install

```
brew tap faizalzakaria/homebrew-tap
brew install octo
```

## Usage

```
AWS_PROFILE=code3 octo ssh -s api -e production
```

---

## To build

```sh
go build
```

## To release

Create a tag

```sh
git tag -a v1.0.0 -m"Hello""
```

Then release

```sh
GITHUB_TOKEN=<Token> goreleaser release
```
