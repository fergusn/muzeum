package cache

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/docker/distribution/registry/storage/driver/testdriver"
)

func TestWhenNotInCacheThenPrimedFromLoader(t *testing.T) {
	s := testdriver.New()
	c := NewCache(s)

	expected := []byte("123456")

	c.Read(context.TODO(), "/abcdef", func() (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewBuffer(expected)), nil
	})

	actual, err := s.GetContent(context.TODO(), "/abcdef")
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(expected, actual) != 0 {
		t.Errorf("expcted cache primed with %v, got %v", expected, actual)
	}
}

func TestWhenInCacheLoaderIsNotCalled(t *testing.T) {
	s := testdriver.New()
	c := NewCache(s)

	s.PutContent(context.TODO(), "/abcdef", []byte{1, 2, 3, 4})

	rd, _ := c.Read(context.TODO(), "/abcdef", func() (io.ReadCloser, error) {
		t.Error("loader should not be called")
		return nil, nil
	})
	rd.Close()
}


func TestWhenInflightLoaderThenLoad(t *testing.T) {
	s := testdriver.New()
	c := NewCache(s)

	wg := sync.WaitGroup{}
	ch := make(chan bool)

	go func() {
		c.Read(context.TODO(), "/abcdef", func() (io.ReadCloser, error) {
			wg.Add(1)
			ch <- true
			wg.Wait()
			return ioutil.NopCloser(bytes.NewBufferString("9876543")), nil
		})
	}()

	<-ch

	c.Read(context.TODO(), "/abcdef", func() (io.ReadCloser, error) {
		wg.Done()
		return ioutil.NopCloser(bytes.NewBufferString("123456")), nil
	})
}
