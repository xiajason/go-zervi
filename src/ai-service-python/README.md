Zervigo AI Service (Mock)
==========================

This lightweight Sanic application provides a mock AI service that is good enough for
local integration testing. It implements the endpoints required by the existing
`scripts/start-local-services.sh` flow and avoids external dependencies so the
service can run entirely offline.

## Endpoints

- `GET /health` — basic health check used by the startup script.
- `POST /api/v1/ai/chat` — accepts `{ "message": string }` and returns a canned
  assistant response along with simple tracing metadata.
- `POST /api/v1/ai/extract-keywords` — accepts `{ "text": string }` and returns
  up to five mock keywords. Useful for resume or job description previews.
- `POST /api/v1/ai/summary` — accepts `{ "text": string }` and returns a short
  deterministic summary to unblock downstream flows that expect AI output.

All endpoints respond immediately and never call external APIs. They are designed
to keep the integration surface identical to the planned production service so the
frontend can be wired up without surprises.

## Quick Start

```bash
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python ai_service_with_zervigo.py
```

By default the service listens on port `8100`. You can customise the port via the
`AI_SERVICE_PORT` environment variable.

## Why a Mock?

The real AI service is cost-sensitive and gated behind certification. Providing a
fully-fledged implementation inside the public deliverable is therefore not
possible. This mock gives the integration team something concrete to work with
while the certification process runs in parallel.

