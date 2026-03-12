# AGENTS.md

## Project: ollamon

`ollamon` is a terminal-based operational monitor for Ollama nodes.

The project provides a **TUI (Terminal User Interface)** similar to `htop`, designed to give operators real-time visibility into:

* Ollama model runtime state
* Model memory residency
* Host system resource usage
* Operational insights and diagnostics

The tool is intended to work on:

* macOS (Apple Silicon / Mac Studio environments)
* Linux servers (GPU or CPU inference nodes)

---

# Core Mission

The goal of `ollamon` is to provide:

**A lightweight operational console for Ollama inference nodes**

It should help operators answer questions like:

* What models are installed?
* What models are currently loaded in memory?
* How much VRAM/RAM are they using?
* Is the system under memory pressure?
* Which node is overloaded?
* Are models thrashing in memory?

---

# Architecture Overview

```
ollamon
│
├── cmd/ollamon
│   └── main.go            → CLI entrypoint
│
├── internal
│   ├── app                → TUI model and state management
│   │   ├── model.go
│   │   └── sample.go
│   │
│   ├── config             → runtime configuration
│   │   └── config.go
│   │
│   ├── ollama             → Ollama API client
│   │   └── client.go
│   │
│   ├── system             → host system metrics
│   │   ├── stats.go
│   │   ├── gpu_darwin.go
│   │   └── gpu_linux.go
│   │
│   ├── tui                → visual styles
│   │   └── styles.go
│   │
│   └── util               → shared formatting utilities
│       └── format.go
```

---

# Data Sources

ollamon gathers metrics from several sources.

### 1. Ollama API

Primary API endpoints:

```
GET /api/tags
GET /api/ps
```

Used to determine:

* installed models
* running models
* VRAM usage
* model expiration
* parameter size
* quantization

---

### 2. Host System Metrics

Collected via:

```
gopsutil
```

Metrics include:

* CPU usage
* memory usage
* disk usage
* load average
* uptime

---

### 3. GPU Telemetry

Platform-specific implementations:

macOS

```
system_profiler SPDisplaysDataType
```

Linux

```
nvidia-smi
```

Future support:

* AMD ROCm
* Intel GPU
* Apple Metal telemetry

---

# Terminal UI Design

The interface is divided into **three panels**.

## Overview Panel

Displays host and Ollama node health.

Example information:

* hostname
* OS and platform
* uptime
* CPU usage
* RAM usage
* disk usage
* GPU summary
* installed model count
* running model count
* operational insights

---

## Running Models Panel

Displays models currently loaded in memory.

Fields:

* model name
* parameter size
* quantization level
* VRAM usage
* expiration timer

This helps identify:

* memory residency
* model churn
* VRAM pressure

---

## Installed Models Panel

Displays models present in the Ollama cache.

Fields:

* model name
* model size
* parameter count
* quantization
* last modification time

Also displays:

```
Total model cache size
```

---

# Operational Insights Engine

The TUI includes a lightweight analysis layer.

It generates hints such as:

* RAM pressure warnings
* disk capacity alerts
* large model identification
* memory residency hints

Example output:

```
• RAM usage is high, large models may load slowly
• Disk usage approaching critical levels
• Largest model: llama3:70b
```

This system is intentionally simple and rule-based.

Future versions may introduce:

* heuristic scoring
* anomaly detection
* ML-driven insights

---

# Configuration

Configuration is environment variable based.

```
OLLAMA_HOST
OLLAMON_INTERVAL_MS
OLLAMON_DISK_PATH
OLLAMON_COMPACT
```

Example:

```
OLLAMA_HOST=http://10.0.0.12:11434
OLLAMON_INTERVAL_MS=1000
```

---

# Release Engineering

The project uses:

```
Makefile
goreleaser
GitHub Actions
Dockerfile
```

Supported build targets:

* linux amd64
* linux arm64
* darwin amd64
* darwin arm64

Binary version is embedded using:

```
-ldflags "-X main.version"
```

---

# Agent Development Guidelines

AI agents modifying this repository must follow these principles.

### 1. Preserve Internal Architecture

All business logic must stay in:

```
internal/
```

The CLI entrypoint must remain minimal.

---

### 2. Avoid External Runtime Dependencies

`ollamon` must remain lightweight.

Avoid adding:

* large frameworks
* runtime daemons
* background services

---

### 3. Cross-Platform Compatibility

All features must support:

```
macOS
Linux
```

Platform-specific code must use:

```
build tags
```

---

### 4. Non-blocking UI

The TUI must remain responsive.

* long operations must run in commands
* avoid blocking in the render loop

---

### 5. Minimal Network Assumptions

The Ollama server may be:

* local
* remote
* containerized
* behind reverse proxy

Agents must avoid assumptions about deployment topology.

---

# Planned Features

## Observability

Future telemetry improvements:

* token/s throughput
* first-token latency
* request queue depth
* inference duration
* context size visibility

---

## GPU Monitoring

Planned GPU telemetry improvements:

macOS

* Metal performance counters
* Apple GPU utilization
* VRAM allocation

Linux

* NVIDIA GPU telemetry
* ROCm support
* Intel GPU metrics

---

## Multi Node Mode

A cluster view of multiple Ollama nodes.

Example:

```
node-1  Mac Studio M3
node-2  Linux L40S
node-3  Linux CPU inference
```

The UI would allow switching between nodes.

---

## Log Panel

Future TUI panel:

```
Ollama logs
```

Sources:

```
~/.ollama/logs/server.log
journalctl -u ollama
```

---

## Prometheus Exporter

A sidecar mode that exposes metrics for:

```
Prometheus
Grafana
```

---

# Security Model

ollamon is **read-only**.

It does NOT:

* modify models
* delete models
* start or stop Ollama

It only reads metrics and displays them.

---

# Contribution Guidelines

Contributors should:

* keep the binary small
* maintain cross-platform compatibility
* prefer standard Go libraries
* document new features

---

# Long Term Vision

`ollamon` should evolve into:

```
The operational console for local LLM infrastructure
```

Not just for Ollama but potentially:

* llama.cpp
* vLLM
* LM Studio
* OpenAI-compatible endpoints

The architecture should remain flexible for this evolution.

---
