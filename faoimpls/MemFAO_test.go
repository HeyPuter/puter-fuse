package faoimpls

import "testing"

func TestMemFAO(t *testing.T) {
	t.Run("clear-box test for MemFAO->resolvePath", func(t *testing.T) {
		fao := CreateMemFAO()
		node1, ok := fao.resolvePath("/")
		if !ok {
			t.Errorf("Expected true, got false")
		}

		node2, ok := fao.resolvePath("/")
		if !ok {
			t.Errorf("Expected true, got false")
		}

		if node1 != node2 {
			t.Errorf("Expected %v, got %v", node1, node2)
		}
	})

	t.Run("MemFAO->write and read", func(t *testing.T) {
		fao := CreateMemFAO()

		// Create a file
		nodeInfo, err := fao.Create("/", "test-file")
		if err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
		if nodeInfo.Name != "test-file" {
			t.Errorf("Expected 'test-file', got '%s'", nodeInfo.Name)
		}

		// Write to the file
		n, err := fao.Write("/test-file", []byte("test"), 0)
		if err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
		if n != 4 {
			t.Errorf("Expected 4, got %d", n)
		}

		// Read the file
		dest := make([]byte, 4)
		n, err = fao.Read("/test-file", dest, 0)
		if err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
		if n != 4 {
			t.Errorf("Expected 4, got %d", n)
		}
		if string(dest) != "test" {
			t.Errorf("Expected 'test', got '%s'", dest)
		}
	})
}
