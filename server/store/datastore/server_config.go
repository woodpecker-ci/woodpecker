package datastore

import "go.woodpecker-ci.org/woodpecker/server/model"

func (s storage) ServerConfigGet(key string) (string, error) {
	config := new(model.ServerConfig)
	err := wrapGet(s.engine.ID(key).Get(config))
	if err != nil {
		return "", err
	}

	return config.Value, nil
}

func (s storage) ServerConfigSet(key, value string) error {
	config := &model.ServerConfig{
		Key: key,
	}

	count, err := s.engine.Count(config)
	if err != nil {
		return err
	}

	config.Value = value

	if count == 0 {
		_, err := s.engine.Insert(config)
		return err
	}

	_, err = s.engine.Where("key = ?", config.Key).AllCols().Update(config)
	return err
}

func (s storage) ServerConfigDelete(key string) error {
	config := &model.ServerConfig{
		Key: key,
	}

	return wrapDelete(s.engine.Delete(config))
}
