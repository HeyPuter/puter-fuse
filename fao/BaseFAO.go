// GENERATED(gen.js) - DO NOT EDIT BY HAND - See meta/models.json.js

package fao

type BaseFAO struct {
	FAO
}

func (base *BaseFAO) ReadAll(path string) ([]byte, error) {
	stat, err := base.Stat(path)
	if err != nil {
	    return nil, err
	}
	buf := make([]byte, stat.Size)
	n, err := base.Read(path, buf, 0)
	if err != nil {
	    return nil, err
	}
	return buf[:n], nil
}
