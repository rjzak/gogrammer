Gogrammer, an n-gram application written in Go.

This application creates a list of byte ngrams from a collection of files. There is an optional `-hash` flag which attempts to use a hashing algorithm, which should make larger values of `N` possible. It's inspired by this paper: Raff, E., & Nicholas, C. K. (2018). "Hash-Grams: Faster N-Gram Features for Classification and Malware Detection", to appear in *Document Engineering* [link](https://www.edwardraff.com/publications/hash-grams-faster.pdf).

This is very much a work-in-progress, don't use it as-is for anything important.

The hash method seems to have about 60% of the resulting n-grams in comming with normal ngramming, which could be attributed to hash collisions. However, the hash method seems to run in about a quarter of the time.