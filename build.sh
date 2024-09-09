#!/bin/sh

# output as html (base)
go run . data/main.yaml > resume.html

# output as pdf via headless Chromium
chromium --headless --disable-gpu --run-all-compositor-stages-before-draw --no-pdf-header-footer --print-to-pdf="resume.pdf" resume.html
