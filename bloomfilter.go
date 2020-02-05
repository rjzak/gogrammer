package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

const (
	DefaultBase float64 = 1.090878326190223750496427194244941608437246127397209913749
	DefaultLogBase float64 = 0.086983175599679411377848736810437843836925507056973208359
)

type CountingBloom struct {
	Base float64
	LogBase float64
	Added uint32
	Divisor int64
	HashSeeds []int32
	Counts []byte
}

func NewBloom(slots, hashFunctions int) CountingBloom {
	rand.Seed(time.Now().UnixNano())
	cb := CountingBloom{}
	cb.Base = DefaultBase
	cb.LogBase = DefaultLogBase
	cb.Added = 0
	cb.Divisor = 0
	cb.HashSeeds = make([]int32, hashFunctions)
	for i := 0; i < hashFunctions; i++ {
		cb.HashSeeds[i] = rand.Int31()
	}
	cb.Counts = make([]byte, slots)
	return cb
}

func LoadBlooms(filename string) (CountingBloom, error) {
	cb := CountingBloom{}
	bytesTemp := make([]byte, 8)
	var bytesRead int
	file, err := os.Open(filename)
	if err != nil {
		return cb, err
	}
	defer file.Close()

	err = binary.Read(file, binary.BigEndian, bytesTemp)
	if err != nil {
		return cb, err
	}
	cb.Base = math.Float64frombits(binary.BigEndian.Uint64(bytesTemp))

	err = binary.Read(file, binary.BigEndian, bytesTemp)
	if err != nil {
		return cb, err
	}
	cb.LogBase = math.Float64frombits(binary.BigEndian.Uint64(bytesTemp))

	err = binary.Read(file, binary.BigEndian, bytesTemp)
	if err != nil {
		return cb, err
	}
	cb.Added = binary.BigEndian.Uint32(bytesTemp)

	err = binary.Read(file, binary.BigEndian, bytesTemp)
	if err != nil {
		return cb, err
	}
	cb.Divisor, bytesRead = binary.Varint(bytesTemp)
	if bytesRead < 1 {
		return cb, errors.New("Couldn't read signed integer")
	}

	err = binary.Read(file, binary.BigEndian, bytesTemp)
	if err != nil {
		return cb, err
	}
	numHashSeeds := binary.BigEndian.Uint32(bytesTemp)
	cb.HashSeeds = make([]int32, numHashSeeds)

	for index := uint32(0); index < numHashSeeds; index++ {
		err = binary.Read(file, binary.BigEndian, bytesTemp)
		if err != nil {
			return cb, err
		}
		hashSeed, bytesRead := binary.Varint(bytesTemp)
		if bytesRead < 1 {
			return cb, errors.New("Couldn't read signed integer")
		}
		cb.HashSeeds[index] = int32(hashSeed)
	}

	err = binary.Read(file, binary.BigEndian, bytesTemp)
	if err != nil {
		return cb, err
	}
	numCounts := binary.BigEndian.Uint32(bytesTemp)
	cb.Counts = make([]byte, numCounts)

	err = binary.Read(file, binary.BigEndian, cb.Counts)
	if err != nil {
		return cb, err
	}

	return cb, nil
}

func (cb *CountingBloom) Put(item interface{}, raw_count int) error {
	byteVal, err := GetBytes(item)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to hash item: %v", err))
	}

	expo_count := math.Min(math.Ceil(math.Log(float64(raw_count)/cb.LogBase)), 255)
	//fmt.Printf("Expo Count: %v\n", expo_count)
	m := md5.New()
	init_hash := binary.LittleEndian.Uint64(m.Sum(byteVal))
	for i := 0; i < len(cb.HashSeeds); i++ {
		h := Hash6432shift( (uint64(cb.HashSeeds[i]) << 32) | init_hash)
		h = h % uint64(len(cb.Counts))
		temp := IntToByte(int(math.Max(float64(ByteToInt(cb.Counts[h])), float64(expo_count))))
		//fmt.Printf("Counts[%d] = %v\n", h, cb.Counts[h])
		cb.Counts[h] = temp
		//fmt.Printf("Counts[%d] = %v\n", h, cb.Counts[h])
	}

	cb.Added++;
	return nil
}

func (cb *CountingBloom) Get(item interface{}) (int, error) {
	byteVal, err := GetBytes(item)
	if err != nil {
		return -1, errors.New(fmt.Sprintf("Unable to hash item: %v", err))
	}

	m := md5.New()
	init_hash := binary.LittleEndian.Uint64(m.Sum(byteVal))

	var min_expo int = 257
	for i := 0; i < len(cb.HashSeeds); i++ {
		h := Hash6432shift( (uint64(cb.HashSeeds[i]) << 32) | init_hash)
		h = h % uint64(len(cb.Counts))
		min_expo = int(math.Min(float64(ByteToInt(cb.Counts[h])), float64(min_expo)))
	}

	if min_expo == 0 {
		return 0, nil
	}
	return int(math.Pow(cb.Base, float64(min_expo))), nil
}

func (cb *CountingBloom) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	bytesTemp := make([]byte, 8)

	binary.BigEndian.PutUint64(bytesTemp, math.Float64bits(cb.Base))
	err = binary.Write(file, binary.BigEndian, bytesTemp)
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint64(bytesTemp, math.Float64bits(cb.LogBase))
	err = binary.Write(file, binary.BigEndian, bytesTemp)
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint32(bytesTemp, cb.Added)
	err = binary.Write(file, binary.BigEndian, bytesTemp)

	binary.PutVarint(bytesTemp, cb.Divisor)
	err = binary.Write(file, binary.BigEndian, bytesTemp)
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint32(bytesTemp, uint32(len(cb.HashSeeds)))
	err = binary.Write(file, binary.BigEndian, bytesTemp)
	if err != nil {
		return err
	}

	for _, seed := range cb.HashSeeds {
		//binary.BigEndian.PutUint32(bytesTemp, uint32(seed))
		binary.PutVarint(bytesTemp, int64(seed))
		err = binary.Write(file, binary.BigEndian, bytesTemp)
		if err != nil {
			return err
		}
	}

	binary.BigEndian.PutUint32(bytesTemp, uint32(len(cb.Counts)))
	err = binary.Write(file, binary.BigEndian, bytesTemp)
	if err != nil {
		return err
	}

	err = binary.Write(file, binary.BigEndian, cb.Counts)
	if err != nil {
		return err
	}

	return nil
}

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func IntToByte(i int) byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(i))
	return bs[0]
}

func ByteToInt(b byte) int {
	val := binary.LittleEndian.Uint16([]byte{b, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	return int(val)
}

func Hash6432shift(key uint64) uint64 {
	key = (1^key) + (key << 18)
	key = key ^ (key >> 31)
	key = key * 21
	key = key ^ (key >> 11)
	key = key + (key << 6)
	key = key ^ (key >> 22)
	return key
}

func CalcSlotsHashes(numItems int, fpRate float64) (int, int) {
	bloom_slots := int(math.Ceil((float64(numItems) * math.Log(fpRate)) / math.Log(1 / math.Pow(2, math.Log(2)))))
	bloom_hashes := int(math.Round((float64(bloom_slots) / float64(numItems) * math.Log(2))))
	return bloom_slots, bloom_hashes
}