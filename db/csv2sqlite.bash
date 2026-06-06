set -e

mkdir -p $SQL_PATH

# Verify that all required CSV files are non-empty before importing.
for csv in $CSV_PATH/adj.csv $CSV_PATH/adverb.csv $CSV_PATH/noun.csv $CSV_PATH/verb.csv; do
    if [ ! -s "$csv" ]; then
        echo "ERROR: CSV file is missing or empty: $csv" >&2
        exit 1
    fi
done

# create sqlite file
# NOTE: sqlite can receive string(heredoc) as manipulation
sqlite3 $SQL_PATH/$SQL_NAME << EOS
/* create tables */
.read ./db/init.sql

/* let col separator "," */
.separator ','

/* import word data csvs */
.mode csv
.import $CSV_PATH/adj.csv adjectives
.import $CSV_PATH/adverb.csv adverbs
.import $CSV_PATH/noun.csv nouns
.import $CSV_PATH/verb.csv verbs

/* insert each table count into counts */
insert into counts select 'adjectives', count(*) from adjectives;
insert into counts select 'adverbs', count(*) from adverbs;
insert into counts select 'nouns', count(*) from nouns;
insert into counts select 'verbs', count(*) from verbs;
EOS

echo created $SQL_NAME

# Verify that all word tables have data after import.
for table in adjectives adverbs nouns verbs; do
    count=$(sqlite3 $SQL_PATH/$SQL_NAME "SELECT row_count FROM counts WHERE table_name='$table';")
    if [ -z "$count" ] || [ "$count" -eq 0 ]; then
        echo "ERROR: table '$table' has no data after import." >&2
        exit 1
    fi
    echo "  $table: $count rows"
done
