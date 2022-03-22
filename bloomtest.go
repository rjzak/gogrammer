package main

import (
	"fmt"
	"math/rand"
	"os"
)

func RandomString() string { // https://yourbasic.org/golang/generate-random-string/
	digits := "0123456789"
	specials := "~=+%^*/()[]{}/!@#$?|"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials
	length := 10
	buf := make([]byte, length)
	buf[0] = digits[rand.Intn(len(digits))]
	buf[1] = specials[rand.Intn(len(specials))]
	for i := 2; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}
	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})
	return string(buf) // E.g. "3i[g0|)z"
}

func TestBloomFilter(insertItems int, fpRate float64, iterations int, output string) {
	bloom_slots, bloom_hashes := CalcSlotsHashes(insertItems, fpRate)
	error_count := 0
	success_count := 0
	fmt.Printf("Slots: %d, Hashes: %d\n", bloom_slots, bloom_hashes)

	for iteration := 0; iteration < iterations; iteration++ {
		fmt.Printf("Iteration %d\n", iteration+1)
		bloom := NewBloom(bloom_slots, bloom_hashes)
		insertedItems := make([]string, 0)
		notInsertedItems := make([]string, 0)
		for item := 0; item < insertItems; item++ {
			temp := RandomString()
			insertedItems = append(insertedItems, temp)
			err := bloom.Put(temp, 1)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to insert %s into bloom filter.\n", temp)
				error_count += 1
			} else {
				success_count += 1
			}
		}
		for item := 0; item < insertItems*2; item++ {
			temp := RandomString()
			alreadySeen := false
			for _, inserted := range insertedItems {
				if inserted == temp {
					alreadySeen = true
					break
				}
			}
			if !alreadySeen {
				notInsertedItems = append(notInsertedItems, temp)
			}
		}

		for _, inserted := range insertedItems {
			itemCount, err := bloom.Get(inserted)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to retrieve inserted item %s from bloom filter.\n", inserted)
				error_count += 1
			} else {
				if itemCount > 0 {
					success_count += 1
				} else {
					fmt.Fprintf(os.Stderr, "Bloom filter reported that inserted item %s was not present.\n", inserted)
					error_count += 1
				}
			}
		}

		for _, notInserted := range notInsertedItems {
			itemCount, err := bloom.Get(notInserted)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to retrieve non-inserted item %s from bloom filter.\n", notInserted)
				error_count += 1
			} else {
				if itemCount > 0 {
					fmt.Fprintf(os.Stderr, "Bloom filter reported that non-inserted item %s was present %d times.\n", notInserted, itemCount)
					error_count += 1
				} else {
					success_count += 1
				}
			}
		}

		err := bloom.Save(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to serialize bloom filter.\n")
			error_count += 1
		} else {
			success_count += 1
			newBloom, err := LoadBlooms(output)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load serialized bloom filter.\n")
				error_count += 1
			} else {
				if newBloom.Added == bloom.Added {
					success_count += 1
				} else {
					fmt.Fprintf(os.Stderr, "Loaded bloom filter has a different added count.\n")
					error_count += 1
				}
				if newBloom.Base == bloom.Base {
					success_count += 1
				} else {
					fmt.Fprintf(os.Stderr, "Loaded bloom filter has a different base value.\n")
					error_count += 1
				}
				if newBloom.LogBase == bloom.LogBase {
					success_count += 1
				} else {
					fmt.Fprintf(os.Stderr, "Loaded bloom filter has a different log base value.\n")
					error_count += 1
				}
				if len(newBloom.HashSeeds) == len(bloom.HashSeeds) {
					success_count += 1
					for hashIndex := 0; hashIndex < len(bloom.HashSeeds); hashIndex++ {
						if newBloom.HashSeeds[hashIndex] == bloom.HashSeeds[hashIndex] {
							success_count += 1
						} else {
							fmt.Fprintf(os.Stderr, "Loaded bloom filter has different hash seed value for index %d.\n", hashIndex)
							error_count += 1
						}
					}
				} else {
					fmt.Fprintf(os.Stderr, "Loaded bloom filter has a different amount of hash functions.\n")
					error_count += 1
				}
				for _, notInserted := range notInsertedItems {
					itemCount, err := newBloom.Get(notInserted)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to retrieve non-inserted item %s from loaded bloom filter.\n", notInserted)
						error_count += 1
					} else {
						if itemCount > 0 {
							fmt.Fprintf(os.Stderr, "Loaded bloom filter reported that non-inserted item %s was present %d times.\n", notInserted, itemCount)
							error_count += 1
						} else {
							success_count += 1
						}
					}
				}
			}
		}
	}

	percentage := 100.0 * (float64(success_count) / float64(error_count+success_count))
	fmt.Printf("Success rate: %1.2f%%\n", percentage)
	fmt.Printf("Total errors: %d\n", error_count)
	fmt.Printf("Total successes: %d\n", success_count)
}
