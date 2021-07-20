import asyncio
from pathlib import Path

from typer import BadParameter, Option, run

from cpi_pandemia import Dowloader


def validate_workers(number: int) -> int:
    if number <= 0:
        raise BadParameter("Número de downlaods paralelos tem que ser positivo")
    return number


def validate_directory(value: str) -> Path:
    path = Path(value)
    if path.exists() and path.is_file():
        raise BadParameter(f"{value} é um arquivo")
    return path


def main(
    workers: int = Option(
        16, help="Número de downloads paralelos.", callback=validate_workers
    ),
    directory: Path = Option(
        Path("data"),
        help="Diretório onde salvar os arquivos.",
        callback=validate_directory,
    ),
    cleanup: bool = Option(
        False,
        help="Limpa o diretório antes de iniciar o download.",
    ),
) -> None:
    downloader = Dowloader(workers, directory, cleanup)
    asyncio.run(downloader())


if __name__ == "__main__":
    run(main)
