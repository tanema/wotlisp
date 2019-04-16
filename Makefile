all:
	go build -o ./build/wot ./cmd/wot

rep: all
	@./build/wot

test: all
	@./test/runtest.py ./test/final.mal -- ./build/wot

host: all
	@./build/wot ./test/mal/runtime.mal

perf: all
	@echo 'Running: ./build/wot ./tests/perf1.mal'
	@./build/wot ./tests/perf1.mal
	@echo 'Running: ./build/wot ./tests/perf2.mal'
	@./build/wot ./tests/perf2.mal
	@echo 'Running: ./build/wot ./tests/perf3.mal'
	@./build/wot ./tests/perf3.mal

clean:
	@rm -rf ./wotlisp/build
