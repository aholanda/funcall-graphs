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
	var sep string = ";"

	file, err := os.Open("scc.dat")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Fprintf(file, "#vertices%s", sep)
	fmt.Fprintf(file, "#arcs%s", sep)
	fmt.Fprintf(file, "avg_degree: %s", sep)
	fmt.Fprintf(file, "std_dev: %s", sep)
	fmt.Fprintf(file, "#components: %s", sep)
	fmt.Fprintf(file, "greatest_comp_size: %s", sep)
	fmt.Fprintf(file, "\n")

	files := listDataFiles()
	for _, f := range files {
		log.Printf("> reading %v\n", f)
		digraph = gio.ReadPajek(f)
		scc := g.NewKosarajuSharirSCC(digraph)
		scc.Compute()
		fmt.Fprintf(file, "%d%s", digraph.V(), sep)
		fmt.Fprintf(file, "%d%s", digraph.A(), sep)
		avgDeg, stdDev := digraph.AverageDegree()
		fmt.Fprintf(file, "%f%s", avgDeg, sep)
		fmt.Fprintf(file, "%f%s", stdDev, sep)
		fmt.Fprintf(file, "%d%s", scc.Count(), sep)
		fmt.Fprintf(file, "%d%s", scc.GreatestComponentSize(), sep)
		fmt.Fprintf(file, "\n")
		break
	}
}
