Gogrammer, a work-in-progress n-gram application written in Go.

This application creates a list of byte ngrams from a collection of files. There is an optional `-hash` flag which attempts to use a hashing algorithm, which should make larger values of `N` possible. It's inspired by this paper: *Raff, E., & Nicholas, C. K. (2018). "Hash-Grams: Faster N-Gram Features for Classification and Malware Detection"*, to appear in *Document Engineering*, available [here](https://www.edwardraff.com/publications/hash-grams-faster.pdf).

The hash method seems to have about 60% of the resulting n-grams in common with normal ngramming, which could be attributed to hash collisions. However, the hash method seems to run in about a quarter of the time.

The ability to train based on a created dataset is new, but requires running with `GODEBUG=cgocheck=0` to work, due to [this bug in golearn](https://github.com/sjwhitworth/golearn/issues/158).

Dependency installation:
* `go get github.com/aclements/go-rabin`
* `go get go get github.com/gonum/blas`
* `go get github.com/sjwhitworth/golearn`