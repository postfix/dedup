package dedup_test

import (
	"fmt"
	"testing"

	"bytes"
	"io"
	"math/rand"

	"io/ioutil"

	"github.com/klauspost/dedup"
)

// Returns a deterministic buffer of size n
func getBufferSize(n int) *bytes.Buffer {
	rand.Seed(0)
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(rand.Intn(255))
	}
	return bytes.NewBuffer(b)
}

func TestFixedWriter(t *testing.T) {
	idx := bytes.Buffer{}
	data := bytes.Buffer{}

	const totalinput = 10 << 20
	input := getBufferSize(totalinput)

	const size = 64 << 10
	b := input.Bytes()
	// Create some duplicates
	for i := 0; i < 50; i++ {
		// Read from 10 first blocks
		src := b[(i%10)*size : (i%10)*size+size]
		// Write into the following ones
		dst := b[(10+i)*size : (i+10)*size+size]
		copy(dst, src)
	}
	input = bytes.NewBuffer(b)
	w, err := dedup.NewWriter(&idx, &data, dedup.ModeFixed, size, size*10)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(w, input)
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	removed := ((totalinput) - data.Len()) / size

	t.Log(dedup.BirthdayProblem(totalinput / size))
	t.Log("Index size:", idx.Len())
	t.Log("Data size:", data.Len())
	t.Log("Removed", removed, "blocks")
	// We should get at least 50 blocks
	if removed < 50 {
		t.Fatal("didn't remove at least 50 blocks")
	}
	if removed > 60 {
		t.Fatal("removed unreasonable high amount of blocks")
	}
}

func TestFixedWriterLimit(t *testing.T) {
	idx := bytes.Buffer{}
	data := bytes.Buffer{}

	const totalinput = 10 << 20
	const limit = 9
	input := getBufferSize(totalinput)

	const size = 64 << 10
	b := input.Bytes()
	// Create some duplicates
	for i := 0; i < 50; i++ {
		// Read from 10 first blocks
		src := b[(i%10)*size : (i%10)*size+size]
		// Write into the following ones
		dst := b[(10+50-i)*size : (10+50-i)*size+size]
		copy(dst, src)
	}
	input = bytes.NewBuffer(b)
	w, err := dedup.NewWriter(&idx, &data, dedup.ModeFixed, size, limit*size)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(w, input)
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	removed := ((totalinput) - data.Len()) / size

	t.Log("Index size:", idx.Len())
	t.Log("Data size:", data.Len())
	t.Log("Removed", removed, "blocks")
	// We should get at least 50 blocks
	if removed > 10 {
		t.Fatal("it did not appear to respect the limit")
	}
	if removed < 8 {
		t.Fatal("removed too many blocks")
	}
	r, err := dedup.NewReader(&idx, &data)
	if err != nil {
		t.Fatal(err)
	}

	useBlocks := r.MaxMem() / size
	if useBlocks > 9 {
		t.Fatal("Uses too much memory, expected", limit, "got", useBlocks)
	}
	t.Log("Maximum estimated use:", r.MaxMem(), "bytes,", useBlocks, "blocks")
	r.Close()
}

func TestDynamicWriter(t *testing.T) {
	idx := bytes.Buffer{}
	data := bytes.Buffer{}

	const totalinput = 10 << 20
	input := getBufferSize(totalinput)

	const size = 64 << 10
	b := input.Bytes()
	// Create some duplicates
	for i := 0; i < 50; i++ {
		// Read from 10 first blocks
		src := b[(i%10)*size : (i%10)*size+size]
		// Write into the following ones
		dst := b[(10+i)*size : (i+10)*size+size]
		copy(dst, src)
	}
	input = bytes.NewBuffer(b)
	w, err := dedup.NewWriter(&idx, &data, dedup.ModeDynamic, size, 10*8*size)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(w, input)
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	removed := ((totalinput) - data.Len()) / size

	t.Log("Dynamic Index size:", idx.Len())
	t.Log("Dynamic Data size:", data.Len())
	t.Log("Removed", removed, "blocks")
	// We don't know how many, but it should remove some blocks
	if removed < 40 {
		t.Fatal("didn't remove at least 40 blocks")
	}
}

func TestFixedStreamWriter(t *testing.T) {
	data := bytes.Buffer{}

	const totalinput = 10 << 20
	input := getBufferSize(totalinput)

	const size = 64 << 10
	b := input.Bytes()
	// Create some duplicates
	for i := 0; i < 50; i++ {
		// Read from 10 first blocks
		src := b[(i%10)*size : (i%10)*size+size]
		// Write into the following ones
		dst := b[(10+i)*size : (i+10)*size+size]
		copy(dst, src)
	}
	input = bytes.NewBuffer(b)
	w, err := dedup.NewStreamWriter(&data, dedup.ModeFixed, size, 10*size)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(w, input)
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	removed := ((totalinput) - data.Len()) / size

	t.Log("Data size:", data.Len())
	t.Log("Removed", removed, "blocks")
	// We should get at least 50 blocks, but there is a little overhead
	if removed < 49 {
		t.Fatal("didn't remove at least 49 blocks")
	}
	if removed > 60 {
		t.Fatal("removed unreasonable high amount of blocks")
	}
}

func TestDynamicStreamWriter(t *testing.T) {
	data := bytes.Buffer{}

	const totalinput = 10 << 20
	input := getBufferSize(totalinput)

	const size = 64 << 10
	b := input.Bytes()
	// Create some duplicates
	for i := 0; i < 50; i++ {
		// Read from 10 first blocks
		src := b[(i%10)*size : (i%10)*size+size]
		// Write into the following ones
		dst := b[(10+i)*size : (i+10)*size+size]
		copy(dst, src)
	}
	input = bytes.NewBuffer(b)
	w, err := dedup.NewStreamWriter(&data, dedup.ModeDynamic, size, 10*8*size)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(w, input)
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	removed := ((totalinput) - data.Len()) / size

	t.Log("Dynamic Data size:", data.Len())
	t.Log("Removed", removed, "blocks")
	// We don't know how many, but it should remove some blocks
	if removed < 40 {
		t.Fatal("didn't remove at least 40 blocks")
	}
}

func BenchmarkFixedWriter64K(t *testing.B) {
	const totalinput = 10 << 20
	input := getBufferSize(totalinput)

	const size = 64 << 10
	b := input.Bytes()
	// Create some duplicates
	for i := 0; i < 50; i++ {
		// Read from 10 first blocks
		src := b[(i%10)*size : (i%10)*size+size]
		// Write into the following ones
		dst := b[(10+i)*size : (i+10)*size+size]
		copy(dst, src)
	}
	t.ResetTimer()
	t.SetBytes(totalinput)
	for i := 0; i < t.N; i++ {
		input = bytes.NewBuffer(b)
		w, _ := dedup.NewWriter(ioutil.Discard, ioutil.Discard, dedup.ModeFixed, size, 0)
		io.Copy(w, input)
		err := w.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func BenchmarkFixedWriter4K(t *testing.B) {
	const totalinput = 10 << 20
	input := getBufferSize(totalinput)

	const size = 4 << 10
	b := input.Bytes()
	// Create some duplicates
	for i := 0; i < 500; i++ {
		// Read from 10 first blocks
		src := b[(i%10)*size : (i%10)*size+size]
		// Write into the following ones
		dst := b[(10+i)*size : (i+10)*size+size]
		copy(dst, src)
	}
	t.ResetTimer()
	t.SetBytes(totalinput)
	for i := 0; i < t.N; i++ {
		input = bytes.NewBuffer(b)
		w, _ := dedup.NewWriter(ioutil.Discard, ioutil.Discard, dedup.ModeFixed, size, 0)
		io.Copy(w, input)
		err := w.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func BenchmarkFixedWriter1K(t *testing.B) {
	const totalinput = 10 << 20
	input := getBufferSize(totalinput)

	const size = 1 << 10
	b := input.Bytes()
	// Create some duplicates
	for i := 0; i < 500; i++ {
		// Read from 10 first blocks
		src := b[(i%10)*size : (i%10)*size+size]
		// Write into the following ones
		dst := b[(10+i)*size : (i+10)*size+size]
		copy(dst, src)
	}
	t.ResetTimer()
	t.SetBytes(totalinput)
	for i := 0; i < t.N; i++ {
		input = bytes.NewBuffer(b)
		w, _ := dedup.NewWriter(ioutil.Discard, ioutil.Discard, dedup.ModeFixed, size, 0)
		io.Copy(w, input)
		err := w.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

// Maximum block size:64k
func BenchmarkDynamicWriter64K(t *testing.B) {
	const totalinput = 10 << 20
	input := getBufferSize(totalinput)

	const size = 64 << 10
	b := input.Bytes()
	// Create some duplicates
	for i := 0; i < 50; i++ {
		// Read from 10 first blocks
		src := b[(i%10)*size : (i%10)*size+size]
		// Write into the following ones
		dst := b[(10+i)*size : (i+10)*size+size]
		copy(dst, src)
	}
	t.ResetTimer()
	t.SetBytes(totalinput)
	for i := 0; i < t.N; i++ {
		input = bytes.NewBuffer(b)
		w, _ := dedup.NewWriter(ioutil.Discard, ioutil.Discard, dedup.ModeDynamic, size, 0)
		io.Copy(w, input)
		err := w.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

// Maximum block size:4k
func BenchmarkDynamicWriter4K(t *testing.B) {
	const totalinput = 10 << 20
	input := getBufferSize(totalinput)

	const size = 4 << 10
	b := input.Bytes()
	// Create some duplicates
	for i := 0; i < 50; i++ {
		// Read from 10 first blocks
		src := b[(i%10)*size : (i%10)*size+size]
		// Write into the following ones
		dst := b[(10+i)*size : (i+10)*size+size]
		copy(dst, src)
	}
	t.ResetTimer()
	t.SetBytes(totalinput)
	for i := 0; i < t.N; i++ {
		input = bytes.NewBuffer(b)
		w, _ := dedup.NewWriter(ioutil.Discard, ioutil.Discard, dedup.ModeDynamic, size, 0)
		io.Copy(w, input)
		err := w.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func BenchmarkFixedStreamWriter4K(t *testing.B) {
	const totalinput = 10 << 20
	input := getBufferSize(totalinput)

	const size = 4 << 10
	b := input.Bytes()
	// Create some duplicates
	for i := 0; i < 500; i++ {
		// Read from 10 first blocks
		src := b[(i%10)*size : (i%10)*size+size]
		// Write into the following ones
		dst := b[(10+i)*size : (i+10)*size+size]
		copy(dst, src)
	}
	t.ResetTimer()
	t.SetBytes(totalinput)
	for i := 0; i < t.N; i++ {
		input = bytes.NewBuffer(b)
		w, _ := dedup.NewStreamWriter(ioutil.Discard, dedup.ModeFixed, size, 10*size)
		io.Copy(w, input)
		err := w.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

// This doesn't actually test anything, but prints probabilities to log
func TestBirthdayProblem(t *testing.T) {
	t.Log("Hash size is", dedup.HashSize*8, "bits")
	t.Log("1GiB, 1KiB blocks:")
	t.Log(dedup.BirthdayProblem((1 << 30) / (1 << 10)))
	w, _ := dedup.NewWriter(ioutil.Discard, ioutil.Discard, dedup.ModeFixed, 1<<10, 0)
	e, _ := w.MemUse(1 << 30)
	t.Logf("It will use %d MiB for encoder.", e>>20)

	t.Log("1TiB, 4KiB blocks:")
	t.Log(dedup.BirthdayProblem((1 << 40) / (4 << 10)))
	w, _ = dedup.NewWriter(ioutil.Discard, ioutil.Discard, dedup.ModeFixed, 4<<10, 0)
	e, _ = w.MemUse(1 << 40)
	t.Logf("It will use %d MiB for encoder.", e>>20)

	t.Log("1PiB, 4KiB blocks:")
	t.Log(dedup.BirthdayProblem((1 << 50) / (4 << 10)))
	e, _ = w.MemUse(1 << 50)
	t.Logf("It will use %d MiB for encoder.", e>>20)

	t.Log("1EiB, 64KiB blocks:")
	t.Log(dedup.BirthdayProblem((1 << 60) / (64 << 10)))
	w, _ = dedup.NewWriter(ioutil.Discard, ioutil.Discard, dedup.ModeFixed, 64<<10, 0)
	e, _ = w.MemUse(1 << 60)
	t.Logf("It will use %d MiB for encoder.", e>>20)

	t.Log("1EiB, 1KiB blocks:")
	t.Log(dedup.BirthdayProblem((1 << 60) / (1 << 10)))
	w, _ = dedup.NewWriter(ioutil.Discard, ioutil.Discard, dedup.ModeFixed, 1<<10, 0)
	e, _ = w.MemUse(1 << 60)
	t.Logf("It will use %d MiB for encoder.", e>>20)
}

// This will deduplicate a buffer of zeros to an indexed stream
func ExampleNewWriter() {
	// We will write to these
	idx := bytes.Buffer{}
	data := bytes.Buffer{}

	// This is our input:
	input := bytes.NewBuffer(make([]byte, 50000))

	// Create a new writer, with each block being 1000 bytes
	w, err := dedup.NewWriter(&idx, &data, dedup.ModeFixed, 1000, 0)
	if err != nil {
		panic(err)
	}
	// Copy our input to the writer.
	io.Copy(w, input)

	// Close the writer
	err = w.Close()
	if err != nil {
		panic(err)
	}

	// Let us inspect what was written:
	fmt.Println("Blocks:", w.Blocks())
	fmt.Println("Index size:", idx.Len())
	fmt.Println("Data size:", data.Len())

	// OUTPUT: Blocks: 50
	// Index size: 66
	// Data size: 1000
}

// This will deduplicate a buffer of zeros to an non-indexed stream
func ExampleNewStreamWriter() {
	// We will write to this
	data := bytes.Buffer{}

	// This is our input:
	input := bytes.NewBuffer(make([]byte, 50000))

	// Create a new writer, with each block being 1000 bytes,
	// And allow it to use 10000 bytes of memory
	w, err := dedup.NewStreamWriter(&data, dedup.ModeFixed, 1000, 10000)
	if err != nil {
		panic(err)
	}
	// Copy our input to the writer.
	io.Copy(w, input)

	// Close the writer
	err = w.Close()
	if err != nil {
		panic(err)
	}

	// Let us inspect what was written:
	fmt.Println("Blocks:", w.Blocks())
	fmt.Println("Data size:", data.Len())

	// OUTPUT: Blocks: 50
	// Data size: 1067
}
