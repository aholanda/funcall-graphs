set terminal pdf
set output 'scc.pdf'

set style function linespoints
set style line 1 \
    linecolor rgb '#0060ad' \
    linetype 1 linewidth 2 \
    pointtype 7 pointsize .2

set style line 2 \
    linecolor rgb '#dd181f' \
    linetype 1 linewidth 2 \
    pointtype 5 pointsize .2
set ytics 5 nomirror tc lt 1
#set ylabel 'avg degree' tc lt 1
set y2tics 15 nomirror tc lt 2
#set y2label 'largest component size' tc lt 2

plot 'scc.dat' using 4:xticlabels(1) ls 1, \
     ''        using 7:xticlabels(1) ls 2 axes x1y2