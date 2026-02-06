# Sidecar Plugins vs. Microservices

When extending ZenoEngine, you have two primary architectural choices for externalizing logic: **Native Sidecar Plugins** and **Microservices**. While both allow you to use other languages (Go, Python, .NET, etc.), they differ significantly in coupling, performance, and operational complexity.

## Comparison at a Glance

| Feature | Sidecar Plugin | Microservice |
| :--- | :--- | :--- |
| **Communication** | Stdin/Stdout Pipe (JSON-RPC) | HTTP / gRPC (Network) |
| **Latency** | **Low** (< 1ms) | **Medium/High** (Network overhead) |
| **Lifecycle** | Managed by ZenoEngine (Auto-start/restart) | Independent (External Orchestrator/K8s) |
| **Coupling** | **Tight** (1:1 with Engine Instance) | **Loose** (N:M, independently scalable) |
| **Deployment** | Bundled with the App (Folder) | Separate Containers / Servers |
| **State** | Shared Scope (via Zeno Context) | Stateless / Explicit State Transfer |
| **Complexity** | Low (File-based) | High (Service Discovery, LB, Network) |

## 1. Native Sidecar Plugins
A **Sidecar Plugin** is a child process spawned and managed directly by the ZenoEngine process.

### Architecture
`ZenoEngine (Host) <==[ Pipe ]==> Plugin Process (Guest)`

### Key Characteristics
*   **Managed Lifecycle**: ZenoEngine starts the plugin when the engine starts (or on demand) and kills it when the engine stops. If the plugin crashes, ZenoEngine automatically restarts it.
*   **Deployment Simplicity**: You simply place the binary in the `plugins/` directory. No Dockerfiles, no Kubernetes services, no ports to manage.
*   **Performance**: Communication happens over standard I/O pipes, avoiding TCP/IP stack overhead. This makes it suitable for high-frequency calls (e.g., text processing, crypto).
*   **Security**: The plugin runs with the same user permissions as the engine (unless configured otherwise), but is restricted to the input provided via JSON-RPC.

### When to use Sidecar?
*   **Helper Libraries**: You need a specific library available in Python or C# (e.g., PDF generation, Image Processing) tightly integrated into your Zeno workflow.
*   **High Performance / Low Latency**: You need to call a function thousands of times per second (e.g., parsing log lines).
*   **Embedded Deployments**: You are running ZenoEngine on a single server or edge device (IoT) where running a full container orchestration platform is overkill.

## 2. Microservices
A **Microservice** is a standalone service running on a network port, accessible via HTTP or gRPC.

### Architecture
`ZenoEngine (Client) --[ Network ]--> Load Balancer --> Service (Server)`

### Key Characteristics
*   **Independent Scaling**: You can have 1 ZenoEngine instance calling a cluster of 50 PDF Generator services. The load is distributed.
*   **Decoupling**: The service can be updated, restarted, or moved without affecting the ZenoEngine process directly.
*   **Technology Agnostic**: As long as it speaks HTTP/JSON, it doesn't matter how it's built. ZenoEngine interacts with it using standard `http.post` or `grpc.call` slots.

### When to use Microservices?
*   **Heavy Compute Scaling**: If generating a report takes 100% CPU for 10 seconds, a Sidecar would starve the host machine resources. A microservice can run on a dedicated heavy-compute cluster.
*   **Shared Logic**: The logic is used by ZenoEngine AND other applications (e.g., a Mobile App backend).
*   **Team Boundaries**: A separate team maintains the "Payment Service" and provides an API contract.

## Summary

*   **Choose Sidecar** if you want to **"extend the capabilities"** of the ZenoEngine runtime itself (like adding a new standard library function) without operational overhead.
*   **Choose Microservices** if you want to **"offload work"** to a scalable, independent system.
