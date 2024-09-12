#!/bin/sh

for path_to in data/*.yaml
do
    filename="$(basename $path_to)"
    name="${filename%.*}"
    # output as html (base)
    go run . data/$filename > out/resume_$name.html
    # output as pdf via headless Chromium
    chromium --headless --disable-gpu --run-all-compositor-stages-before-draw --no-pdf-header-footer --print-to-pdf="out/resume_$name.pdf" out/resume_$name.html
done

cp out/resume_main.pdf out/resume.pdf