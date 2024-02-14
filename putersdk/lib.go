package putersdk

import (
	"encoding/json"
	"fmt"
)

func unmarshalIntoStruct(obj map[string]interface{}, dest interface{}) error {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("error marshaling into JSON: %s", err)
	}

	if err := json.Unmarshal(jsonData, dest); err != nil {
		return fmt.Errorf("error unmarshaling into struct: %s", err)
	}

	return nil
}
