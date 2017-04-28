#!/bin/bash 

if [ $# -eq 0 ]; then
	echo "csv file required"
fi

for f in $@; do
	nkf -w80 $f >  ${f/.csv/.utf}
done
