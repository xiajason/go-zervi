"""Lightweight Sanic-based mock AI service used during local integration.

The real AI service is gated behind certification. This module provides a
drop-in stand-in that mimics the public contract required by the frontend and
gateway so the rest of the stack can be verified end-to-end without incurring
external costs.
"""

from __future__ import annotations

import os
import textwrap
import uuid
from datetime import datetime
from typing import Dict, List

from sanic import Sanic
from sanic.log import logger
from sanic.request import Request
from sanic.response import HTTPResponse, json


def _safe_get(data: Dict[str, str], key: str) -> str:
    value = data.get(key)
    if value is None:
        return ""
    return str(value).strip()


def _current_timestamp() -> str:
    return datetime.utcnow().isoformat(timespec="seconds") + "Z"


def _build_trace() -> Dict[str, str]:
    return {
        "requestId": str(uuid.uuid4()),
        "generatedAt": _current_timestamp(),
    }


app = Sanic("ZervigoAIService")


@app.middleware("request")
async def add_cors_headers(request: Request) -> None:  # noqa: D401 - Sanic middleware signature
    """Inject permissive CORS headers for local testing."""
    request.ctx.cors_headers = {
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Methods": "GET,POST,OPTIONS",
        "Access-Control-Allow-Headers": "Content-Type,Authorization",
    }


@app.middleware("response")
async def apply_cors_headers(request: Request, response: HTTPResponse) -> None:  # noqa: D401
    """Attach the CORS headers captured during the request middleware."""
    for key, value in getattr(request.ctx, "cors_headers", {}).items():
        response.headers[key] = value


@app.get("/health")
async def health(_: Request) -> HTTPResponse:
    """Simple health endpoint used by the startup script."""
    return json(
        {
            "service": "zervigo-ai-mock",
            "status": "healthy",
            "timestamp": _current_timestamp(),
        }
    )


@app.post("/api/v1/ai/chat")
async def chat(request: Request) -> HTTPResponse:
    payload = request.json or {}
    message = _safe_get(payload, "message")

    if not message:
        reply = (
            "Hello from the Zervigo AI mock service. Provide a 'message' field to"
            " receive a sample response."
        )
    else:
        reply = textwrap.dedent(
            f"""
            Acknowledged your message: "{message}".
            Because this is the offline mock, the reply is deterministic so that
            the integration flow can be verified without the production AI stack.
            """
        ).strip()

    response = {
        "trace": _build_trace(),
        "data": {
            "response": reply,
            "tokensUsed": {
                "prompt": max(len(message) // 4, 1),
                "completion": 32,
            },
        },
    }
    return json(response)


def _extract_keywords(text: str) -> List[str]:
    words = [w.strip(".,!?:;") for w in text.split() if len(w) > 3]
    unique_words: List[str] = []
    for word in words:
        clean = word.lower()
        if clean not in unique_words:
            unique_words.append(clean)
        if len(unique_words) == 5:
            break
    if not unique_words:
        return ["zervigo", "ai", "mock"]
    return unique_words


@app.post("/api/v1/ai/extract-keywords")
async def extract_keywords(request: Request) -> HTTPResponse:
    payload = request.json or {}
    text = _safe_get(payload, "text")
    keywords = _extract_keywords(text)
    return json({
        "trace": _build_trace(),
        "data": {
            "keywords": keywords,
            "sourceLength": len(text),
        },
    })


@app.post("/api/v1/ai/summary")
async def summary(request: Request) -> HTTPResponse:
    payload = request.json or {}
    text = _safe_get(payload, "text")
    if not text:
        summary_text = "No source text provided; returning placeholder summary."
    else:
        snippet = text.replace("\n", " ").strip()
        if len(snippet) > 160:
            snippet = snippet[:157] + "..."
        summary_text = f"Summary: {snippet}"

    return json({
        "trace": _build_trace(),
        "data": {
            "summary": summary_text,
        },
    })


def main() -> None:
    port = int(os.environ.get("AI_SERVICE_PORT", "8100"))
    host = os.environ.get("AI_SERVICE_HOST", "0.0.0.0")
    logger.info("Starting Zervigo AI mock service on %s:%d", host, port)
    app.run(host=host, port=port, access_log=False, auto_reload=False, single_process=True)


if __name__ == "__main__":
    main()

