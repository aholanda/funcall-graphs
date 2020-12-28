scc.pdf: scc.dat
	gnuplot $<
TRASH += scc.pdf scc.dat

scc.dat: %.go
	@go build
	@./funcall-graphs
TRASH += funcall-graphs

cleandata:
	$(RM) data/.net data/*dontuse*.net data/*pre*.net

clean:
	$(RM) $(TRASH)

TIDY: clean
	$(RM) data/*.net

.PHONY: clean cleandata
