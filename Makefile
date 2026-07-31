.PHONY: build test test-coverage lint lint-fix wasm serve-wasm

build:
	go build -o koikoi .

# ブラウザ版 (GitHub Pages 用) を dist/ に生成
wasm:
	rm -rf dist
	mkdir -p dist
	cp web/index.html web/tcell.js web/termstyle.css web/beep.wav dist/
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" dist/
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o dist/main.wasm .

serve-wasm: wasm
	go run ./web -dir dist -addr localhost:8080

test:
	go test -count=1 -timeout 30s ./...

test-coverage:
	go test -coverprofile=coverage.out -count=1 -timeout 30s ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix
