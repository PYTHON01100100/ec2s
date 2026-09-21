# ec2s

Easily manage AWS EC2 resources in the terminal (~g1c for EC2) 🚀

`ec2s` is a terminal UI for browsing and (eventually) managing AWS EC2
instances, in the spirit of [k9s](https://github.com/derailed/k9s)
(Kubernetes), [e1s](https://github.com/keidarcy/e1s) (ECS), and
[g1c](https://github.com/nlamirault/g1c) (GCP VMs).

Unlike e1s, which shows one AWS profile/context at a time, `ec2s`'s
headline feature is an **aggregated multi-account view**: you define
several AWS accounts (profiles) and regions once, and `ec2s` fetches EC2
instances from all of them concurrently into a single table, each row
tagged with the account and region it came from. The idea is borrowed
from AWS CDK's `Environment{Account, Region}` concept — just applied to
runtime resource discovery instead of infrastructure deployment.

## Install

```sh
go install github.com/PYTHON01100100/ec2s/cmd/ec2s@latest
```

## Configuration

`ec2s` looks for a config file in this order:

1. `--config <path>`
2. `$EC2S_CONFIG`
3. `./ec2s.yaml`
4. `~/.config/ec2s/config.yaml` (`%AppData%\ec2s\config.yaml` on Windows)

See [ec2s.example.yaml](ec2s.example.yaml) for the schema:

```yaml
accounts:
  - name: prod
    profile: prod-profile
    regions: [us-east-1, eu-west-1]
  - name: staging
    profile: staging-profile
    region: us-east-1 # shorthand for a single region
```

Each `profile` is an AWS CLI profile from `~/.aws/config` /
`~/.aws/credentials`. Static credentials, assume-role (`role_arn` /
`source_profile`), `credential_process`, and AWS SSO profiles all work
out of the box — `ec2s` uses the AWS SDK for Go v2's standard shared-config
credential chain, it doesn't parse credentials itself.

**Zero config:** if no config file is found, `ec2s` falls back to
`$AWS_PROFILE` (if set), otherwise every profile it finds in your local AWS
config/credentials files, otherwise a single `default` account using the
SDK's default credential chain (environment variables, instance metadata,
container credentials, ...). Region resolution per account falls back
through `$AWS_REGION` → `$AWS_DEFAULT_REGION` → that profile's `region`
setting in `~/.aws/config` → `us-east-1`.

A bad or expired profile (e.g. an expired SSO session) doesn't stop the
rest of your accounts from loading — it shows up as a warning in the
footer instead.

## Usage

```sh
ec2s                      # uses config search order above
ec2s --config my.yaml     # explicit config file
```

### Keybindings

| Key                | Action                                   |
| ------------------- | ---------------------------------------- |
| `j` / `k` / arrows  | move down / up                           |
| `g` / `G`           | jump to top / bottom                     |
| `/`                 | filter (plain text, or `column:value`)   |
| `Esc`               | clear filter / close overlay             |
| `Ctrl-A`            | filter by configured account             |
| `Ctrl-R`            | refresh (re-fetch all accounts/regions)  |
| `?`                 | help                                     |
| `q` / `Ctrl-C`      | quit                                     |

Filter syntax supports `state:running`, `account:prod`, `region:us-east-1`,
`type:t3.micro`, `zone:us-east-1a`, `vpc:vpc-…`, `subnet:subnet-…`, or plain
substring matching against instance name/ID.

### What it shows

The table and the info panel (top of the screen, updates as you move the
selection) surface the fields that matter most when you're trying to find
or reach a specific instance: **name**, instance ID, state, type, **account**
and **region**, **availability zone**, **VPC ID** and **subnet ID**, and
both the **public (external)** and **private (internal)** IP addresses.

## Inspiration

`ec2s` stands on the shoulders of similar keyboard-driven terminal UIs for
cloud/infra resources:

- [k9s](https://github.com/derailed/k9s) — the original: a terminal UI for Kubernetes clusters.
- [e1s](https://github.com/keidarcy/e1s) — a terminal UI for AWS ECS, which `ec2s`'s table/filter/help UX is closely modeled on.
- [g1c](https://github.com/nlamirault/g1c) — a terminal UI for Google Cloud VM instances.
- [a1s](https://github.com/PYTHON01100100/a1s) — a terminal UI for Alibaba Cloud ECS instances, another project by the same author.

## Roadmap

See [agenda.md](agenda.md) for the full phased roadmap. Implemented so
far: project foundation (Phase 1) and AWS auth + concurrent multi-account
discovery (Phase 2). Instance actions (start/stop/reboot/SSM), CloudWatch
metrics, and distribution tooling are tracked as Phases 3-5.

## License

MIT — see [LICENSE](LICENSE).
