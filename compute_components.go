package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"

	g "github.com/aholanda/graphs"
	gio "github.com/aholanda/graphs/io"
)

func listDigraph(d *g.Digraph) {
	vIter := g.NewVertexIterator(d)
	for vIter.HasNext() {
		v := vIter.Value()
		fmt.Printf("%v:", v)

		aIter := g.NewArcIterator(d, v)
		for aIter.HasNext() {
			w := aIter.Value()
			fmt.Printf(" %v", w)
		}
		fmt.Println()
	}
}

func checkDataDirExists(path string) {
	dirInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Fatalf("fatal: data directory \"%s\" does not exist.", path)
	}
	log.Printf("ok> Directory \"%d\" found.\n", dirInfo)
}

func listDataFiles() []string {
	pathS, err := os.Getwd()
	pathS = path.Join(pathS, "data")
	if err != nil {
		panic(err)
	}

	checkDataDirExists(pathS)

	var files []string
	filepath.Walk(pathS, func(fpath string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			r, err := regexp.MatchString(".net", f.Name())
			if err == nil && r {
				filepath := path.Join("data/", f.Name())
				files = append(files, filepath)
			}
		}
		return nil
	})
	return files
}

func computeComponents() {
	var digraph *g.Digraph

	files := listDataFiles()

	for _, f := range files {
		log.Printf("> reading %v\n", f)
		digraph = gio.ReadPajek(f)
		scc := g.NewKosarajuSharirSCC(digraph)
		scc.Compute()
		fmt.Printf("#vertices: %d\n", digraph.V())
		fmt.Printf("#arcs: %d\n", digraph.A())
		fmt.Printf("#components: %d\n", scc.Count())
		break
	}
}
