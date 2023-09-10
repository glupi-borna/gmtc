#!/bin/bash

while true; do
    # REALLY clear the screen.
    printf "\033c"
    bash ./build.sh
    # ./gmtc -path tests/scribble/Scribble.yyp
    # ./gmtc -path tests/test1.gml
    ./gmtc -path tests/scribble/scripts/UnicodeToKrutidev/UnicodeToKrutidev.gml
    inotifywait -e modify -e move -e create -e delete -q --recursive --include ".+go" .
done
