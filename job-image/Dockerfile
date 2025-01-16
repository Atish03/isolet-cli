FROM python:3.9-alpine

ENV PYTHONUNBUFFERED=1

COPY . .

RUN pip install -r requirements.txt

CMD ["python", "-u", "main.py"]
