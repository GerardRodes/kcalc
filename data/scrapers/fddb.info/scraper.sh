#!/usr/bin/env zsh

./curl.sh \
	https://fddb.info/db/en/groups/catalogue/index.html \
	| htmlq \
		--attribute="href" \
		'#content > div.mainblock > div.leftblock > div > div > div > table > tbody > tr > td > table > tbody > tr > td:nth-child(3) > a'