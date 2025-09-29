# Stipes.tech Chatbot — System Context

## Overview
This document describes how the chatbot backend works — its architecture, request flow, and key design decisions. It is embedded as context so the AI can transparently explain or reason about the system.

---

## TL;DR
- HTTP endpoint `POST /chat` backed by OpenAI’s Responses API.
- Bundles Markdown files into the binary and includes them in system context.
- Runs either as a standalone Go binary (`go run ./cmd/server`) or as a Vercel serverless function (`/api/chat`).

---

## Architecture

```mermaid
flowchart TB
  A[Client: curl / Browser] -->|POST { message }| B[Go HTTP Server (net/http + mux)]
  subgraph S[Server Side]
    B --> C[Handlers: /healthz, /files, /chat]
    E[Embedded FS: pkg/embedded] --> C
    C --> D[OpenAI Client: pkg/openai]
  end
  D -->|SDK call| O[(OpenAI Responses API)]
  O -->|text| D --> C -->|JSON { output }| A

  %% Serverless path
  A -.->|POST /api/chat| V[Vercel Function: api/chat.go]
  V --> E
  V --> D
  V --> A