FROM python:3.9

ENV VERSION=1.1.7

RUN curl -sSL https://raw.githubusercontent.com/python-poetry/poetry/$VERSION/get-poetry.py | python -

ENV PATH="/root/.poetry/bin:${PATH}"

COPY pyproject.toml pyproject.toml

RUN echo $PATH

RUN poetry install

CMD [ "poetry", "run", "python", "-m" "cpi_pandemia" ]