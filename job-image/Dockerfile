FROM python:3.9-alpine

RUN apk add --no-cache podman postgresql-dev

COPY ./program /program
WORKDIR /program

ENV PYTHONUNBUFFERED=1
ENV STORAGE_DRIVER=vfs

RUN pip install -r requirements.txt

CMD ["python", "-u", "main.py"]
