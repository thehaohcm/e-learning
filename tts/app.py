"""Tiny text-to-speech service for the API (not public).

GET /speak?text=apple&voice=en-US-AriaNeural&rate=-10%
    -> audio/wav, 16 kHz, mono, 16-bit (what the study device plays)

It uses Microsoft Edge's online voices through the edge-tts package. That
service is unofficial: when Microsoft changes it, rebuild this container to
get the newest edge-tts (bash deploy.sh --clean).
"""

import asyncio
import io
import logging
import os
import wave

import edge_tts
from aiohttp import web

DEFAULT_VOICE = os.environ.get("TTS_VOICE", "en-US-AriaNeural")
MAX_CHARS = 300
RATE_HZ = 16000

log = logging.getLogger("tts")


async def mp3_to_wav(mp3: bytes) -> bytes:
    """MP3 -> 16 kHz mono 16-bit WAV, without the silence at both ends."""
    trim = (
        "silenceremove=start_periods=1:start_threshold=-50dB,"
        "areverse,silenceremove=start_periods=1:start_threshold=-50dB,areverse"
    )
    proc = await asyncio.create_subprocess_exec(
        "ffmpeg", "-hide_banner", "-loglevel", "error", "-i", "pipe:0",
        "-af", trim, "-ac", "1", "-ar", str(RATE_HZ), "-f", "s16le", "pipe:1",
        stdin=asyncio.subprocess.PIPE, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE,
    )
    pcm, err = await proc.communicate(mp3)
    if proc.returncode != 0 or not pcm:
        raise RuntimeError("ffmpeg: " + err.decode(errors="replace")[:200])
    out = io.BytesIO()
    with wave.open(out, "wb") as w:
        w.setnchannels(1)
        w.setsampwidth(2)
        w.setframerate(RATE_HZ)
        w.writeframes(pcm)
    return out.getvalue()


async def speak(request: web.Request) -> web.Response:
    text = (request.query.get("text") or "").strip()
    if not text or len(text) > MAX_CHARS:
        return web.json_response({"error": f"text: 1..{MAX_CHARS} characters"}, status=400)
    voice = request.query.get("voice") or DEFAULT_VOICE
    rate = request.query.get("rate") or "+0%"
    mp3 = bytearray()
    try:
        com = edge_tts.Communicate(text, voice, rate=rate, connect_timeout=10, receive_timeout=20)
        async for chunk in com.stream():
            if chunk["type"] == "audio":
                mp3 += chunk["data"]
        if not mp3:
            raise RuntimeError("no audio")
        wav = await mp3_to_wav(bytes(mp3))
    except Exception as e:  # network error, voice not found, service changed...
        log.warning("speak %r failed: %s", text[:40], e)
        return web.json_response({"error": str(e)}, status=502)
    return web.Response(body=wav, content_type="audio/wav")


async def health(_: web.Request) -> web.Response:
    return web.json_response({"ok": True, "edge_tts": edge_tts.__version__})


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    app = web.Application()
    app.router.add_get("/speak", speak)
    app.router.add_get("/health", health)
    web.run_app(app, host="0.0.0.0", port=int(os.environ.get("PORT", "5050")), access_log=None)


if __name__ == "__main__":
    main()
