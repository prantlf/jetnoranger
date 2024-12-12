all:: lint build

lint::
	cd success && go vet success.go
	cd failure && go vet failure.go

format::
	cd success && go fmt success.go
	cd failure && go fmt failure.go

build::
	cd success && go build success.go
	cd failure && go build failure.go
