FROM python:3.12-slim
RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg && rm -rf /var/lib/apt/lists/*
WORKDIR /app
# edge-tts follows the changes Microsoft makes to its service: rebuild to update
RUN pip install --no-cache-dir "edge-tts>=7.2" "aiohttp>=3.9"
COPY app.py .
# readable by "nobody" whatever the permissions of the copied file were
RUN chmod 644 app.py
USER nobody
ENV PYTHONUNBUFFERED=1
EXPOSE 5050
CMD ["python", "app.py"]
