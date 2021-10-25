package datastore_xorm

import (
	"io"
	"time"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

const perPage = 50

func (s storage) GetUser(id int64) (*model.User, error) {
	user := &model.User{}
	_, err := s.engine.ID(id).Get(&user)
	return user, err
}

func (s storage) GetUserLogin(login string) (*model.User, error) {
	user := &model.User{}
	_, err := s.engine.Where("user_login=?", login).Get(&user)
	return user, err
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
	_, err := s.engine.Where("repo_full_name = ?", fullName).Get(&repo)
	return repo, err
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
	_, err := s.engine.ID(id).Get(&build)
	return build, err
}

func (s storage) GetBuildNumber(repo *model.Repo, num int) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Number: num,
	}
	_, err := s.engine.Get(&build)
	return build, err
}

func (s storage) GetBuildRef(repo *model.Repo, ref string) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Ref:    ref,
	}
	_, err := s.engine.Get(&build)
	return build, err
}

func (s storage) GetBuildCommit(repo *model.Repo, sha, branch string) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Branch: branch,
		Commit: sha,
	}
	_, err := s.engine.Get(&build)
	return build, err
}

func (s storage) GetBuildLast(repo *model.Repo, branch string) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Branch: branch,
		Event:  "push",
	}
	_, err := s.engine.Desc("build_number").Get(build)
	return build, err
}

func (s storage) GetBuildLastBefore(repo *model.Repo, branch string, num int64) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Branch: branch,
	}
	_, err := s.engine.Desc("build_number").Where("build_id < ?", num).Get(build)
	return build, err
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

	var repo *model.Repo
	if _, err := sess.ID(build.RepoID).Get(&repo); err != nil {
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
`).Where("perms.perm_user_id = ?", user.ID).Limit(perPage).Desc("build_id").Find(&feed)
	return feed, err
}

func (s storage) RepoList(user *model.User) ([]*model.Repo, error) {
	panic("implement me")
}

func (s storage) RepoListLatest(user *model.User) ([]*model.Feed, error) {
	panic("implement me")
}

func (s storage) RepoBatch(repos []*model.Repo) error {
	panic("implement me")
}

func (s storage) PermFind(user *model.User, repo *model.Repo) (*model.Perm, error) {
	panic("implement me")
}

func (s storage) PermUpsert(perm *model.Perm) error {
	panic("implement me")
}

func (s storage) PermBatch(perms []*model.Perm) error {
	panic("implement me")
}

func (s storage) PermDelete(perm *model.Perm) error {
	panic("implement me")
}

func (s storage) PermFlush(user *model.User, before int64) error {
	panic("implement me")
}

func (s storage) ConfigsForBuild(buildID int64) ([]*model.Config, error) {
	panic("implement me")
}

func (s storage) ConfigFindIdentical(repoID int64, sha string) (*model.Config, error) {
	panic("implement me")
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
	panic("implement me")
}
