from asyncio import Semaphore, gather
from dataclasses import dataclass
from functools import cached_property
from pathlib import Path
from shutil import rmtree
from typing import Optional, Tuple
from unicodedata import category, normalize
from urllib.parse import urlparse
from zipfile import ZipFile

import aiofiles
import backoff
from httpx import AsyncClient, HTTPError

from pyquery import PyQuery
from tqdm import tqdm


RETRIES = 8
URL = "https://legis.senado.leg.br/comissoes/docsRecCPI?codcol=2441"
PREFIX = "https://legis.senado.leg.br/sdleg-getter/documento/download/"


@dataclass
class Dowloader:
    workers: int
    target: Path
    cleanup: bool

    def __post_init__(self) -> None:
        self.cache = self.target / ".cache"

    def cache_for(self, url) -> Path:
        *_, uuid = urlparse(url).path.split("/")
        parts = uuid.split("-")
        return self.cache.joinpath(*parts)

    def set_cache(self, url: str) -> None:
        path = self.cache_for(url)
        path.parent.mkdir(exist_ok=True, parents=True)
        path.touch()

    def skip_download(self, url) -> bool:
        return self.cache_for(url).exists()

    @cached_property
    def urls(self) -> Tuple[str, ...]:
        doc = PyQuery(url=URL)
        urls = (link.attrib.get("href", "") for link in doc("a"))
        return tuple(set(url for url in urls if url.startswith(PREFIX)))

    @staticmethod
    def normalize(value: str) -> str:
        return "".join(
            char for char in normalize("NFD", value) if category(char) != "Mn"
        )

    def normalize_path(self, path: Path):
        new_path = path.parent / self.normalize(path.name)
        path.rename(new_path)

    def normalize_path_recursively(self, path: Path):
        for item in path.glob("**/*"):
            if item.is_dir():
                continue

            if item.name == self.normalize(item.name):
                continue

            self.normalize_path(item)

        for item in path.glob("**/*"):
            if item.is_file():
                continue

            if item.name == self.normalize(item.name):
                continue

            self.normalize_path(item)
            self.normalize_path_recursively(path)
            return

        self.normalize_path(path)

    def unzip(self, path: Path) -> None:
        directory = path.parent / path.stem
        directory.mkdir(exist_ok=True)

        with ZipFile(path, "r") as archive:
            archive.extractall(directory)

        self.normalize_path_recursively(directory)
        path.unlink()

    @backoff.on_exception(backoff.expo, HTTPError, max_tries=RETRIES)
    async def _download(self, client: AsyncClient, url: str) -> Optional[str]:
        resp = await client.get(url)
        if resp.status_code != 200:
            return f"Erro no download de {url} (HTTP {resp.status_code})"

        *_, name = resp.headers["content-disposition"].split("=")
        path = self.target / self.normalize(name)
        async with aiofiles.open(path, mode="wb") as handler:
            await handler.write(resp.content)

        self.set_cache(url)
        if path.suffix.lower() == ".zip":
            self.unzip(path)

    async def download(self, bar: tqdm, client: AsyncClient, url: str) -> Optional[str]:
        if self.skip_download(url):
            bar.update(1)
            return

        try:
            result = await self._download(client, url)
        except HTTPError as error:
            result = f"Erro no download de {url} ({error})"

        bar.update(1)
        return result

    async def __call__(self) -> None:
        if self.cleanup:
            for item in self.target.iterdir():
                if item.is_file():
                    item.unlink()
                if item.is_dir():
                    rmtree(item)

        self.target.mkdir(exist_ok=True, parents=True)
        self.cache.mkdir(exist_ok=True)

        total = len(self.urls)
        with tqdm(total=total, desc="Baixando arquivos", unit="arquivos") as bar:
            async with AsyncClient() as client:
                semaphore = Semaphore(self.workers)
                async with semaphore:
                    tasks = tuple(self.download(bar, client, url) for url in self.urls)
                    results = await gather(*tasks)

        errors = tuple(error for error in results if error)
        if errors:
            (self.target / "erros.txt").write_text("\n".join(errors))
