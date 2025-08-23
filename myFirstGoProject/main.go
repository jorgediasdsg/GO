// main.go — Manual Go em 1 arquivo
// Go 1.24+ requerido
// Tema: app de demonstração CLI

package main // pacote principal

import ( // imports stdlib
	"errors"  // erros simples
	"fmt"     // saída padrão
	"strings" // util de strings
	"time"    // medir tempo
)

// ===== Checklist rápido =====
// [ ] Compila e roda
// [ ] globalVar e local em uso
// [ ] if/else if/else presentes
// [ ] 3 formas de for
// [ ] Array [N]T iterado
// [ ] Ponteiro criado e alterado
// [ ] Exportada + privada + variádica + múltiplos retornos
// [ ] defer em LIFO + tempo
// [ ] Todas funções invocadas
// [ ] Somente stdlib

// ===== Variáveis e constantes =====
const appTitle string = "Manual Go em 1 arquivo" // constante curta
const tema string = "app de demonstração CLI"    // constante curta
const nArray = 5                                 // N do array

var globalVar int = 42 // variável global

// ===== Função exportada =====
// Mini-demo:
// DoThing("oi") // maiúsculas
func DoThing(msg string) string { // função exportada
	// retorna em maiúsculas
	return strings.ToUpper(msg)
}

// ===== Função não exportada =====
// Mini-demo:
// doHelper(2, 3) // soma
func doHelper(a, b int) int { // função privada
	// soma simples
	return a + b
}

// ===== Variádica =====
// Mini-demo:
// sum(1,2,3) // 6
func sum(nums ...int) int { // função variádica
	total := 0               // acumulador local
	for _, v := range nums { // for range
		total += v // soma valores
	}
	return total // retorna total
}

// ===== Múltiplos retornos (T, error) =====
// Mini-demo:
// parseEven(4) // ok
// parseEven(3) // erro
func parseEven(n int) (string, error) { // dois retornos
	if n%2 != 0 { // if simples
		return "", errors.New("n deve ser par") // retorna erro
	} else if n == 0 { // else if extra
		return "zero é par", nil // caso especial
	} else { // else final
		return fmt.Sprintf("par:%d", n), nil // mensagem ok
	}
}

// ===== Ponteiros =====
// Mini-demo:
// x := 10; p := &x // & pega endereço
// increment(p)     // * deref e altera
func increment(p *int) { // altera via ponteiro
	*p = *p + 1 // usa *
}

// ===== For em três formas + Array =====
// Mini-demo:
// loopDemo(4) // executa três fors
func loopDemo(n int) (sumClassic int, countWhile int, sumRange int) {
	// for clássico
	for i := 0; i < n; i++ { // estilo C
		sumClassic += i // acumula i
	}
	// for estilo while
	c := n      // contador local
	for c > 0 { // condição booleana
		countWhile++ // conta iterações
		c--          // decrementa
	}
	// array fixo [N]T
	var arr [nArray]int           // array fixo
	for i := 0; i < nArray; i++ { // preencher
		arr[i] = i + 1 // valores 1..N
	}
	// range sobre array
	for _, v := range arr { // itera array
		sumRange += v // soma valores
	}
	return // nomeados acima
}

// ===== defer: LIFO + tempo =====
// Mini-demo:
// timedSection("simulação", func(){ time.Sleep(20*time.Millisecond) })
func timedSection(name string, work func()) {
	start := time.Now() // início medição
	defer func() {      // defer de tempo
		elapsed := time.Since(start)         // duração total
		fmt.Println("tempo:", name, elapsed) // imprime tempo
	}() // executa ao retornar

	defer fmt.Println("defer 2 (antes)")    // empilhado depois
	defer fmt.Println("defer 1 (primeiro)") // LIFO demonstração

	work() // executa tarefa
}

// ===== main: usa tudo =====
func main() {
	fmt.Println("==", appTitle, "==") // título cabeçalho
	fmt.Println("Tema:", tema)        // mostra tema

	// variáveis locais
	localMsg := "olá, go"                                       // inferência :=
	var localNum int = 7                                        // tipo explícito
	fmt.Println("globalVar:", globalVar, "localNum:", localNum) // imprime vars

	// usar função exportada
	up := DoThing(localMsg)     // transforma msg
	fmt.Println("DoThing:", up) // resultado string

	// usar função privada
	sumAB := doHelper(2, 3)         // soma dois ints
	fmt.Println("doHelper:", sumAB) // imprime soma

	// usar variádica
	total := sum(1, 2, 3, 4)   // soma variádica
	fmt.Println("sum:", total) // imprime total

	// múltiplos retornos com if/else
	txt, err := parseEven(4) // caso par
	if err != nil {          // checar erro
		fmt.Println("erro:", err) // trata erro
	} else if strings.Contains(txt, "zero") { // else if exemplo
		fmt.Println("parseEven zero:", txt) // caso zero
	} else { // else final
		fmt.Println("parseEven ok:", txt) // imprime ok
	}

	// for + array demonstração
	a, b, c := loopDemo(4)                 // chama demo
	fmt.Println("for clássico soma:", a)   // resultado clássico
	fmt.Println("while-like contagem:", b) // resultado while-like
	fmt.Println("range array soma:", c)    // resultado array

	// ponteiro: criar e alterar
	x := 10                             // valor inicial
	p := &x                             // & pega endereço
	fmt.Println("antes increment:", x)  // valor antes
	increment(p)                        // altera via *
	fmt.Println("depois increment:", x) // valor depois

	// defer com tempo e LIFO
	timedSection("tarefa rápida", func() { // mede seção
		time.Sleep(20 * time.Millisecond)  // simula trabalho
		fmt.Println("executando trabalho") // log simples
	})

	// if/else erro real
	_, err2 := parseEven(3) // valor ímpar
	if err2 != nil {        // erro esperado
		fmt.Println("erro esperado:", err2) // imprime erro
	} else {
		fmt.Println("não deveria ocorrer") // caminho alternativo
	}

	// Finalização amigável
	fmt.Println("Fim da execução.") // encerramento
}
