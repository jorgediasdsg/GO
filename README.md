<h1 align="center">:zap: Processador de Dados em Go — 1B Linhas :zap:</h1>

<p align="center">
<b>:us:</b> High-performance data processing in Go, handling datasets with up to <i>1,000,000,000</i> lines.<br>
<b>:brazil:</b> Processamento em alta performance com Go, operando arquivos de até <i>1.000.000.000</i> linhas.
</p>

<p align="center">
 <a href="#history">História</a> •
 <a href="#objective">Objetivo</a> •
 <a href="#technologies">Tecnologias</a> •
 <a href="#dataset">Dataset</a> •
 <a href="#how-to-run">Como Executar</a> •
 <a href="#expected-output">Saída Esperada</a> •
 <a href="#code-notes">Notas de Código</a> •
 <a href="#perf-notes">Dicas de Performance</a> •
 <a href="#next-steps">Benchmarks de Execução</a>
</p>

---

<h1 id="history">:book: História</h1>

**EN** — This project was built to train Go by chewing through a massive text file and producing per-location stats (min/avg/max), sorted and printed in a single line, with total execution time. It’s a practical benchmark to feel Go’s I/O and aggregation performance under pressure.

**BR** — Este projeto foi criado para treinar Go processando um arquivo texto gigantesco e gerando estatísticas por localidade (mín/méd/máx), ordenadas e exibidas em uma única linha, além do tempo total de execução. Serve como um benchmark prático para sentir a performance de I/O e agregação do Go sob carga.

---

<h1 id="objective">:bulb: Objetivo</h1>

- **Processar** um arquivo `measurements.txt` com até **1 bilhão de linhas**.  
- **Calcular** estatísticas por localidade: **mínimo**, **média** e **máximo**.  
- **Ordenar** alfabeticamente as localidades e **imprimir** o resultado e o **tempo de execução**.

---

<h1 id="technologies">:rocket: Tecnologias</h1>

- **Go** (>= 1.24)  
- **Python 3** (apenas para gerar o dataset de teste)

---

<h1 id="dataset">:card_file_box: Dataset</h1>

Cada linha do arquivo segue o formato:
```
<localidade>;<temperatura>
Lisbon;13.7
Reykjavik;-1.0
Florianopolis;24.3
```


Para criar um arquivo de **1 bilhão de linhas**, use o gerador em Python:

```bash
python3 create.py 1_000_000_000
```

⚠️ Tamanho & Disco: 1B linhas costuma significar dezenas de GB. Garanta espaço em disco e sistema de arquivos adequado.

<h1 id="how-to-run">:computer: Como Executar</h1>
1) Rodar direto com go run

```shell
# Certifique-se de que measurements.txt está no mesmo diretório de main.go
go run main.go
```

2) Compilar (binários nativos e cross-compile)
```shell
# Linux (x86_64)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o processor_linux main.go

# Windows (x86_64)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o processor_windows.exe main.go

# macOS Intel/AMD (x86_64)
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o processor_macos_amd64 main.go

# macOS Apple Silicon (arm64)
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o processor_macos_arm64 main.go

```

3) Executar o binário
```shell
# Linux/macOS
./processor_linux

# Windows
processor_windows.exe
```

<h1 id="expected-output">:printer: Saída Esperada</h1>
Formato simplificado:

```shell
{'Florianopolis=10.2/18.5/33.7', 'Lisbon=7.0/15.1/28.2', 'Reykjavik=-5.0/0.4/5.3'}
Execution time: 2.345678s
```

Cada entrada é localidade=min/avg/max com uma casa decimal.
Linhas são ordenadas por localidade.
Ao final, é exibido o tempo total decorrido.

<h1 id="code-notes">:microscope: Notas de Código</h1>
Sombras de nome (measurements): o identificador é usado tanto para o arquivo quanto para o valor do mapa. Funciona, mas reduz a legibilidade. Considere renomear o valor do mapa para m ou agg.
Erros ignorados em ParseFloat: o retorno de erro é descartado (_). Para robustez em dados reais, trate erros e contabilize linhas inválidas.
Formatação da saída: há uma vírgula e espaço após o último item. Se quiser uma saída estritamente limpa, trate o separador (ex.: strings.Builder + join manual).
bufio.Scanner: ótimo para linhas curtas. Para linhas muito longas, aumente o Buffer. Aqui as linhas são pequenas, então está ok.
I/O da impressão: imprimir dentro do loop final é aceitável; em dumps gigantes, use strings.Builder para reduzir syscalls.

<h1 id="perf-notes">:stopwatch: Dicas de Performance</h1>
Executáveis enxutos: use -trimpath -ldflags="-s -w" para reduzir o tamanho do binário.
CPU: deixe o Go usar todos os núcleos (GOMAXPROCS padrão já faz isso).
Disco: prefira SSD NVMe; o gargalo geralmente é I/O.
Formato de entrada: manter linhas curtas acelera Scanner.
Mapas: se o número de localidades for conhecido/estimado, faça make(map[string]Measurement, N) para evitar rehash inicial.

<h1 id="next-steps">:stopwatch: Benchmarks de Execução</h1>

| Versão   | Descrição                                                                                                                                          | Tempo            | Commit                                                                                       |
|----------|----------------------------------------------------------------------------------------------------------------------------------------------------|------------------|----------------------------------------------------------------------------------------------|
| Versão 1 | Teste com o exemplo da Rocketseat                                                                                                                  | 2m24.23696897s   | [40454f0](https://github.com/jorgediasdsg/GO/commit/40454f0ec5aa576b7e994f2a74199a0642293b2f) |
| Versão 2 | Adicionado multi-thread direto, o tempo aumentou                                                                                                   | 8m57.803627255s  | [fd1edb0](https://github.com/jorgediasdsg/GO/commit/fd1edb095ccda1fc5d9061925346209eda26b7a1) |
| Versão 3 | Utilizado projeto de estudo da **shraddhaag**, trocando leitura linha a linha (overhead) por chunks, aproveitando melhor a memória bloqueada       | 26.98176896s     | [29f1fd6](https://github.com/jorgediasdsg/GO/commit/29f1fd6d91e45c7db6686678f44810aaf5e753ff) |

Utilizei neste projeto apoio de IA com ChatGPT e Gemini para entender melhor os fluxos da linguagem GO para facilitar meu aprendizado.

<p align="center"> <sub>@jorgediasdsg — 2025</sub> </p> 


