FROM python:3.9-alpine

RUN apk add --no-cache podman

WORKDIR /program

COPY ./requirements.txt requirements.txt
RUN pip install -r requirements.txt

COPY ./program /program

ENV PYTHONUNBUFFERED=1
ENV STORAGE_DRIVER=vfs

CMD ["python", "-u", "main.py"]
