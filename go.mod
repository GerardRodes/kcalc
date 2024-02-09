module github.com/GerardRodes/kcalc

go 1.22.0

replace github.com/eatonphil/gosqlite => ../../repos/gosqlite

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/eatonphil/gosqlite v0.9.0
	github.com/gertd/go-pluralize v0.2.1
	github.com/jxskiss/base62 v1.1.0
	github.com/kolesa-team/go-webp v1.0.4
	github.com/mitchellh/go-server-timing v1.0.1
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/rs/zerolog v1.32.0
	github.com/segmentio/ksuid v1.0.4
	go.uber.org/automaxprocs v1.5.3
	golang.org/x/sync v0.6.0
	golang.org/x/text v0.14.0
)

require (
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/golang/gddo v0.0.0-20210115222349-20d68f94ee1f // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.17.0 // indirect
)
