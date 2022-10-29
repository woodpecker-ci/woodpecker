package types

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// FileCaches represents a list of service volumes in compose file.
// It has several representation, hence this specific struct.
type FileCaches struct {
	Caches []*FileCache
}

// FileCache represent a file-based cache
type FileCache struct {
	Name        string `yaml:"-"`
	Destination string `yaml:"-"`
}

func (v *FileCache) Validate() bool {
	return !strings.ContainsAny(v.Name, "/\\:.") && !strings.Contains(v.Destination, ":")
}

// String implements the Stringer interface.
func (v *FileCache) String(basePath, repoBase string) string {
	hostPath := v.HostCachePath(basePath, repoBase)
	if hostPath == "" {
		return ""
	}
	return fmt.Sprintf("%s:%s", hostPath, v.Destination)
}

func (v *FileCache) HostCachePath(basePath, repoBase string) string {
	path := filepath.Join(basePath, repoBase, v.Name)
	err := os.MkdirAll(path, 0o750)
	if err != nil {
		log.Error().Err(err).Msgf("could not path %s for file caching", path)
		return ""
	}
	return path
}

// MarshalYAML implements the Marshaller interface.
func (v FileCaches) MarshalYAML() (interface{}, error) {
	var vs []string
	for _, volume := range v.Caches {
		vs = append(vs, volume.String("", ""))
	}
	return vs, nil
}

// UnmarshalYAML implements the Unmarshaler interface.
func (v *FileCaches) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sliceType []interface{}
	if err := unmarshal(&sliceType); err == nil {
		v.Caches = []*FileCache{}
		for _, volume := range sliceType {
			name, ok := volume.(string)
			if !ok {
				return fmt.Errorf("cannot unmarshal '%v' to type %T into a string value", name, name)
			}
			elts := strings.SplitN(name, ":", 2)
			var vol *FileCache
			if len(elts) == 2 {
				vol = &FileCache{
					Name:        elts[0],
					Destination: elts[1],
				}
			} else {
				return fmt.Errorf("cache must have a name and a destination")
			}
			v.Caches = append(v.Caches, vol)
		}
		return nil
	}

	return errors.New("failed to unmarshal FileCaches")
}
