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

Each profile's IAM identity needs `ec2:DescribeInstances` to list instances,
and, if you use `s`/`S`/`D`, `ec2:StartInstances` / `ec2:StopInstances` /
`ec2:TerminateInstances` too.

A bad or expired profile (e.g. an expired SSO session) doesn't stop the
rest of your accounts from loading — it shows up as a warning in the
footer instead, along with exactly where `ec2s` looked for AWS config
(resolved home directory, and whether `config`/`credentials` were found
there) so a bad credential lookup is self-diagnosable from the UI.

### Troubleshooting: "it's not reading my keys"

`ec2s` resolves `~/.aws/config` and `~/.aws/credentials` using the *running
binary's* OS-native home directory — this matters if you use WSL. A
Windows-built `ec2s.exe` reads `%USERPROFILE%\.aws` even when launched from
inside a WSL/Linux shell, which is a completely separate filesystem from
that shell's own `$HOME/.aws`. If you ran `aws configure` inside WSL but
you're running a Windows binary (or vice versa), `ec2s` will pick up
whatever profile happens to exist at *that OS's* home directory — which may
be missing, empty, or stale, and AWS will reject the request (commonly as
an HTTP 403).

Fixes, in order of preference:

- Run a binary built for the same OS as the shell you're using (`go build`
  inside WSL for a WSL shell, on Windows for PowerShell/cmd).
- Or point `ec2s` explicitly at the right files with `$AWS_CONFIG_FILE` /
  `$AWS_SHARED_CREDENTIALS_FILE` (these are honored automatically — no
  `ec2s`-specific flag needed).
- Or check the footer's warning line, which prints the exact `home=`,
  `config=`, and `credentials=` paths `ec2s` resolved and whether each file
  was found, so you can confirm at a glance whether it's even looking in
  the right place.

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
| `s`                 | start the selected instance              |
| `S`                 | stop the selected instance (asks to confirm) |
| `D`                 | terminate the selected instance (asks to confirm, irreversible) |
| `?`                 | help                                     |
| `q` / `Ctrl-C`      | quit                                     |

`s`/`S`/`D` act on whichever instance is currently selected, in its own
account/region, and the footer shows the result. `s` (start) runs
immediately since it's non-destructive; `S` (stop) and `D` (terminate) ask
for confirmation first — `D` warns that termination can't be undone.

`q` quits from anywhere in the app — the main table, the account/help
screens, an open confirmation — except while typing into the filter box,
where `q` is just a character to search for. `Ctrl-C` always quits,
including from the filter box.

Filter syntax supports `state:running`, `account:prod`, `region:us-east-1`,
`type:t3.micro`, `os:windows`, `zone:us-east-1a`, `vpc:vpc-…`,
`subnet:subnet-…`, or plain substring matching against instance name/ID.

### What it shows

The table and the info panel (top of the screen, updates as you move the
selection) surface the fields that matter most when you're trying to find
or reach a specific instance: **name**, instance ID, state, type, **OS**,
**account** and **region**, **availability zone**, **VPC ID** and
**subnet ID**, and both the **public (external)** and **private (internal)**
IP addresses.

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
