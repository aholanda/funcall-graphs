scc.pdf: scc.dat
	gnuplot $<
TRASH += scc.pdf scc.dat

funcall-graphs: main.go compute_components.go generate_data.go
	@go build
	@go test

scc.dat: funcall-graphs
	./funcall-graphs
TRASH += funcall-graphs

cleandata:
	$(RM) data/.net data/*dontuse*.net data/*pre*.net

clean:
	$(RM) $(TRASH) `find . -name *~`

TIDY: clean
	$(RM) data/*.net

.PHONY: clean cleandata
