#!/bin/bash
set -e

# IPADIC_PATH can be overridden by the caller; if not set, auto-detect the
# top-level directory that was extracted from the tarball.
if [ -z "$IPADIC_PATH" ]; then
    IPADIC_PATH=$(find . -maxdepth 1 -type d -name 'mecab-ipadic-*' | head -1)
fi

if [ -z "$IPADIC_PATH" ] || [ ! -d "$IPADIC_PATH" ]; then
    echo "ERROR: IPADIC directory not found (looked for mecab-ipadic-* under $(pwd))." >&2
    echo "       Extracted directories:" >&2
    ls -d */ 2>/dev/null >&2 || true
    exit 1
fi

echo "Using IPADIC directory: $IPADIC_PATH"

REQUIRED_CSVS="
Adj.csv
Adverb.csv
Noun.adjv.csv
Noun.adverbal.csv
Noun.csv
Noun.nai.csv
Noun.verbal.csv
Verb.csv
"

# "-p" means "make dir if it doesn't exist"
mkdir -p $PATH_TO

# copy csv files to db/csv and convert them to utf-8
for csv in $REQUIRED_CSVS; do
    iconv $IPADIC_PATH/$csv -f EUC-JP -t UTF-8 -o $PATH_TO/$csv
    echo convert $csv to utf-8
done