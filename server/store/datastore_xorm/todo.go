package datastore_xorm

import (
	"errors"
	"io"
	"time"

	"github.com/woodpecker-ci/woodpecker/server/model"

	"github.com/rs/zerolog/log"
)

const perPage = 50

var RecordNotExist = errors.New("requested object not exist")

func wrapGet(exist bool, err error) error {
	if err != nil {
		return err
	}
	if !exist {
		return RecordNotExist
	}
	return nil
}

func (s storage) GetUser(id int64) (*model.User, error) {
	user := new(model.User)
	err := wrapGet(s.engine.ID(id).Get(user))
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s storage) GetUserLogin(login string) (*model.User, error) {
	user := new(model.User)
	err := wrapGet(s.engine.Where("user_login=?", login).Get(user))
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s storage) GetUserList() ([]*model.User, error) {
	users := make([]*model.User, 0, 10)
	err := s.engine.Find(&users)
	return users, err
}

func (s storage) GetUserCount() (int, error) {
	c, err := s.engine.Count(&model.User{})
	return int(c), err
}

func (s storage) CreateUser(user *model.User) error {
	_, err := s.engine.InsertOne(user)
	return err
}

func (s storage) UpdateUser(user *model.User) error {
	_, err := s.engine.ID(user.ID).AllCols().Update(user)
	return err
}

func (s storage) DeleteUser(user *model.User) error {
	_, err := s.engine.ID(user.ID).Delete(&user)
	// TODO: delete related content that need this user to work
	return err
}

func (s storage) GetRepo(i int64) (*model.Repo, error) {
	panic("implement me")
}

func (s storage) GetRepoName(fullName string) (*model.Repo, error) {
	repo := &model.Repo{}
	err := wrapGet(s.engine.Where("repo_full_name = ?", fullName).Get(repo))
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (s storage) GetRepoCount() (int, error) {
	c, err := s.engine.Count(&model.Repo{IsActive: true})
	return int(c), err
}

func (s storage) CreateRepo(repo *model.Repo) error {
	_, err := s.engine.InsertOne(repo)
	return err
}

func (s storage) UpdateRepo(repo *model.Repo) error {
	_, err := s.engine.ID(repo.ID).AllCols().Update(repo)
	return err
}

func (s storage) DeleteRepo(repo *model.Repo) error {
	_, err := s.engine.ID(repo.ID).Delete(repo)

	// TODO: delete related within a session

	return err
}

func (s storage) GetBuild(id int64) (*model.Build, error) {
	build := &model.Build{}
	err := wrapGet(s.engine.ID(id).Get(build))
	if err != nil {
		return nil, err
	}
	return build, nil
}

func (s storage) GetBuildNumber(repo *model.Repo, num int) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Number: num,
	}
	err := wrapGet(s.engine.Get(build))
	if err != nil {
		return nil, err
	}
	return build, nil
}

func (s storage) GetBuildRef(repo *model.Repo, ref string) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Ref:    ref,
	}
	err := wrapGet(s.engine.Get(build))
	if err != nil {
		return nil, err
	}
	return build, err
}

func (s storage) GetBuildCommit(repo *model.Repo, sha, branch string) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Branch: branch,
		Commit: sha,
	}
	err := wrapGet(s.engine.Get(build))
	if err != nil {
		return nil, err
	}
	return build, nil
}

func (s storage) GetBuildLast(repo *model.Repo, branch string) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Branch: branch,
		Event:  "push",
	}
	if err := wrapGet(s.engine.Desc("build_number").Get(build)); err != nil {
		return nil, err
	}
	return build, nil
}

func (s storage) GetBuildLastBefore(repo *model.Repo, branch string, num int64) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Branch: branch,
	}
	if err := wrapGet(s.engine.Desc("build_number").Where("build_id < ?", num).Get(build)); err != nil {
		return nil, err
	}
	return build, nil
}

func (s storage) GetBuildList(repo *model.Repo, page int) ([]*model.Build, error) {
	builds := make([]*model.Build, 0, perPage)
	err := s.engine.Where("build_repo_id = ?", repo.ID).
		Desc("build_number").
		Limit(perPage, perPage*(page-1)).
		Find(&builds)
	return builds, err
}

func (s storage) GetBuildQueue() ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)
	// TODO: fix model & use xorm.Builder
	err := s.engine.SQL(`
SELECT
 repo_owner
,repo_name
,repo_full_name
,build_number
,build_event
,build_status
,build_created
,build_started
,build_finished
,build_commit
,build_branch
,build_ref
,build_refspec
,build_remote
,build_title
,build_message
,build_author
,build_email
,build_avatar
FROM
 builds b
,repos r
WHERE b.build_repo_id = r.repo_id
  AND b.build_status IN ('pending','running')
`).Find(&feed)
	return feed, err
}

func (s storage) GetBuildCount() (int, error) {
	c, err := s.engine.Count(&model.Build{})
	return int(c), err
}

func (s storage) CreateBuild(build *model.Build, procs ...*model.Proc) error {
	build.Trim()

	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	// increment counter
	if _, err := sess.ID(build.RepoID).Incr("repo_counter").Update(new(model.Repo)); err != nil {
		return err
	}

	repo := new(model.Repo)
	if err := wrapGet(sess.ID(build.RepoID).Get(repo)); err != nil {
		return err
	}

	build.Number = repo.Counter
	build.Created = time.Now().UTC().Unix()
	build.Enqueued = build.Created
	if _, err := sess.InsertOne(build); err != nil {
		return err
	}

	if _, err := sess.InsertMulti(procs); err != nil {
		return err
	}

	return sess.Commit()
}

func (s storage) UpdateBuild(build *model.Build) error {
	_, err := s.engine.ID(build.ID).AllCols().Update(build)
	return err
}

func (s storage) UserFeed(user *model.User) ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)
	// TODO use xorm.Builder
	err := s.engine.SQL(`
SELECT
 repo_owner
,repo_name
,repo_full_name
,build_number
,build_event
,build_status
,build_created
,build_started
,build_finished
,build_commit
,build_branch
,build_ref
,build_refspec
,build_remote
,build_title
,build_message
,build_author
,build_email
,build_avatar
FROM repos
INNER JOIN perms  ON perms.perm_repo_id   = repos.repo_id
INNER JOIN builds ON builds.build_repo_id = repos.repo_id
`).Where("perms.perm_user_id = ?", user.ID).
		Limit(perPage).Desc("build_id").
		Find(&feed)
	return feed, err
}

func (s storage) RepoList(user *model.User) ([]*model.Repo, error) {
	repos := make([]*model.Repo, 0, perPage)
	err := s.engine.Table("repos").
		Join("INNER", "perms", "perms.perm_repo_id = repos.repo_id").
		Where("perms.perm_user_id = ?", user.ID).
		Asc("repo_full_name").
		Find(&repos)
	// TODO: limit
	return repos, err
}

func (s storage) RepoListLatest(user *model.User) ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)
	const feedLatestBuild = `
SELECT
 repo_owner
,repo_name
,repo_full_name
,build_number
,build_event
,build_status
,build_created
,build_started
,build_finished
,build_commit
,build_branch
,build_ref
,build_refspec
,build_remote
,build_title
,build_message
,build_author
,build_email
,build_avatar
FROM repos LEFT OUTER JOIN builds ON build_id = (
	SELECT build_id FROM builds
	WHERE builds.build_repo_id = repos.repo_id
	ORDER BY build_id DESC
	LIMIT 1
)
INNER JOIN perms ON perms.perm_repo_id = repos.repo_id
WHERE perms.perm_user_id = ?
  AND repos.repo_active = true
ORDER BY repo_full_name ASC;
`
	err := s.engine.SQL(feedLatestBuild).Find(&feed)
	return feed, err
}

func (s storage) RepoBatch(repos []*model.Repo) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	for _, repo := range repos {
		if repo.UserID == 0 || len(repo.Owner) == 0 || len(repo.Name) == 0 || len(repo.FullName) == 0 {
			log.Debug().Msgf("skip insert/update repo: %v", repo)
			continue
		}
		exist, err := sess.Exist(&repo)
		if err != nil {
			return err
		}
		if exist {
			if _, err := sess.Update(&repo); err != nil {
				return err
			}
		} else {
			if _, err := sess.InsertOne(&repo); err != nil {
				return err
			}
		}
	}

	return sess.Commit()
}

func (s storage) PermFind(user *model.User, repo *model.Repo) (*model.Perm, error) {
	perm := &model.Perm{
		UserID: user.ID,
		RepoID: repo.ID,
	}
	if err := wrapGet(s.engine.Get(perm)); err != nil {
		return nil, err
	}
	return perm, nil
}

func (s storage) PermUpsert(perm *model.Perm) error {
	// TODO: do we need sql?!? - what is this func for?
	_, err := s.engine.SQL(`
REPLACE INTO perms (
 perm_user_id
,perm_repo_id
,perm_pull
,perm_push
,perm_admin
,perm_synced
) VALUES (?,(SELECT repo_id FROM repos WHERE repo_full_name = ? ),?,?,?,?)
`,
		perm.UserID,
		perm.Repo,
		perm.Pull,
		perm.Push,
		perm.Admin,
		perm.Synced,
	).Exec()

	return err
}

func (s storage) PermBatch(perms []*model.Perm) error {
	for i := range perms {
		if err := s.PermUpsert(perms[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s storage) PermDelete(perm *model.Perm) error {
	_, err := s.engine.
		Where("perm_user_id = ? AND perm_repo_id = ?", perm.UserID, perm.RepoID).
		Delete(new(model.Perm))
	return err
}

func (s storage) PermFlush(user *model.User, before int64) error {
	_, err := s.engine.
		Where("perm_user_id = ? AND perm_synced < ?", user.ID, before).
		Delete(new(model.Perm))
	return err
}

func (s storage) ConfigsForBuild(buildID int64) ([]*model.Config, error) {
	configs := make([]*model.Config, 0, perPage)
	err := s.engine.
		Table("config").
		Join("LEFT", "build_config", "config.config_id = build_config.config_id").
		Where("build_config.build_id = ?", buildID).
		Find(&configs)
	return configs, err
}

func (s storage) ConfigFindIdentical(repoID int64, hash string) (*model.Config, error) {
	var configFindRepoHash = `
SELECT
 config_id
,config_repo_id
,config_hash
,config_data
,config_name
FROM config
WHERE config_repo_id = ?
  AND config_hash    = ?
`
	conf := &model.Config{
		RepoID: repoID,
		Hash:   hash,
	}
	if err := wrapGet(s.engine.Get(conf)); err != nil {
		return nil, err
	}
	return conf, nil
}

func (s storage) ConfigFindApproved(config *model.Config) (bool, error) {
	panic("implement me")
}

func (s storage) ConfigCreate(config *model.Config) error {
	panic("implement me")
}

func (s storage) BuildConfigCreate(config *model.BuildConfig) error {
	panic("implement me")
}

func (s storage) SenderFind(repo *model.Repo, s2 string) (*model.Sender, error) {
	panic("implement me")
}

func (s storage) SenderList(repo *model.Repo) ([]*model.Sender, error) {
	panic("implement me")
}

func (s storage) SenderCreate(sender *model.Sender) error {
	panic("implement me")
}

func (s storage) SenderUpdate(sender *model.Sender) error {
	panic("implement me")
}

func (s storage) SenderDelete(sender *model.Sender) error {
	panic("implement me")
}

func (s storage) SecretFind(repo *model.Repo, s2 string) (*model.Secret, error) {
	panic("implement me")
}

func (s storage) SecretList(repo *model.Repo) ([]*model.Secret, error) {
	panic("implement me")
}

func (s storage) SecretCreate(secret *model.Secret) error {
	panic("implement me")
}

func (s storage) SecretUpdate(secret *model.Secret) error {
	panic("implement me")
}

func (s storage) SecretDelete(secret *model.Secret) error {
	panic("implement me")
}

func (s storage) RegistryFind(repo *model.Repo, s2 string) (*model.Registry, error) {
	panic("implement me")
}

func (s storage) RegistryList(repo *model.Repo) ([]*model.Registry, error) {
	panic("implement me")
}

func (s storage) RegistryCreate(registry *model.Registry) error {
	panic("implement me")
}

func (s storage) RegistryUpdate(registry *model.Registry) error {
	panic("implement me")
}

func (s storage) RegistryDelete(registry *model.Registry) error {
	panic("implement me")
}

func (s storage) ProcLoad(i int64) (*model.Proc, error) {
	panic("implement me")
}

func (s storage) ProcFind(build *model.Build, i int) (*model.Proc, error) {
	panic("implement me")
}

func (s storage) ProcChild(build *model.Build, i int, s2 string) (*model.Proc, error) {
	panic("implement me")
}

func (s storage) ProcList(build *model.Build) ([]*model.Proc, error) {
	panic("implement me")
}

func (s storage) ProcCreate(procs []*model.Proc) error {
	panic("implement me")
}

func (s storage) ProcUpdate(proc *model.Proc) error {
	panic("implement me")
}

func (s storage) ProcClear(build *model.Build) error {
	panic("implement me")
}

func (s storage) LogFind(proc *model.Proc) (io.ReadCloser, error) {
	panic("implement me")
}

func (s storage) LogSave(proc *model.Proc, reader io.Reader) error {
	panic("implement me")
}

func (s storage) FileList(build *model.Build) ([]*model.File, error) {
	panic("implement me")
}

func (s storage) FileFind(proc *model.Proc, s2 string) (*model.File, error) {
	panic("implement me")
}

func (s storage) FileRead(proc *model.Proc, s2 string) (io.ReadCloser, error) {
	panic("implement me")
}

func (s storage) FileCreate(file *model.File, reader io.Reader) error {
	panic("implement me")
}

func (s storage) TaskList() ([]*model.Task, error) {
	panic("implement me")
}

func (s storage) TaskInsert(task *model.Task) error {
	panic("implement me")
}

func (s storage) TaskDelete(s2 string) error {
	panic("implement me")
}

func (s storage) Ping() error {
	return s.engine.Ping()
}
