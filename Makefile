TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=registry.terraform.io
NAMESPACE=kota65535
NAME=alternator
BINARY=terraform-provider-${NAME}
VERSION=0.0.38
OS_ARCH=darwin_amd64

default: install

build:
	go build -o ${BINARY}

release:
	goreleaser release --rm-dist --snapshot --skip-publish  --skip-sign

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

compose-up:
	docker-compose up -d
	while ! (mysqladmin ping -h 127.0.0.1 -P 23306 -u root --silent); do sleep 5; done

test:
	go test -i $(TEST) || exit 1                                                   
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4                    

testacc: compose-up
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m   
