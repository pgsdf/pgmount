#!/bin/sh
# Generate man pages from Markdown files

set -e

DOC_DIR="$(dirname "$0")"
OUTPUT_DIR="${DOC_DIR}/man"

mkdir -p "${OUTPUT_DIR}"

# Check if pandoc is available
if ! command -v pandoc >/dev/null 2>&1; then
    echo "Warning: pandoc not found. Man pages will not be generated."
    echo "Install pandoc: pkg install hs-pandoc"
    exit 1
fi

echo "Generating man pages..."

for md_file in "${DOC_DIR}"/*.md; do
    if [ -f "$md_file" ]; then
        basename=$(basename "$md_file" .md)
        man_file="${OUTPUT_DIR}/${basename}"
        
        echo "  ${basename}.md -> ${basename}"
        pandoc -s -t man "$md_file" -o "$man_file"
    fi
done

echo "Man pages generated in ${OUTPUT_DIR}/"
echo ""
echo "To install manually:"
echo "  sudo cp ${OUTPUT_DIR}/*.8 /usr/local/man/man8/"
echo "  sudo makewhatis /usr/local/man"
