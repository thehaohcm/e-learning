"""Biến pmp_podcast_script.txt thành các file mp3 (mỗi phần một file + một file ghép).

Cách dùng (chạy trên máy của bạn, cần internet):
    pip install edge-tts
    python make_podcast.py                          # podcast dài
    python make_podcast.py pmp_drill_script.txt     # bài luyện ngắn

Giọng đọc: Microsoft Edge neural voices (miễn phí). Hai người dẫn:
  A = giọng nữ, B = giọng nam. Giọng "Multilingual" đọc được cả tiếng Việt lẫn
  tiếng Anh nên các thuật ngữ và câu hỏi tiếng Anh nghe tự nhiên hơn.
Nếu giọng multilingual không chạy, đổi sang VOICES_VI bên dưới.
"""
import asyncio
import re
import sys
from pathlib import Path

import edge_tts

HERE = Path(__file__).parent
ARGS = [a for a in sys.argv[1:] if not a.startswith("--")]
SCRIPT = HERE / (ARGS[0] if ARGS else "pmp_podcast_script.txt")  # vd: python make_podcast.py pmp_drill_script.txt
OUT = HERE / ("mp3_" + SCRIPT.stem)

VOICES = {"A": "en-US-AvaMultilingualNeural", "B": "en-US-AndrewMultilingualNeural"}
VOICES_VI = {"A": "vi-VN-HoaiMyNeural", "B": "vi-VN-NamMinhNeural"}  # dự phòng
RATE = "-5%"  # chậm hơn một chút để dễ nghe; đổi "+0%" nếu muốn bình thường

if "--vi" in sys.argv:
    VOICES = VOICES_VI


def parse(text: str):
    parts, cur = [], None
    for raw in text.splitlines():
        line = raw.strip()
        if not line:
            continue
        if line.startswith("#"):
            cur = {"title": line.lstrip("# ").strip(), "lines": []}
            parts.append(cur)
        elif cur is not None:
            m = re.match(r"^([AB]):\s*(.+)$", line)
            if m:
                cur["lines"].append((m.group(1), m.group(2)))
    return parts


async def speak(text: str, voice: str) -> bytes:
    data = b""
    comm = edge_tts.Communicate(text, voice, rate=RATE)
    async for chunk in comm.stream():
        if chunk["type"] == "audio":
            data += chunk["data"]
    return data


async def main():
    OUT.mkdir(exist_ok=True)
    parts = parse(SCRIPT.read_text(encoding="utf-8"))
    full = b""
    for n, part in enumerate(parts, 1):
        audio = b""
        for who, line in part["lines"]:
            for attempt in range(3):
                try:
                    audio += await speak(line, VOICES[who])
                    break
                except Exception as e:  # mạng chập chờn: thử lại
                    if attempt == 2:
                        raise
                    print("  thử lại:", e)
        name = f"{n:02d}_{re.sub(r'[^0-9A-Za-z]+', '_', part['title'])[:40]}.mp3"
        (OUT / name).write_bytes(audio)
        full += audio
        print("xong:", name)
    (OUT / "full.mp3").write_bytes(full)
    print("File ghép:", OUT / "full.mp3")


if sys.platform == "win32":
    # aiodns (do aiohttp dùng) cần SelectorEventLoop trên Windows
    asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())
asyncio.run(main())
