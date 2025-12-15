package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"

	"golang.org/x/tools/go/cfg"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <go-file>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		os.Exit(1)
	}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		fmt.Printf("=== Function: %s ===\n", fn.Name.Name)

		g := cfg.New(fn.Body, func(ce *ast.CallExpr) bool {
			return false 
		})

		fmt.Println("\nCFG Blocks:")
		printCFG(g, fset)

		graph, liveNodes := buildGraph(g)

		printGraphInfo(graph, liveNodes)

		simplePaths := findAllSimplePaths(graph, liveNodes)

		primePaths := filterPrimePaths(simplePaths)

		fmt.Println("\nPrime Paths:")
		for i, path := range primePaths {
			fmt.Printf("  %d: %v\n", i+1, path)
		}
		fmt.Println()
	}
}

func printCFG(g *cfg.CFG, fset *token.FileSet) {
	for _, block := range g.Blocks {
		fmt.Printf("  Block %d", block.Index)
		if block.Live {
			fmt.Print(" (live)")
		}
		fmt.Println()

		for _, node := range block.Nodes {
			var buf bytes.Buffer
			printer.Fprint(&buf, fset, node)
			fmt.Printf("      %s\n", buf.String())
		}

		if len(block.Succs) > 0 {
			fmt.Print("    -> ")
			for i, succ := range block.Succs {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("Block %d", succ.Index)
			}
			fmt.Println()
		}
	}
}

func buildGraph(g *cfg.CFG) ([][]int, int) {
	liveCount := 0
	for _, block := range g.Blocks {
		if block.Live {
			liveCount++
		}
	}

	graph := make([][]int, liveCount)
	for i := range graph {
		graph[i] = []int{}
	}

	for _, block := range g.Blocks {
		if !block.Live {
			continue
		}
		for _, succ := range block.Succs {
			if succ.Live {
				graph[block.Index] = append(graph[block.Index], int(succ.Index))
			}
		}
	}

	return graph, liveCount
}

func printGraphInfo(graph [][]int, n int) {
	fmt.Println("\nGraph Info:")

	fmt.Println("Edges:")
	for from := 0; from < n; from++ {
		for _, to := range graph[from] {
			fmt.Printf("  %d %d\n", from, to)
		}
	}

	hasIncoming := make([]bool, n)
	for from := 0; from < n; from++ {
		for _, to := range graph[from] {
			hasIncoming[to] = true
		}
	}
	fmt.Print("Initial nodes: ")
	first := true
	for i := 0; i < n; i++ {
		if !hasIncoming[i] {
			if !first {
				fmt.Print(", ")
			}
			fmt.Printf("%d", i)
			first = false
		}
	}
	fmt.Println()

	fmt.Print("Final nodes: ")
	first = true
	for i := 0; i < n; i++ {
		if len(graph[i]) == 0 {
			if !first {
				fmt.Print(", ")
			}
			fmt.Printf("%d", i)
			first = false
		}
	}
	fmt.Println()
}

func findAllSimplePaths(graph [][]int, n int) [][]int {
	var allPaths [][]int

	for start := 0; start < n; start++ {
		visited := make([]bool, n)
		path := []int{start}
		findPathsDFS(graph, start, visited, path, &allPaths, start)
	}

	return allPaths
}

func findPathsDFS(graph [][]int, node int, visited []bool, path []int, allPaths *[][]int, startNode int) {
	pathCopy := make([]int, len(path))
	copy(pathCopy, path)
	*allPaths = append(*allPaths, pathCopy)

	visited[node] = true

	for _, next := range graph[node] {
		if next == startNode && len(path) > 1 {
			cyclePath := make([]int, len(path)+1)
			copy(cyclePath, path)
			cyclePath[len(path)] = next
			*allPaths = append(*allPaths, cyclePath)
		} else if !visited[next] {
			path = append(path, next)
			findPathsDFS(graph, next, visited, path, allPaths, startNode)
			path = path[:len(path)-1]
		}
	}

	visited[node] = false
}

func filterPrimePaths(paths [][]int) [][]int {
	var primePaths [][]int

	for _, path := range paths {
		if isPrimePath(path, paths) {
			primePaths = append(primePaths, path)
		}
	}

	return primePaths
}

func isPrimePath(path []int, allPaths [][]int) bool {

	for _, other := range allPaths {
		if len(other) > len(path) && isProperSubpath(path, other) {
			return false
		}
	}
	return true
}

func isProperSubpath(sub, full []int) bool {
	if len(sub) >= len(full) {
		return false
	}

	for i := 0; i <= len(full)-len(sub); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
			if full[i+j] != sub[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
