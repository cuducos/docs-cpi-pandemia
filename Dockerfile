FROM python:3.9

ENV POETRY_VERSION=1.1.7

RUN curl -sSL https://raw.githubusercontent.com/python-poetry/poetry/$POETRY_VERSION/get-poetry.py | python -

ENV PATH="/root/.poetry/bin:${PATH}"

COPY pyproject.toml pyproject.toml

RUN echo $PATH

RUN pip install --upgrade pip poetry
RUN poetry config virtualenvs.create false
RUN poetry install --no-interaction --no-ansi

COPY . .

CMD ["poetry","run","python","-m","cpi_pandemia"]
