// GENERATED(gen.js) - DO NOT EDIT BY HAND - See meta/models.json.js

package fao

import (
	"fmt"
)

type BaseFAO struct {
	FAO
}

func (base *BaseFAO) ReadAll(path string) ([]byte, error) {
	stat, exists, err := base.Stat(path)
	if err != nil {
	    return nil, err
	}
	if !exists {
	    return nil, fmt.Errorf("file does not exist")
	}
	buf := make([]byte, stat.Size)
	n, err := base.Read(path, buf, 0)
	if err != nil {
	    return nil, err
	}
	return buf[:n], nil
}
