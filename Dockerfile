FROM python:3.12-slim
WORKDIR /app
COPY server.py index.html ./
EXPOSE 8765
ENV PORT=8765
CMD ["sh","-c","python3 server.py $PORT"]
