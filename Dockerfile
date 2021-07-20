FROM python:3.9-slim-buster

ENV POETRY_VERSION=1.1.7
VOLUME /data

RUN curl -sSL https://raw.githubusercontent.com/python-poetry/poetry/$POETRY_VERSION/get-poetry.py | python -

ENV PATH="/root/.poetry/bin:${PATH}"

COPY pyproject.toml pyproject.toml
COPY poetry.lock poetry.lock

RUN echo $PATH

RUN pip install --upgrade pip poetry
RUN poetry config virtualenvs.create false
RUN poetry install --no-interaction --no-ansi

COPY . .

ENTRYPOINT ["poetry","run","python","-m","cpi_pandemia"]

CMD ["--directory","/data"]
