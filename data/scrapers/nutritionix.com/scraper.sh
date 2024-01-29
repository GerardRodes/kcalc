#!/usr/bin/env zsh

scr_dir=$(dirname "$0")
curl_sh="$scr_dir/curl.sh"

langs=("es_ES" "en_US" "en_GB") # "es_MX"

# Iterate over each string in the array
for lang in "${langs[@]}"; do
	echo "$lang >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
	data=$($curl_sh "https://www.nutritionix.com/nixapi/search/$lang?page=30")

	total=$(echo $data | jq '.total')
	per_page=$(echo $data | jq '.foods | length')
	pages=$((($total/$per_page)+1))

	echo "total:    $total"
	echo "per_page: $per_page"
	echo "pages:    $pages"

	seq 1 "$pages" \
	| parallel --memfree 1G -j0 --retries 3 \
		noglob $curl_sh "https://www.nutritionix.com/nixapi/search/$lang?page={}" 1>/dev/null

	echo "$lang <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
done
