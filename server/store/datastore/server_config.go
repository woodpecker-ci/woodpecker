package datastore

import "github.com/woodpecker-ci/woodpecker/server/model"

func (s storage) ServerConfigGet(key string) (string, error) {
	config := &model.ServerConfig{
		Key: key,
	}

	err := wrapGet(s.engine.Get(config))
	if err != nil {
		return "", err
	}

	return config.Value, nil
}

func (s storage) ServerConfigSet(key string, value string) error {
	config := &model.ServerConfig{
		Key:   key,
		Value: value,
	}

	count, err := s.engine.Count(config)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err := s.engine.Insert(config)
		return err
	}

	_, err = s.engine.Where("key = ?", config.Key).AllCols().Update(config)
	return err
}
