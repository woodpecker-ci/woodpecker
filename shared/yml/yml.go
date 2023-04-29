package yml

import (
	"os"

	"codeberg.org/6543/go-yaml2json"
)

func ToJSON(data []byte) ([]byte, error) {
	return yaml2json.Convert(data)
}

func LoadYmlFileAsJSON(path string) (j []byte, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	j, err = ToJSON(data)
	if err != nil {
		return nil, err
	}

	return j, nil
}
