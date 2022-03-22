package main

import "testing"

const SLOTS int = 100_000
const HASH_FUNCS int = 20

// Ensure the bloom filter contains what was added
func TestBloomFilterInsertRetrieveWorks(t *testing.T) {
	bloom := NewBloom(SLOTS, HASH_FUNCS)

	random_string := RandomString()
	err := bloom.Put(random_string, 1)
	if err != nil {
		t.Errorf("Error inserting random string: %s\n", err)
		return
	}

	count, err := bloom.Get(random_string)
	if err != nil {
		t.Errorf("Error retrieving random string: %s\n", err)
		return
	}
	if count != 1 {
		t.Errorf("Error received incorrect count value: 1 vs. %d\n", count)
		return
	}
}

// Ensure the bloom filter contains what was added
func TestBloomFilterCountRetrieveWorks(t *testing.T) {
	bloom := NewBloom(SLOTS, HASH_FUNCS)
	someInt := 99
	random_string := RandomString()
	err := bloom.Put(random_string, someInt)
	if err != nil {
		t.Errorf("Error inserting random string: %s\n", err)
		return
	}

	count, err := bloom.Get(random_string)
	if err != nil {
		t.Errorf("Error retrieving random string: %s\n", err)
		return
	}
	if count != someInt {
		t.Errorf("Error received incorrect count value: %d vs. %d\n", someInt, count)
		return
	}
}

// Ensure the bloom filter doesn't contain what wasn't added
func TestBloomFilterRetrieveUniqueWorks(t *testing.T) {
	bloom := NewBloom(SLOTS, HASH_FUNCS)

	random_string := RandomString()
	err := bloom.Put(random_string, 1)
	if err != nil {
		t.Errorf("Error inserting random string: %s\n", err)
		return
	}

	random_string = RandomString()
	count, err := bloom.Get(random_string)
	if err != nil {
		t.Errorf("Error retrieving random string: %s\n", err)
		return
	}
	if count != 0 {
		t.Errorf("Error received incorrect count value: 0 vs. %d\n", count)
		return
	}
}
