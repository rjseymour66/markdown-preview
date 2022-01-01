#! /bin/bash

# Receives the name of the file that you want to preview as an argument
# and calculates its checksum every 5s. If the checksum is different,
# the markdown-preview tool executed

# backticks replace vars and execute the command 
FHASH=`md5sum $1`
while true; do
	NHASH=`md5sum $1`
	if [ "$NHASH" != "$FHASH" ]; then
		./markdown-preview -file $1
		FHASH=$NHASH
	fi
	sleep 5
done
