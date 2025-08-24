// Version 2.0 - 2025-08-23
// Execution time: 8m57.803627255s
// with optimizations multi-threaded

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Measurement struct {
	Min   float64
	Max   float64
	Sum   float64
	Count int64
}

func (m *Measurement) add(v float64) {
	if m.Count == 0 {
		m.Min, m.Max, m.Sum, m.Count = v, v, v, 1
		return
	}
	if v < m.Min {
		m.Min = v
	}
	if v > m.Max {
		m.Max = v
	}
	m.Sum += v
	m.Count++
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func main() {
	start := time.Now()

	// Flags úteis para tunar paralelismo e buffer
	input := flag.String("input", "../../measurements.txt", "arquivo de entrada")
	workers := flag.Int("workers", runtime.NumCPU(), "número de workers")
	scanBuf := flag.Int("scanbuf", 256*1024, "tamanho do buffer do scanner (bytes)") // 256 KB
	flag.Parse()

	f, err := os.Open(*input)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// 1) Producer: lê linhas e envia para o canal
	lines := make(chan string, *workers*4)

	var prodWG sync.WaitGroup
	prodWG.Add(1)
	go func() {
		defer prodWG.Done()
		sc := bufio.NewScanner(f)

		// Aumenta o buffer para linhas maiores (por segurança)
		buf := make([]byte, 0, *scanBuf)
		sc.Buffer(buf, *scanBuf)

		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
		// Se der erro de scanner (linha gigante, etc.)
		if err := sc.Err(); err != nil {
			// Numa solução robusta, trate isso melhor (log/retentar/contabilizar)
			fmt.Fprintf(os.Stderr, "scanner error: %v\n", err)
		}
	}()

	// 2) Pool de workers: cada worker mantém um mapa local para reduzir contenção
	type partial = map[string]Measurement
	partitions := make([]partial, *workers)

	var wg sync.WaitGroup
	wg.Add(*workers)
	for i := 0; i < *workers; i++ {
		idx := i
		partitions[idx] = make(partial, 4096) // estimativa inicial; ajuste conforme seu dataset
		go func(local partial) {
			defer wg.Done()
			for raw := range lines {
				semicolon := strings.IndexByte(raw, ';')
				if semicolon <= 0 || semicolon >= len(raw)-1 {
					// linha inválida; ignore ou contabilize
					continue
				}
				loc := raw[:semicolon]
				rawTemp := raw[semicolon+1:]
				temp, err := strconv.ParseFloat(rawTemp, 64)
				if err != nil {
					// dado inválido; ignore ou contabilize
					continue
				}
				m := local[loc] // zero value ok
				// caminho hot sem chamada de função extra (mais rápido que m.add)
				if m.Count == 0 {
					m.Min, m.Max, m.Sum, m.Count = temp, temp, temp, 1
				} else {
					if temp < m.Min {
						m.Min = temp
					}
					if temp > m.Max {
						m.Max = temp
					}
					m.Sum += temp
					m.Count++
				}
				local[loc] = m
			}
		}(partitions[idx])
	}

	// Aguarda producer terminar de alimentar o canal e workers consumirem tudo
	prodWG.Wait()
	wg.Wait()

	// 3) Reduce: mescla todos os mapas locais em um mapa global
	global := make(map[string]Measurement, 1<<15) // chute inicial
	for _, p := range partitions {
		for k, v := range p {
			if g, ok := global[k]; ok {
				// merge: min/max/sum/count
				g.Min = min(g.Min, v.Min)
				g.Max = max(g.Max, v.Max)
				g.Sum += v.Sum
				g.Count += v.Count
				global[k] = g
			} else {
				global[k] = v
			}
		}
	}

	// 4) Ordena chaves e imprime
	locations := make([]string, 0, len(global))
	for name := range global {
		locations = append(locations, name)
	}
	sort.Strings(locations)

	// Constrói saída com builder para reduzir syscalls
	var b strings.Builder
	b.WriteByte('{')
	for i, name := range locations {
		m := global[name]
		avg := m.Sum / float64(m.Count)
		fmt.Fprintf(&b, "'%s=%.1f/%.1f/%.1f'", name, m.Min, avg, m.Max)
		if i != len(locations)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteByte('}')
	b.WriteByte('\n')
	fmt.Print(b.String())

	fmt.Printf("Workers: %d\n", *workers)
	fmt.Printf("Execution time: %s\n", time.Since(start))
}
