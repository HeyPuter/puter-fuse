package engine

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/afero"
)

type MockConfig struct {
	params map[string]string
}

func (c *MockConfig) GetString(key string) string {
	return c.params[key]
}

func TestBLOBCacheService(t *testing.T) {
	memfs := afero.NewMemMapFs()
	config := &MockConfig{
		params: map[string]string{
			"cacheDir": "/",
		},
	}

	svc := CreateBLOBCacheService(memfs)
	svc.ConfigService = config

	testFileData := []byte("test data")
	testFileReader := bytes.NewReader(testFileData)

	// store a blob
	ref := svc.Store(testFileReader)

	testFileHash := ref.GetHash()

	t.Run("hash is not empty string", func(t *testing.T) {
		if testFileHash == "" {
			t.Errorf("expected non-empty hash, got empty string")
		}
	})

	t.Run("blob is stored", func(t *testing.T) {
		if _, err := memfs.Stat("/" + testFileHash); err != nil {
			t.Errorf("expected blob to be stored, got error: %v", err)
		}
	})

	t.Run("blob is retrievable", func(t *testing.T) {
		reader := svc.Get(testFileHash)
		buf := new(bytes.Buffer)
		buf.ReadFrom(reader)
		if !bytes.Equal(buf.Bytes(), testFileData) {
			t.Errorf("expected retrieved blob to match original blob")
		}
	})

	t.Run("blob is released", func(t *testing.T) {
		ref.Release()
		<-ref.AwaitForgotten()
		newRef := svc.Hold(testFileHash)
		if newRef != nil {
			t.Errorf("got non-nil reference after releasing")
		}
		<-ref.AwaitRemovedFromFS()
		if _, err := memfs.Stat("/" + testFileHash); err == nil {
			t.Errorf("expected file to be deleted, got no error")
		}
	})

	t.Run("multiple references", func(t *testing.T) {
		fmt.Println("beginning multiple references test")

		ref1 := svc.Store(testFileReader)
		testFileHash := ref1.GetHash()
		ref2 := svc.Hold(testFileHash)
		ref3 := svc.Hold(testFileHash)

		// ensure there are 3 references on the entry
		if len(ref1.entry.References) != 3 {
			t.Errorf("expected 3 references, got %d",
				len(ref1.entry.References))
		}

		fmt.Println("releasing references")

		ref1.Release()
		ref2.Release()

		maybeRef := svc.Hold(testFileHash)
		if maybeRef == nil {
			t.Errorf(
				"expected non-nil reference after releasing 2/3 references")
		}
		maybeRef.Release()

		ref3.Release()
		<-ref3.AwaitForgotten()

		maybeRef = svc.Hold(testFileHash)
		if maybeRef != nil {
			t.Errorf(
				"expected nil reference after releasing 3/3 references")
		}
	})
}
