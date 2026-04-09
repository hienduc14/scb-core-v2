# SCB Core Feature Inventory

This document summarizes what this repository currently does so it can be refactored to Go without losing behavior.

## Purpose

`scb-core` is a controller service for SecureCodeBox-based security scans. It listens for scan requests from Kafka, validates and enriches them, creates Kubernetes `Scan` resources from YAML templates, monitors pod lifecycle, and publishes status/error events back to Kafka.

## High-Level Responsibilities

1. Consume scan requests from Kafka.
2. Validate request payloads and required fields.
3. Decrypt authentication tokens when needed.
4. Verify target repositories or container images before starting a scan.
5. Generate temporary SecureCodeBox scan manifests from templates.
6. Apply scan manifests to Kubernetes with `kubectl apply`.
7. Watch scan-related pods and translate container state into pipeline stages.
8. Publish status and error updates to Kafka.
9. Periodically clean up finished and stale failed scans from Kubernetes.

## Runtime Architecture

The current Python service starts three long-running background responsibilities:

- Kafka request consumer
- Kubernetes pod watcher and findings coordinator
- Scheduler for deleting completed and old failed scans

Main entrypoint: `scripts/main.py`

Core request handling: `scripts/controller.py`

## Request Intake

Incoming requests are consumed from Kafka topic `KAFKA_SCAN_REQUEST_TOPIC` with default `scb_request`.

Important request fields used by the service:

- `requestUUID`: unique scan identifier
- `projectUri`: repository URL, image reference, or target URL depending on scanner
- `targetUrl`: alternate target field used by DAST/nuclei flows
- `scanner`: requested scanner type
- `projectType`: `GITHUB`, `GITHUB_SERVER`, `GITLAB`, `GITLAB_SERVER`
- `authToken`: encrypted or plain auth payload
- `branch`
- `commitId` / `base_commit`
- `refId`
- `requestTime`
- `tag`: platform for trivy image scans
- `nucleiConfig`
- `requestParams`

The latest code also carries merge-request or pull-request context implicitly through:

- `commitId` mapped into internal `base_commit`
- `refId`

This is used for commit-based repository scanning rather than for a separate dedicated MR resource type.

## Supported Scan Types

The controller accepts both legacy and product-facing scanner names.

### Repository scanners

- `FS_CODE` and `FS_CODE_COMMIT` mapped to `semgrep`
- `FS_SECRET` and `FS_SECRET_COMMIT` mapped to `gitleaks`
- `FS_IAC` and `FS_IAC_COMMIT` mapped to `kics`

### Image scanner

- `FS_IMAGE` mapped to `trivy-image`

### DAST scanner

- `FS_DAST`
- `DAST_WEB`
- `DAST_NUCLEI`
- `NUCLEI`
- legacy `nuclei`

### Legacy direct names

- `semgrep`
- `gitleaks`
- `kics`
- `trivy-image`Dưới đây là bản prompt dạng Markdown (.md) để ông đưa cho Codex. Tôi giữ đúng tinh thần đoạn kiến trúc vừa viết + clean architecture + Kafka + controller-runtime.

Ông có thể copy nguyên file này làm CODEX_PROMPT.md.

⸻

SCB Core v2 – Architecture & Scaffold Prompt (Clean Architecture)

🎯 Objective

Build a Go codebase for scb-core-v2, a scan orchestration controller service.

The service:
	•	receives scan requests from Kafka (portal backend)
	•	orchestrates SecureCodeBox scans on Kubernetes
	•	observes runtime scan progress via Kubernetes (controller-runtime)
	•	translates runtime state into business lifecycle stages
	•	publishes status/error events back to Kafka for portal backend

This system is NOT a simple Kafka consumer.
It is a scan orchestration controller with event-driven status tracking.

⸻

🧱 Architecture Principles (Uncle Bob Clean Architecture)

Follow Clean Architecture strictly:

Layers
	•	Domain
	•	entities, value objects, enums
	•	pure business rules
	•	Use cases
	•	orchestration logic
	•	application-specific workflows
	•	Interface adapters
	•	controllers
	•	presenters
	•	mappers
	•	repository adapters
	•	Infrastructure
	•	Kafka
	•	Kubernetes
	•	config
	•	logging

Rules
	•	Dependencies point inward only
	•	Domain must not depend on infrastructure
	•	Use cases depend on interfaces (ports) only
	•	No business logic in Kafka or Kubernetes adapters
	•	Avoid “god service”
	•	Design for testability first

⸻

🔁 Core Communication Model

Key principle
	•	Kubernetes = detect state
	•	Status engine = decide state
	•	Kafka = notify external system

Flow

Portal backend
   ↓
Kafka request
   ↓
Kafka consumer
   ↓
ProcessScanRequest use case
   ↓
Submit Scan to Kubernetes
   ↓
controller-runtime watcher (Scan/Pod)
   ↓
internal observation
   ↓
status engine (evaluate + dedupe + order)
   ↓
Kafka publisher
   ↓
Portal backend updates state


⸻

🚫 Important Constraints

DO NOT
	•	publish Kafka directly from Kubernetes watcher
	•	embed business logic inside adapters
	•	implement lifecycle via polling loops
	•	use kubectl shell execution in core logic
	•	create a monolithic service file
	•	mix domain logic into transport layers

MUST
	•	separate detection vs decision vs notification
	•	use event-driven watcher (controller-runtime)
	•	enforce stage ordering and deduplication centrally
	•	keep domain and use cases framework-independent

⸻

📦 Project Structure

/cmd/controller/main.go

/internal/domain
  /entities
  /valueobjects
  /enums

/internal/usecase
  /services
  /ports
  /requests
  /responses

/internal/interfaceadapters
  /controllers
  /presenters
  /mappers

/internal/infrastructure
  /kafka
  /k8s
  /config
  /logging
  /scheduler

/internal/status
  evaluator.go
  engine.go


⸻

🧠 Domain Modeling

Define core concepts:
	•	ScanRequest
	•	ScanTarget
	•	ScannerType
	•	ScanStage
	•	ScanStatusEvent
	•	ScanError
	•	VerificationResult

Stage lifecycle (must preserve)

initializing → cloning → scanning → parsing → hooking → results

Rules
	•	stage progression must be monotonic
	•	terminal state must stop further transitions
	•	duplicate stage must be ignored

⸻

⚙️ Use Cases

Implement these use cases:
	•	ProcessScanRequest
	•	VerifyTarget
	•	DecryptAuthToken
	•	BuildScanManifest
	•	SubmitScan
	•	HandleK8sObservation
	•	PublishStatusEvent
	•	PublishErrorEvent
	•	CleanupScans

Important
	•	orchestration logic lives here
	•	use interfaces (ports) for all external dependencies

⸻

🔌 Ports (Interfaces)

Define interfaces for:

Kafka
	•	ScanRequestConsumer
	•	StatusEventPublisher
	•	ErrorEventPublisher

Kubernetes
	•	ScanSubmitter
	•	ScanObserver (from watcher)

Other
	•	AuthDecryptor
	•	TargetVerifier (repo/image/url)
	•	ManifestRenderer

⸻

☸️ Kubernetes Design (controller-runtime)

Use controller-runtime for event-driven observation.

Watch resources
	•	Scan CR
	•	Pods
	•	optionally Jobs

Responsibilities
	•	detect runtime changes
	•	filter relevant scan resources
	•	emit internal observation events

Non-responsibilities
	•	no Kafka publishing
	•	no business stage decision
	•	no deduplication logic

⸻

🧠 Status Engine Design

This is the most critical component.

Responsibilities
	•	translate observations → business stage
	•	enforce stage ordering
	•	suppress duplicates
	•	handle terminal state
	•	generate status/error events

Rules
	•	only forward transitions allowed
	•	repeated stage → ignore
	•	terminal → ignore future updates
	•	failure must include:
	•	stage
	•	error code
	•	message

⸻

🔄 Scanner Strategy

Avoid if/else spaghetti.

Implement strategy pattern:

Each scanner provides:
	•	target verification logic
	•	manifest input mapping
	•	auth requirements

Support scanners:
	•	semgrep
	•	gitleaks
	•	kics
	•	trivy-image
	•	nuclei

Support alias mapping:
	•	FS_CODE → semgrep
	•	FS_SECRET → gitleaks
	•	FS_IAC → kics
	•	FS_IMAGE → trivy
	•	FS_DAST / NUCLEI → nuclei

⸻

📡 Kafka Adapter

Responsibilities:
	•	consume scan requests
	•	deserialize payload
	•	invoke use case
	•	publish status/error events

Must NOT:
	•	contain lifecycle logic
	•	map pod states
	•	dedupe events

⸻

🔍 Observation Flow

controller-runtime watcher
   ↓
K8s event
   ↓
filter + correlate (requestUUID)
   ↓
emit internal observation
   ↓
status engine evaluate
   ↓
if new stage → publish Kafka
   ↓
else → ignore


⸻

🧪 Testing Requirements

Write tests for:
	•	scanner alias mapping
	•	stage ordering
	•	deduplication logic
	•	error code behavior
	•	ProcessScanRequest flow (mocked ports)

Example:

if ErrStatusPublish.IsTerminal() {
    t.Fatal("status publish errors should not be terminal")
}


⸻

⚙️ Bootstrap

Main should:
	•	load config
	•	init logger
	•	init Kafka adapter
	•	init K8s watcher (controller-runtime manager)
	•	init status engine
	•	wire use cases
	•	start all components
	•	support graceful shutdown

⸻

📌 Implementation Strategy

Follow this order:
	1.	Domain models
	2.	Ports (interfaces)
	3.	Use cases
	4.	Status engine
	5.	Scanner registry
	6.	Kafka adapter
	7.	Kubernetes adapter (controller-runtime)
	8.	Bootstrap
	9.	Tests
	10.	README

⸻

📖 README Requirements

Explain:
	•	architecture layers
	•	Kafka vs Kubernetes responsibilities
	•	status engine role
	•	why controller-runtime is used
	•	what is implemented vs TODO

⸻

🔥 Final Principle

Kubernetes detects
Status engine decides
Kafka notifies

Do not break this rule.

⸻

Nếu ông muốn bước tiếp theo, tôi có thể viết thêm:
	•	prompt phase 2: implement watcher + status engine thật
	•	hoặc skeleton code mẫu để Codex bám theo cho chuẩn luôn

## Scanner-Specific Behavior

### Semgrep

- Verifies repository existence and access through GitHub or GitLab APIs.
- Clones the repository in an init container.
- Downloads semgrep rules from S3-compatible object storage.
- Supports full scan and commit-based scan behavior.
- Commit-based mode is suitable for MR/PR-style scans by using `commitId` as the baseline commit.
- For root-commit edge cases, commit scan logic may degrade to full scan behavior.

### Gitleaks

- Verifies repository existence and access.
- Clones repository in an init container.
- Downloads `gitleaks.toml` from S3-compatible object storage.
- Supports full scan and commit-based scan behavior.
- Commit-based mode is suitable for MR/PR-style scans by using `commitId` as the baseline commit.
- Commit mode adjusts `--log-opts` and removes `--no-git`.

### KICS

- Verifies repository existence and access.
- Clones repository in an init container.
- Runs IaC scan through SecureCodeBox `kics` scan template.
- `FS_IAC_COMMIT` is accepted as an input type, but the current implementation does not apply commit-diff-specific behavior the way semgrep and gitleaks do.

### Trivy Image

- Verifies image existence by calling the registry manifest API.
- Supports both tag and digest image references.
- Uses auth from decrypted token.
- Retries insecure TLS mode when certificate verification fails.
- Downloads trivy secret rules from S3-compatible object storage.
- Passes image platform via request field `tag`.

### Nuclei

- Targets a URL instead of a Git repository.
- Accepts optional auth payload for static auth configuration.
- Requires an OpenAPI document from S3 or S3-like URL inputs.
- Normalizes Swagger/OpenAPI bucket and URI values from multiple request fields.
- Flattens `requestParams` into `-var key=value` arguments for nuclei.
- Injects auth YAML into an init container volume.
- Downloads the OpenAPI document before scan start.

## External Integrations

### Kafka

- Consumes scan requests.
- Publishes stage updates to status topic `KAFKA_SCAN_STATUS`, default `scb_status`.
- Publishes structured error events to the same status topic.
- Watches findings topic `scb_response_finding` to know when results are available.

### Kubernetes

- Runs inside cluster or falls back to local kubeconfig.
- Creates SecureCodeBox `Scan` resources using YAML templates.
- Lists and monitors pods in scanner namespace.
- Deletes finished and stale failed scans.

### SecureCodeBox

- Uses `execution.securecodebox.io/v1` `Scan` manifests.
- Relies on SecureCodeBox operator workflow for scan, parse, and hook pods.

### GitHub / GitLab APIs

- Verifies repository existence and access.
- Checks whether a baseline commit is a root commit for commit-scan logic.
- This commit check is part of MR/PR-style differential scan handling.

### Container Registries

- Verifies image manifests via registry HTTP API.
- Supports insecure retry on TLS verification failures.

### S3-Compatible Storage

- Downloads semgrep rules
- Downloads gitleaks rules
- Downloads trivy rules
- Downloads nuclei OpenAPI input

## MR/PR Scan Support

The latest codebase supports merge-request and pull-request style scanning as an extension of repository scanning.

Current behavior inferred from the request mapping and controller logic:

- Kafka requests can include `commitId`, which is mapped to internal `base_commit`.
- `FS_CODE_COMMIT` and `FS_SECRET_COMMIT` use that baseline commit to run differential scans.
- The controller checks whether the baseline commit is a root commit before choosing final semgrep behavior.
- `refId` is accepted and propagated through request handling, but it is not currently used to change manifest generation.
- There is no separate MR-specific Kubernetes resource type; MR scanning is implemented as a variant of existing repository scan flows.

Practical implication for a Go rewrite:

- Preserve commit-based scan behavior as a first-class feature.
- Model MR/PR scans as repository scans with review-context fields, not as a disconnected workflow.
- Keep the distinction between accepted scanner types and scanners that actually consume baseline commit data today.

## Status Lifecycle

The pod watcher converts Kubernetes pod/container state into ordered pipeline stages:

1. `initializing`
2. `cloning`
3. `scanning`
4. `parsing`
5. `hooking`
6. `results`

Behavior to preserve:

- Stage messages are emitted in order.
- Errors are emitted with structured codes and human-readable messages.
- Findings from `scb_response_finding` can advance the flow to `hooking` and `results`.
- Pod/container failures may include container log excerpts in the error payload.

## Error Handling

The repo defines stable application error codes in `scripts/utils/error_handler.py`.

Main categories:

- request validation
- configuration generation/application
- Kafka/Kubernetes infrastructure
- clone/scan/parse/hook execution
- repository lookup and generic failures

The Go version should preserve:

- machine-readable error codes
- stage-aware error reporting
- Kafka publication on failure paths

## Cleanup and Scheduling

Background cleanup jobs run on an interval:

- delete scans in `Done` state
- delete scans in `Error`/`Failed`-like states older than a retention threshold

Cleanup currently works by:

- listing SecureCodeBox `Scan` custom resources
- patching finalizers to `null`
- force deleting scans with `kubectl`

## Configuration Surface

Important environment variables used by the current implementation:

- `SECURECODEBOX_NAMESPACE_SCANNER`
- `SECURECODEBOX_NAMESPACE_OPERATOR`
- `KAFKA_ADDR`
- `KAFKA_USER`
- `KAFKA_PASSWORD`
- `KAFKA_SCAN_REQUEST_TOPIC`
- `KAFKA_SCAN_STATUS`
- `KAFKA_SCB_REQUEST_CONSUMER`
- `KAFKA_SCB_REQUEST_CONSUME`
- `ENCRYPTION-KEY`
- `WORKER_NODE`
- `MINUTES`
- `SECONDS`
- `ERROR_SCAN_RETENTION_DAYS`
- `ERROR_SCAN_CHECK_DAYS`
- `ERROR_SCAN_CHECK_MINUTES`
- `ERROR_SCAN_CHECK_SECONDS`
- `FINDING_PROCESS_WAIT_TIME_SEC`
- `FINDING_PROCESS_MAX_THREAD`
- `FINDING_PROCESS_MIN_QUEUE`
- `SCB_HTTP_POOL_CONNECTIONS`
- `SCB_HTTP_POOL_MAXSIZE`

Also note:

- dotenv file path is hardcoded to `/vault/secrets/config.env`
- scanner templates are read from `/app/configs/*.yaml`
- temporary rendered manifests are written to `/tmp`

## Deployment Characteristics

- Runs as a containerized long-lived service.
- Helm chart deploys `2` replicas.
- Uses `kubectl` inside the container to apply manifests and delete scans.
- Requires cluster access, Kafka access, registry/network access, and object storage access.

## Suggested Go Module Split

A practical Go rewrite can be split into these packages:

- `cmd/controller`: service bootstrap
- `internal/config`: env and runtime config
- `internal/kafka`: producers, consumers, message contracts
- `internal/request`: payload validation and normalization
- `internal/auth`: token decryption and auth parsing
- `internal/verify`: repo/image verification clients
- `internal/scanners`: scanner-specific request translation
- `internal/templates`: manifest rendering
- `internal/k8s`: apply/watch/delete logic
- `internal/status`: stage ordering and status emission
- `internal/cleanup`: scheduled retention jobs
- `internal/errors`: error code catalog and serializers

## Behavior Worth Preserving Exactly

- Kafka topic contracts and payload shape
- accepted scanner names and mapping rules
- commit-based MR/PR scan behavior for `FS_CODE_COMMIT` and `FS_SECRET_COMMIT`
- stage names and ordering
- nuclei request parameter flattening
- repo verification before manifest creation
- image verification before trivy scan creation
- auth token decryption flow
- cleanup semantics for done and stale failed scans

## Areas That Can Be Cleaned Up During Refactor

- Replace shelling out to `kubectl` with Kubernetes API calls where possible.
- Centralize scanner definitions instead of branching in one large controller function.
- Normalize environment variable names and remove duplicate variants.
- Replace ad hoc YAML string replacement with typed manifest generation.
- Make topic names, status payloads, and request schemas explicit Go structs.
- Remove hidden filesystem assumptions like `/app` and `/vault/secrets/config.env` where possible.
- Separate transport concerns from business logic so scanner flows are testable.

## Most Important Outcome for the Go Rewrite

If rewritten in Go, this service should still behave as a scan orchestration controller, not just a Kafka consumer. The critical product behavior is:

- accept scan requests
- validate and verify targets
- materialize the correct SecureCodeBox scan
- publish lifecycle status updates
- surface failures consistently
- clean up old scan resources
