FROM python:3.9-slim-buster

ENV POETRY_VERSION=1.1.7
VOLUME /data

RUN pip install --upgrade pip poetry

COPY pyproject.toml pyproject.toml
COPY poetry.lock poetry.lock

RUN poetry config virtualenvs.create false
RUN poetry install --no-interaction --no-ansi

COPY . .

ENTRYPOINT ["poetry","run","python","-m","cpi_pandemia"]

CMD ["--directory","/data"]
