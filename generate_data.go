package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	g "github.com/aholanda/graphs"
	gio "github.com/aholanda/graphs/io"
)

const (
	// GraphFormatExtension is the suffix used in the files
	// containing graph description, e.g., ".dot", ".net".
	GraphFormatExtension string = gio.PajekFormatExtension
	// GraphDataRelativePath is the relative path in the current
	// directory where the main program is running. Inside it data with
	// graph description are saved.
	GraphDataRelativePath   string = "data"
	termBrowser             string = "/usr/bin/lynx -dump "
	downloader              string = "/usr/bin/wget"
	compressedFileExtension string = ".tar.xz"
)

func check(e error) {
	if e != nil {
		log.Fatalf("failed with %s", e)
	}
}

func extractFilePrefixFromFileURL(fileURL string) string {
	return strings.TrimRight(filepath.Base(fileURL),
		compressedFileExtension)
}

func listRemoteDir(url string, filter string) []string {
	cmdStr := termBrowser + url + " " + filter
	log.Println(cmdStr)
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	out, err := cmd.CombinedOutput()

	// TODO: check error due broken connection
	check(err)
	return strings.Fields(string(out))
}

func downloadFile(filepath string, fileURL string) error {
	// Get the data
	resp, err := http.Get(fileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	log.Println("downloading", filepath)
	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func unpackFile(file string) (string, error) {
	// Get the temporary absolute path where the file
	// is unpacked.
	dir, err := filepath.Abs(filepath.Dir(file))
	if err != nil {
		return "", err
	}

	// Unpack the file
	cmdStr := "/usr/bin/tar xfJ " + file + " -C " + dir
	log.Println(cmdStr)
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	// Get the name of the program with the version appended as suffix.
	versionName := extractFilePrefixFromFileURL(file)
	progName := strings.Split(versionName, "-")[0]

	// Fix when directory name is without the version as suffix.
	// e.g. $ [ -d /tmp/linux1234/linux ] \
	// && mv /tmp/linux1234/linux /tmp/linux1234/linux-v1.0
	cmdStr = fmt.Sprintf("[ -d  %s ] && mv %s %s",
		path.Join(dir, progName),
		path.Join(dir, progName),
		path.Join(dir, versionName))
	log.Println(cmdStr)
	cmd = exec.Command("/bin/bash", "-c", cmdStr)
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return versionName, nil
}

func addVertex(v2Adj map[string][]string, vertex string) map[string][]string {
	if _, ok := v2Adj[vertex]; !ok {
		v2Adj[vertex] = []string{}
	}
	return v2Adj
}

func addAdjVertex(v2Adj map[string][]string, from, to string) map[string][]string {
	if _, ok := v2Adj[from]; ok {
		v2Adj[from] = append(v2Adj[from], to)
	} else {
		v2Adj[from] = []string{to}
	}

	return v2Adj
}

func buildGraph(v2Adj map[string][]string) *g.Digraph {
	var digraph = g.NewDigraph(len(v2Adj))
	var vcount g.VertexId

	// The digraph is built in two passes. In the first
	// the vertices' names are indexed and in the second
	// the adjacencies are assigned.

	// First pass: vertices' names
	vcount = 0
	for key := range v2Adj {
		digraph.NameVertex(vcount, key)
		vcount++
	}

	// Second pass: adjacencies
	for from, adj := range v2Adj {
		v, err := digraph.VertexIndex(from)
		if err != nil {
			log.Fatalf("%v", err)
		}
		for _, to := range adj {
			w, err := digraph.VertexIndex(to)
			if err != nil {
				log.Fatalf("%v", err)
			}
			digraph.AddArc(v, w)
		}
	}
	return digraph
}

func createGraphFromCflowsOutput(dir string) (*g.Digraph, error) {
	// Map to accumulate vertices and its adjacent lists before
	// adding to graph, this procedure is needed because the
	// number of vertices is not know at first sight
	var vertexToAdj map[string][]string = make(map[string][]string)
	// Levels of indentation in the cflow output;
	// "0": first level, refers to the function caller
	// "1": second level, refers to the function callee
	var indentLevels [2]string = [2]string{"0", "1"}

	// List all C files recursivelly in the specified directory.
	cmdStr := fmt.Sprintf("/usr/bin/find %s -name \\*.c", dir)
	log.Println(cmdStr)
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	cfiles := strings.Fields(string(out))
	for _, cfile := range cfiles {
		// Current function caller during the call flow
		var curCaller string = ""

		// Run cflows on C files to get the function caller
		// at the first level (no indent) and the functions
		// callees at the second level (indented).
		cmdStr := fmt.Sprintf("/usr/bin/cflow --depth 2 "+
			" --omit-arguments --print-level %s", cfile)
		log.Println(cmdStr)
		cmd := exec.Command("/bin/bash", "-c", cmdStr)
		out, err = cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		cflowOutLines := strings.Split(string(out), "\n")
		for _, line := range cflowOutLines {
			for l, level := range indentLevels {
				exp := fmt.Sprintf(`^\{\s+%s\}.+`, level)
				re := regexp.MustCompile(exp)
				matched := re.MatchString(line)
				if matched {
					exp = fmt.Sprintf(`\{\s+%s\}\s+(?P<function>\w+)\(\)`+
						`\s*(?P<rest>.*)`, level)
					re = regexp.MustCompile(exp)
					ret := re.FindStringSubmatch(line)
					funcName := ret[1]
					// Sometimes a function performs no function call, so we
					// add to hashmap to be counted as vertex after.
					vertexToAdj = addVertex(vertexToAdj, funcName)
					if l == 0 { // level = 0 -> function caller
						curCaller = funcName
					} else { // level = 1 -> function callee
						vertexToAdj = addAdjVertex(vertexToAdj, curCaller, funcName)
					}
					//fmt.Println(level, ret[1])
					break
				}
			}
		}
	}
	return buildGraph(vertexToAdj), nil
}

func _GenerateData(p *program, version string) {
	// Append the program version to its base URL
	// to have access to the files.
	url := p.baseURL + "/" + version

	// List existing remote files reached by url.
	remoteFiles := listRemoteDir(url, p.listFilter)

	tmpDir, err := ioutil.TempDir("", p.dirPrefix)
	check(err)

	for _, remFile := range remoteFiles {
		if alreadyHasData(remFile) == true {
			continue
		}

		// Construct the file path to write the downloaded data.
		filepath := path.Join(tmpDir, path.Base(remFile))
		err = downloadFile(filepath, remFile)
		check(err)
		// Dont check error of unpack file due
		// some problems with shell commando to rename
		// linux files without version.
		// When there is a version
		// sometimes the command returns an error.
		versionName, err := unpackFile(filepath)
		check(err)

		digraph, err := createGraphFromCflowsOutput(path.Join(tmpDir, versionName))
		check(err)
		digraph.NameIt(versionName)
		fn := path.Join("data", versionName+GraphFormatExtension)
		gio.WritePajek(digraph, fn)

	}
}

func buildDataPath(filePrefix string) string {
	return path.Join(GraphDataRelativePath, filePrefix+
		GraphFormatExtension)
}

func alreadyHasData(fileURL string) bool {
	filePrefix := extractFilePrefixFromFileURL(fileURL)
	dataPath := buildDataPath(filePrefix)

	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		return false
	}
	log.Printf("Data file %s already exists.\n", dataPath)
	return true
}

func cleanTmpFiles() {
	// Clean the linux files
	cmd := exec.Command("/bin/bash", "-c", "rm -rf /tmp/linux*")
	err := cmd.Run()
	check(err)
}

func generateData(p *program) {
	for _, v := range p.versions {
		_GenerateData(p, v)
		cleanTmpFiles()
	}
}
