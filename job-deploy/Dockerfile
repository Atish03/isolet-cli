FROM python:3.9-alpine

WORKDIR /program

COPY ./requirements.txt requirements.txt
RUN pip install -r requirements.txt

COPY ./program /program

ENV PYTHONUNBUFFERED=1

CMD ["python", "-u", "main.py"]
