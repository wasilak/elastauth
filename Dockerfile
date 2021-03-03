FROM python:3-alpine

RUN apk --no-cache --update add build-base

COPY ./requirements.txt /requirements.txt

RUN pip install -U -r requirements.txt

ADD src /app

WORKDIR /app

EXPOSE 5000

ENTRYPOINT ["gunicorn", "--bind=0.0.0.0:5000", "--workers=1", "--worker-class=gthread", "--preload", "app:app"]

CMD [ "--log-level=debug" ]
