#!/bin/sh
KEYS=$(jq -r 'keys[]' i18n/translations.en.json)

for file in $(find i18n/ -name "translations.*.json" -not -name "translations.en.json"); do
    FOUND=$(jq -r 'keys[]' $file)
    echo "Checking file: $file";
    cp $file $file.tmp;
    for key in $KEYS; do
        if [[ ! " ${FOUND[*]} " =~ [[:space:]]${key}[[:space:]] ]]; then
            val=$(jq -r ".$key" i18n/translations.en.json)
            echo "Adding missing key '$key' = '$val' to $file";
            jq ".$key=\"$val\"" $file.tmp > $file.2.tmp && mv $file.2.tmp $file.tmp
        fi
    done
    mv $file.tmp $file;
done