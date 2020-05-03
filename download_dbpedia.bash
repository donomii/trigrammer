#!/bin/bash

function download_dbpedia {
curl -s -o- $URL | bunzip2 | perl -pe 's/\t/ /g' |  perl -pe 's/@[^@]+\.$//g' | perl -pe 's/\^\^[^^]+\.$//g' | perl -pe 's/> </\t/g' | perl -pe 's/^<//g' | perl -pe 's/> "/\t"/g' | perl -pe 's/> \.$//g' | perl -pe 's!http://dbpedia.org/resource/!!g' | perl -pe 's!http://dbpedia.org/property/!!g' | perl -pe 's/"//g' | perl -pe 's/\s+$/\n/g' | ./trigrammr-import-csv --tsv --no_columns $DBNAME.sqlite
}

echo Starting to download 3 wikipedia extracts and import them as trigrams
echo
echo I hope you weren't planning to use your computer for a while...

URL=http://downloads.dbpedia.org/2016-04/core-i18n/en/infobox_properties_en.ttl.bz2 DBNAME=infobox download_dbpedia &

URL=http://downloads.dbpedia.org/2016-04/core-i18n/en/geonames_links_en.ttl.bz2 DBNAME=geonames download_dbpedia &

URL=http://downloads.dbpedia.org/2016-04/core-i18n/en/short_abstracts_en.ttl.bz2 DBNAME=abstracts download_dbpedia &

wait

echo You can now try out your new databases by typing:
echo
echo ./trigrammr-client *sqlite
echo
