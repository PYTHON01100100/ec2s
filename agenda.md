# ec2s Development Agenda (Golang)

### 📌 Phase 1: Foundation & Project Setup
* **Initialize Go Module:** Initialize the repository using `go mod init github.com/yourusername/ec2s` and structure the project layout using standard Go CLI patterns (e.g., `/cmd`, `/pkg`, `/internal`).
* **CLI Framework Integration:** Integrate a robust CLI framework like [Cobra](https://github.com/spf13/cobra) (the engine behind `kubectl`) or a TUI library like [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm to give it a sleek, modern layout.
* **Repository Architecture:** Set up standard files including `README.md`, `LICENSE` (MIT), `.gitignore`, and the basic configuration files.

---

### ⚙️ Phase 2: AWS Authentication & Context Discovery
* **AWS SDK v2 Integration:** Install and configure the official [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2).
* **Credential & Profile Loading:** Write the logic to automatically detect and parse local AWS credential files (`~/.aws/credentials`), config files, `AWS_PROFILE` env vars, and SSO tokens.
* **Concurrent EC2 Discovery:** Use Go routines and Channels to fetch EC2 instance details across multiple regions concurrently, extracting data like Instance ID, Name tags, IP addresses, and Lifecycle states.

---

### 🚀 Phase 3: Core Control & Management Functions
* **Instance Lifecycle Subcommands:** Implement fast-acting Go functions linked to CLI commands for instance control:
  * `ec2s start` (Triggers `StartInstances` API)
  * `ec2s stop` (Triggers `StopInstances` API)
  * `ec2s reboot` (Triggers `RebootInstances` API)
* **High-Speed Filtering:** Build instant filtering algorithms to parse through running instances based on Security Groups, VPC IDs, or custom key-value tags.
* **Native SSM/SSH Session Tunneling:** Integrate direct secure shell shortcuts by invoking the AWS Systems Manager (SSM) StartSession API directly through the Go code, bypassing the need for open port 22.

---

### 📊 Phase 4: Observability & Advanced Tooling
* **Live Metrics Stream (CloudWatch):** Set up a ticker to query CloudWatch APIs every few seconds, feeding real-time CPU utilization and network metrics into your terminal view.
* **Attached Storage Mapping:** Map EBS volume mappings to their respective instances, allowing users to check storage thresholds or detached volume states.
* **Cost Allocation Insights:** Correlate instance type definitions with the AWS Pricing API to display real-time hourly burn rates (On-Demand vs Spot).

---

### 📦 Phase 5: Testing, Compiling & Distribution
* **Unit Testing with AWS Mocks:** Use the `smithy-go` or `mock` code generation tools to write isolated unit tests without making live HTTP requests to AWS endpoints.
* **Cross-Compilation:** Leverage Go's native cross-compilation compiler tags (`GOOS=darwin GOARCH=amd64 go build`) to output seamless, static binaries for macOS, Linux, and Windows.
* **Automated CI/CD Release Pipeline:** Set up a GitHub Action utilizing [GoReleaser](https://goreleaser.com/) to automatically bundle your binaries, generate changelogs, and push new formulas to Homebrew or Scoop upon a new Git tag push.
