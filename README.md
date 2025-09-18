# Documentos da CPI da Covid

A [CPI da Covid](https://legis.senado.leg.br/comissoes/comissao?codcol=2441) recebeu milhares de documentos **públicos**, [todos disponibilizados no site do Senado Federal](https://legis.senado.leg.br/comissoes/docsRecCPI?codcol=2441).

Mas como clicar um por um leva tempo, esse programa facilita o **download** e a **descopactação** de todos esses arquivos, possibilitando assim não só o acesso, mas também buscas nos arquivos com ferramentas como Evernote, Spotlight, etc.

## Avisos importantes

### Nome dos arquivos

Para sincronizar esses arquivos na nuvem e evitar erros no sistema de arquivos, todos os nomes de arquivos foram normalizados retirando acentuação e caracteres especiais.

Por exemplo, um arquivo chamado `Ofício.text` é renomeado para `Oficio.txt`.

### Erros

Algumas links para baixar os documentos públicos não funcionam. Mesmo com estratégias de repetir a tentativa em caso de erro, pode ser que nem todos os arquivos listados estejam, de fato, disponível.

Executando o programa com `--tolerant` faz com que o download dos arquivos restantes prossiga mesmo se o downlaod de algum arquivo específico falhar.

## Só quero baixar os arquivos

### Rodando o `docs-cpi-pandemia` localmente, sem saber de programação

[Baixe o executável](https://github.com/cuducos/docs-cpi-pandemia/releases) compatível com o seu sistema operacional e arquitetura. Execute esse programa no terminal (ou prompt de comando) do seu computador.

Existem opções que podem ser configuradas, as instruções e valores padrões podem ser vistos adicionando `--help` ao final do comando.

## Sou _hacker_ e quero mais

Você também pode baixar tudo direto do Senado Federal, instalando esse pacote e digitando apenas um comando.

### Utilizando Go nativo

Requer [Go](https://golang.org/) 1.25.

```console
$ go run main.go --help
```

### Utilizando com docker

Requer [Docker](https://docker.com):

```console
$ docker build -t docs-cpi-pandemia .
$ docker run -it -v $PWD/data:/docs-cpi-pandemia/data docs-cpi-pandemia
```

Os arquivos serão baixados em um diretório `data/` dentro da pasta onde você executou esse comando.
