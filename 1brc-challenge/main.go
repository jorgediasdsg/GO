// Execution time: 26.98176896s (no seu ambiente/entrada)
// Credits: https://github.com/shraddhaag/1brc
//
// Este programa resolve a variação do problema "1BRC":
// - Lê um arquivo gigante de medições (city;temp\n).
// - Faz parsing e agrega por cidade: min, max, soma e contagem.
// - Paraleliza o processamento dividindo o arquivo em "chunks" (blocos)
//   alinhados em '\n', envia para um pool de workers via canais,
//   e faz o "reduce" mesclando os mapas parciais.
// - No fim, ordena as cidades e imprime "cidade=min/avg/max".
// Observação: temperaturas são tratadas como inteiros em décimos (ex.: 24.3 -> 243)
// para evitar custo de ponto flutuante durante o parsing/acúmulo; o float só aparece na saída.
// Sim, estou usando IA para aprender melhor a linguagem GO com este projeto.

package main

import (
	"bytes"         // usado para achar o índice do último '\n' dentro do chunk lido
	"errors"        // comparação de erros (ex.: io.EOF)
	"flag"          // leitura de flags de CLI (ex.: -input, -cpuprofile)
	"fmt"           // impressão formatada
	"io"            // io.EOF e interfaces de leitura
	"log"           // logs para erros ao criar perfis
	"math"          // math.Round para arredondamento de 1 casa decimal
	"os"            // acesso a arquivos e criação de perfis
	"runtime"       // runtime.NumCPU, GC, etc.
	"runtime/pprof" // perfis de CPU e memória (pprof)
	"runtime/trace" // trace de execução (timeline)
	"sort"          // ordenação da lista final de cidades
	"strings"       // strings.Builder para construir a saída
	"sync"          // WaitGroup para coordenar goroutines
	"time"          // medição do tempo total de execução
)

// Flags globais (lidas em main via flag.Parse)
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

// Nota: "tarce" está com typo; é apenas mensagem de ajuda, não afeta execução.
var executionprofile = flag.String("execprofile", "", "write tarce execution to `file`")
var input = flag.String("input", "", "path to the input file to evaluate")

func main() {
	start := time.Now() // marca o início para medir tempo total
	flag.Parse()        // lê as flags passadas via CLI

	// Se pediram trace (-execprofile), abrimos arquivo e iniciamos o trace.
	if *executionprofile != "" {
		f, err := os.Create("./profiles/" + *executionprofile)
		if err != nil {
			log.Fatal("could not create trace execution profile: ", err)
		}
		defer f.Close()
		trace.Start(f)
		defer trace.Stop()
	}

	// Se pediram pprof de CPU (-cpuprofile), iniciamos/stoppamos o perfil.
	if *cpuprofile != "" {
		f, err := os.Create("./profiles/" + *cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// Executa a lógica principal: leitura, parsing concorrente e agregação
	evaluate(*input)

	// Se pediram pprof de memória (-memprofile), força um GC e escreve o heap profile.
	if *memprofile != "" {
		f, err := os.Create("./profiles/" + *memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}

	// Tempo total
	fmt.Printf("Execution time: %s\n", time.Since(start))
}

// computedResult é apenas para organizar os dados calculados na forma pronta para ordenação/saída.
type computedResult struct {
	city          string
	min, max, avg float64
}

// evaluate coordena o fluxo alto nível:
// - chama o leitor/paralelizador para obter o mapa final (por cidade),
// - converte para slice, calcula médias em float, arredonda,
// - ordena por nome e formata a string de saída.
func evaluate(input string) string {
	mapOfTemp, err := readFileLineByLineIntoAMap(input)
	if err != nil {
		panic(err)
	}

	// Converte o mapa agregado em slice para ordenar e imprimir
	resultArr := make([]computedResult, len(mapOfTemp))
	var count int
	for city, calculated := range mapOfTemp {
		resultArr[count] = computedResult{
			city: city,
			// min/max/sum estão em décimos (int).
			// Convertemos para float e arredondamos para 1 casa decimal na saída.
			min: round(float64(calculated.min) / 10.0),
			max: round(float64(calculated.max) / 10.0),
			avg: round(float64(calculated.sum) / 10.0 / float64(calculated.count)),
		}
		count++
	}

	// Ordena alfabeticamente por cidade
	sort.Slice(resultArr, func(i, j int) bool {
		return resultArr[i].city < resultArr[j].city
	})

	// Monta a linha final (ex.: "City=10.2/15.3/22.1, ...")
	var stringsBuilder strings.Builder
	for _, i := range resultArr {
		stringsBuilder.WriteString(fmt.Sprintf("%s=%.1f/%.1f/%.1f, ", i.city, i.min, i.avg, i.max))
	}
	// Remove a ", " final cortando os últimos 2 caracteres.
	return stringsBuilder.String()[:stringsBuilder.Len()-2]
}

// cityTemperatureInfo guarda as estatísticas por cidade.
// Importante: usamos int64 em DÉCIMOS (ex.: 24.3 -> 243) para evitar custo com float durante parsing/acúmulo.
type cityTemperatureInfo struct {
	count int64
	min   int64
	max   int64
	sum   int64
}

// readFileLineByLineIntoAMap faz todo o pipeline de I/O e paralelização:
//
// 1) Abre o arquivo.
// 2) Cria dois canais:
//   - chunkStream: recebe blocos []byte (chunks) alinhados em '\n' prontos para parsing.
//   - resultStream: recebe mapas parciais (por chunk) calculados pelos workers.
//
// 3) Lança (NumCPU-1) workers que consomem chunks e produzem resultados parciais.
// 4) Em uma goroutine produtora, lê o arquivo em "chunkSize" e:
//   - acha o último '\n' do bloco,
//   - concatena com "leftover" do bloco anterior,
//   - envia o trecho com linhas completas para o chunkStream,
//   - mantém o restante (após o último '\n') como novo leftover.
//
// 5) Quando termina, fecha chunkStream, espera workers (wg.Wait) e fecha resultStream.
// 6) Consome resultStream e faz o "reduce" no mapa global.
// 7) Retorna o mapa final.
func readFileLineByLineIntoAMap(filepath string) (map[string]cityTemperatureInfo, error) {
	file, err := os.Open(filepath)
	if err != nil {
		panic(err) // aqui preferiram panicar; poderia retornar err para o caller decidir
	}
	defer file.Close()

	mapOfTemp := make(map[string]cityTemperatureInfo)

	// Canal de saída dos workers: cada item é um mapa parcial do chunk processado.
	resultStream := make(chan map[string]cityTemperatureInfo, 10)
	// Canal de entrada para os workers: cada item é um []byte com várias linhas completas.
	chunkStream := make(chan []byte, 15)

	// Tamanho do chunk: 32 MiB (menor que o anterior de 64 MiB no outro código).
	// Esse valor foi o que, no seu caso, deu o melhor desempenho.
	chunkSize := 32 * 1024 * 1024

	var wg sync.WaitGroup

	// -------------- POOL DE WORKERS --------------
	// Lança N-1 workers (deixa 1 core para o produtor I/O).
	for i := 0; i < runtime.NumCPU()-1; i++ {
		wg.Add(1)
		go func() {
			// Cada worker consome chunks e manda mapa parcial no resultStream.
			for chunk := range chunkStream {
				processReadChunk(chunk, resultStream)
			}
			wg.Done()
		}()
	}

	// -------------- PRODUTOR DE CHUNKS --------------
	// Lê o arquivo em blocos, garante que cada "toSend" termina em '\n',
	// e envia para os workers. Também gerencia "leftover" (a parte final
	// incompleta que ficou após o último '\n').
	go func() {
		buf := make([]byte, chunkSize)         // buffer de leitura reaproveitado
		leftover := make([]byte, 0, chunkSize) // sobra do bloco, sem '\n' no final, será completada no próximo loop
		for {
			readTotal, err := file.Read(buf) // lê até chunkSize bytes do arquivo
			if err != nil {
				if errors.Is(err, io.EOF) {
					break // fim do arquivo
				}
				panic(err)
			}
			buf = buf[:readTotal] // limita o slice aos bytes realmente lidos

			// Defensive copy do bloco lido (não é estritamente necessário se buf não for reusado depois,
			// mas aqui eles preferem copiar para garantir isolamento do slice).
			toSend := make([]byte, readTotal)
			copy(toSend, buf)

			// Encontra o último '\n' para não quebrar linhas entre chunks.
			lastNewLineIndex := bytes.LastIndex(buf, []byte{'\n'})

			// "toSend" vira leftover anterior + linhas completas deste bloco até o '\n' final.
			toSend = append(leftover, buf[:lastNewLineIndex+1]...)

			// "leftover" passa a ser o pedaço após o último '\n' (início de uma linha incompleta).
			leftover = make([]byte, len(buf[lastNewLineIndex+1:]))
			copy(leftover, buf[lastNewLineIndex+1:])

			// Envia chunk pronto (linhas completas) ao pool de workers.
			chunkStream <- toSend
		}
		// Fim da leitura: fecha o canal de chunks para sinalizar que não virão mais dados.
		close(chunkStream)

		// Espera todos os workers terminarem de processar os chunks enviados.
		wg.Wait()
		// Agora pode fechar o resultStream para sinalizar fim do reduce.
		close(resultStream)
	}()

	// -------------- REDUCE (MESCLA GLOBAL) --------------
	// Consome mapas parciais de cada chunk e mescla em mapOfTemp.
	for t := range resultStream {
		for city, tempInfo := range t {
			if val, ok := mapOfTemp[city]; ok {
				// Se a cidade já existe no mapa global, acumula:
				val.count += tempInfo.count
				val.sum += tempInfo.sum
				if tempInfo.min < val.min {
					val.min = tempInfo.min
				}
				if tempInfo.max > val.max {
					val.max = tempInfo.max
				}
				mapOfTemp[city] = val
			} else {
				// Cidade nova: só atribui.
				mapOfTemp[city] = tempInfo
			}
		}
	}

	return mapOfTemp, nil
}

// processReadChunk faz o parsing de UM chunk e produz um mapa parcial por cidade.
// Observação: esta versão converte o chunk inteiro em string (stringBuf := string(buf))
// e itera com "range" por rune. Apesar desse custo, no seu caso específico,
// isto performou melhor que alternativas por conta do equilíbrio com chunkSize, pool e GC.
// O formato esperado de cada linha é:
//
//	city;temp\n
//
// onde "temp" é texto do tipo -12.3, 0.0, 25.4 etc. (uma casa decimal).
func processReadChunk(buf []byte, resultStream chan<- map[string]cityTemperatureInfo) {
	toSend := make(map[string]cityTemperatureInfo) // mapa parcial local do worker
	var start int                                  // índice onde começa o campo atual na string
	var city string                                // cidade atual (capturada antes do ';')

	stringBuf := string(buf) // converte todo o chunk []byte -> string (gera alocação/cópia)
	for index, char := range stringBuf {
		switch char {
		case ';':
			// Encontrou separador entre cidade e temperatura.
			// city = trecho [start:index)
			city = stringBuf[start:index]
			// move o início (start) para após o ';' (temperatura começa aqui)
			start = index + 1

		case '\n':
			// Encontrou fim de linha. Temperatura está em [start:index).
			// Confere se há conteúdo e se city foi capturada.
			if (index-start) > 1 && len(city) != 0 {
				// Faz parsing rápido da temperatura para inteiro em décimos.
				temp := customStringToIntParser(stringBuf[start:index])
				// Próxima linha começa após o '\n'
				start = index + 1

				// Atualiza estatísticas no mapa parcial
				if val, ok := toSend[city]; ok {
					val.count++
					val.sum += temp
					if temp < val.min {
						val.min = temp
					}
					if temp > val.max {
						val.max = temp
					}
					toSend[city] = val
				} else {
					// Primeira ocorrência da cidade neste chunk
					toSend[city] = cityTemperatureInfo{
						count: 1,
						min:   temp,
						max:   temp,
						sum:   temp,
					}
				}
				// Limpa city para a próxima linha
				city = ""
			}
		}
	}
	// Envia o mapa parcial para o reduce global
	resultStream <- toSend
}

// round arredonda para 1 casa decimal, evitando imprimir "-0.0".
func round(x float64) float64 {
	rounded := math.Round(x * 10)
	if rounded == -0.0 {
		return 0.0
	}
	return rounded / 10
}

// customStringToIntParser converte uma string de temperatura no formato [-99.9, 99.9]
// em um inteiro em décimos (ex.: "24.3" -> 243, "-1.0" -> -10).
// É propositalmente enxuta e assume formatos restritos para ser rápida.
// Regras assumidas:
// - sinal opcional '-'
// - um ou dois dígitos antes do ponto
// - um dígito após o ponto
func customStringToIntParser(input string) (output int64) {
	var isNegativeNumber bool
	// Trata sinal
	if input[0] == '-' {
		isNegativeNumber = true
		input = input[1:] // remove o '-'
	}

	// Usa o comprimento para decidir o cálculo:
	// len==3: "d.d"  -> (d0*10 + d2) - '0'*11
	// len==4: "dd.d" -> (d0*100 + d1*10 + d3) - '0'*111
	switch len(input) {
	case 3:
		// Ex.: "3.5" -> ( '3'*10 + '5' ) - '0'*11 => (51 + 5) - 528 = 56 - 528? Não parece,
		// mas lembre: chars são bytes; a fórmula usa os códigos ASCII para subtrair '0' corretamente.
		// Explicando de forma mais clara:
		// '3' (51) * 10 + '5' (53) - '0'(48) * 11 = 510 + 53 - 528 = 35 -> 3.5 * 10 = 35
		output = int64(input[0])*10 + int64(input[2]) - int64('0')*11
	case 4:
		// Ex.: "12.3" -> '1'*100 + '2'*10 + '3' - '0'*111
		// 49*100 + 50*10 + 51 - 48*111 = 4900 + 500 + 51 - 5328 = 123 -> 12.3 * 10 = 123
		output = int64(input[0])*100 + int64(input[1])*10 + int64(input[3]) - (int64('0') * 111)
	}

	// Aplica sinal se necessário
	if isNegativeNumber {
		return -output
	}
	// teste de código
	return
}
