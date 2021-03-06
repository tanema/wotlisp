# Wotlisp

Wotlisp is a small closure inspired implementation of lisp done by following the
[mal tutorial](https://github.com/kanaka/mal)

All tests and examples are from the `kanaka/mal` repository and is their work.

To try it out
- run `make rep` to build and run repl
- run `make test` to run all tests
- run `make perf` to run perf tests
- run `make host` to run a self hosted version

To run the examples you can run
- `make`
- `cd examples`
- `../build/wot ./examplename.mal`

## Perf output to compare with other implementations

```
> make perf
Running: ./build/wot ./test/perf1.mal
"Elapsed time: 0 msecs"
Running: ./build/wot ./test/perf2.mal
"Elapsed time: 1 msecs"
Running: ./build/wot ./test/perf3.mal
iters over 10 seconds: 30129
```
