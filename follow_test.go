package follow

import (
	"io"
	"os"
	"testing"
	"time"
)

const fileSize = 4096

func read(t *testing.T, fn string, ch chan<- int) {
	f, err := os.Open(fn)
	if err != nil {
		t.Errorf("Error opening tmp file: %v", err)
		return
	}
	defer f.Close()
	tailer := New(f)

	b := make([]byte, fileSize)
	n, err := io.ReadFull(tailer, b)
	if err != nil {
		t.Errorf("Error reading stuff: %v", err)
		return
	}

	tailer.Close()

	n2, err := io.ReadFull(tailer, b)
	if n2 != 0 {
		t.Errorf("Expected 0 bytes read after closing, got %v", n2)
	}
	if err != io.EOF {
		t.Errorf("Expected EOF after closing, got %v", err)
	}

	ch <- n
}

func TestFileTail(t *testing.T) {
	fn := os.TempDir() + "/,thefile"
	defer os.Remove(fn)

	f, err := os.OpenFile(fn, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		t.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()

	ch := make(chan int)
	go read(t, fn, ch)

	arbitraryData := make([]byte, fileSize/4)
	written := 0
	for written < fileSize {
		w, err := f.Write(arbitraryData)
		if err != nil {
			t.Fatalf("Error writing data to file: %v", err)
		}
		written += w
		// Give the reader time to catch up.
		time.Sleep(time.Millisecond * 10)
	}

	var bytesRead int
	select {
	case bytesRead = <-ch:
	case <-time.After(time.Second):
		t.Fatalf("Took too long tailing.")
	}

	if bytesRead != fileSize {
		t.Fatalf("Expected to read %v bytes, read %v",
			fileSize, bytesRead)
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, exp time.Duration
	}{
		{0, 0, 0},
		{0, 100, 0},
		{200, 100, 100},
	}

	for _, test := range tests {
		got := min(test.a, test.b)
		if got != test.exp {
			t.Errorf("min(%v, %v) = %v, want %v", test.a, test.b, got, test.exp)
		}
	}
}
