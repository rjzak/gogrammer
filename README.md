## Gogrammer, a work-in-progress n-gram application written in Go.

This application creates a list of byte ngrams from a collection of files. There is an optional `-hash` flag which attempts to use a hashing algorithm, which should make larger values of `N` possible. It's inspired by this paper: *Raff, E., & Nicholas, C. K. (2018). "Hash-Grams: Faster N-Gram Features for Classification and Malware Detection"*, available [here](https://www.edwardraff.com/publications/hash-grams-faster.pdf).

The hash method seems to have about 60% of the resulting n-grams in common with normal ngramming, which could be attributed to hash collisions. However, the hash method seems to run in about a quarter of the time.

The ability to train based on a created dataset is new, but requires running with `GODEBUG=cgocheck=0` to work, due to [this bug in golearn](https://github.com/sjwhitworth/golearn/issues/158).

## Dependencies:
* [go-rabin](https://www.github.com/aclements/go-rabin)
* [golearn](https://www.github.com/sjwhitworth/golearn)

## How to use it:
1. Find the n-grams in your dataset: `./gogrammer NGRAM /path/to/goodware /path/to/malware`. Additional options are available, including changing the number of n-grams to keep, and the size of the n-grams.
2. Build a CSV or LibSVM dataset file based on the n-grams: `./gogrammer DATASET -goodware /path/to/goodware -malware /path/to/malware -kl output.grams`.
3. Train the model: `GODEBUG=cgocheck=0 ./gogrammer TRAIN -hasFlags -dataset dataset.csv -output my-model.model`. Additional options are available. The model is saved using [liblinear](https://github.com/cjlin1/liblinear) 's format.

