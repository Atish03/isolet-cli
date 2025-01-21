FROM python:3.9-alpine

RUN apk add --no-cache buildah fuse-overlayfs libc-dev

ENV PYTHONUNBUFFERED=1
ENV STORAGE_DRIVER=vfs

WORKDIR /source

COPY kubernetes_client.py .
COPY main.py .
COPY requirements.txt .
COPY bin/netavark /usr/local/bin/
COPY config/containers.conf /etc/containers/containers.conf

RUN chown root:root /usr/local/bin/netavark
RUN chmod +x /usr/local/bin/netavark

RUN pip install -r requirements.txt

CMD ["python", "-u", "main.py"]
