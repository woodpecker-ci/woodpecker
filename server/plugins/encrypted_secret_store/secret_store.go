package encrypted_secret_store

import (
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (wrapper *EncryptedSecretStore) SecretFind(repo *model.Repo, s string) (*model.Secret, error) {
	result, err := wrapper.store.SecretFind(repo, s)
	if err != nil {
		return nil, err
	}
	wrapper.decrypt(result)
	return result, nil
}

func (wrapper *EncryptedSecretStore) SecretList(repo *model.Repo, b bool) ([]*model.Secret, error) {
	results, err := wrapper.store.SecretList(repo, b)
	if err != nil {
		return nil, err
	}
	wrapper.decryptList(results)
	return results, nil
}

func (wrapper *EncryptedSecretStore) SecretCreate(secret *model.Secret) error {
	new := &model.Secret{}
	err := wrapper.store.SecretCreate(new)
	if err != nil {
		return err
	}
	secret.ID = new.ID
	wrapper.encrypt(secret)
	err = wrapper.store.SecretUpdate(secret)
	wrapper.decrypt(secret)
	return err
}

func (wrapper *EncryptedSecretStore) SecretUpdate(secret *model.Secret) error {
	wrapper.encrypt(secret)
	err := wrapper.store.SecretUpdate(secret)
	wrapper.decrypt(secret)
	return err
}

func (wrapper *EncryptedSecretStore) SecretDelete(secret *model.Secret) error {
	return wrapper.store.SecretDelete(secret)
}

func (wrapper *EncryptedSecretStore) OrgSecretFind(s string, s2 string) (*model.Secret, error) {
	result, err := wrapper.store.OrgSecretFind(s, s2)
	if err != nil {
		return nil, err
	}
	wrapper.decrypt(result)
	return result, nil
}

func (wrapper *EncryptedSecretStore) OrgSecretList(s string) ([]*model.Secret, error) {
	results, err := wrapper.store.OrgSecretList(s)
	if err != nil {
		return nil, err
	}
	wrapper.decryptList(results)
	return results, nil
}

func (wrapper *EncryptedSecretStore) GlobalSecretFind(s string) (*model.Secret, error) {
	result, err := wrapper.store.GlobalSecretFind(s)
	if err != nil {
		return nil, err
	}
	wrapper.decrypt(result)
	return result, nil
}

func (wrapper *EncryptedSecretStore) GlobalSecretList() ([]*model.Secret, error) {
	results, err := wrapper.store.GlobalSecretList()
	if err != nil {
		return nil, err
	}
	wrapper.decryptList(results)
	return results, nil
}
