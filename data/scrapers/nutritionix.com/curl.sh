#!/usr/bin/env zsh

scr_dir=$(dirname "$0")
data_dir="$scr_dir/../../ready/nutritionix.com"
mkdir -p "$data_dir"

id=`echo $1 | sed 's/[^0-9a-Z]*//g'`
f_name="$data_dir/$id.json"

if [ -f "$f_name" ]; then
  < "$f_name"
  exit 0
fi

curl $1 -s -o "$f_name" \
  -H 'authority: www.nutritionix.com' \
  -H 'accept: application/json, text/plain, */*' \
  -H 'accept-language: en-US,en;q=0.9,es-ES;q=0.8,es;q=0.7' \
  -H 'cache-control: no-cache' \
  -H 'cookie: AWSELB=2B0FF5CD0605B75731CF5A1A9C8743998EAB3A76261ED56F8EE8DFDFBA7CE8465E3C7D313E3AA9D0C41199E174133837928EDB9E7483E43C9E770784A84807D272FF8F19D0; AWSELBCORS=2B0FF5CD0605B75731CF5A1A9C8743998EAB3A76261ED56F8EE8DFDFBA7CE8465E3C7D313E3AA9D0C41199E174133837928EDB9E7483E43C9E770784A84807D272FF8F19D0' \
  -H 'pragma: no-cache' \
  -H 'referer: https://www.nutritionix.com/es/database/common-foods' \
  -H 'sec-ch-ua: "Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Linux"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-origin' \
  -H 'user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36' \
  --compressed

< "$f_name"
exit 0