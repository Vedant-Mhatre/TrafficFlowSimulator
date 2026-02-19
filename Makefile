APP := trafficsim

.PHONY: build run compare rush test vet clean

build:
	go build -o $(APP) ./cmd/trafficsim

run:
	go run ./cmd/trafficsim -config configs/baseline.json

compare:
	go run ./cmd/trafficsim -compare configs/baseline.json,configs/improved.json

rush:
	go run ./cmd/trafficsim -config configs/rush-hour.json -no-render

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f $(APP) coverage.out
