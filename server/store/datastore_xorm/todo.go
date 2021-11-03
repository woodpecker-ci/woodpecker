package datastore_xorm

import (
	"bytes"
	"io"
	"io/ioutil"
	"time"

	"github.com/woodpecker-ci/woodpecker/server/model"

	"github.com/rs/zerolog/log"
	"xorm.io/builder"
)

func (s storage) GetRepo(id int64) (*model.Repo, error) {
	repo := new(model.Repo)
	err := wrapGet(s.engine.ID(id).Get(repo))
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (s storage) GetRepoName(fullName string) (*model.Repo, error) {
	repo := new(model.Repo)
	err := wrapGet(s.engine.Where("repo_full_name = ?", fullName).Get(repo))
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (s storage) GetRepoCount() (int64, error) {
	return s.engine.Count(&model.Repo{IsActive: true})
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

func (s storage) GetBuildCount() (int64, error) {
	return s.engine.Count(&model.Build{})
}

func (s storage) CreateBuild(build *model.Build, procList ...*model.Proc) error {
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

	if _, err := sess.InsertMulti(procList); err != nil {
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
		And(builder.Eq{"perms.perm_push": true}.Or(builder.Eq{"perms.perm_admin": true})).
		Limit(perPage).Desc("build_id").
		Find(&feed)
	return feed, err
}

func (s storage) RepoList(user *model.User, owned bool) ([]*model.Repo, error) {
	repos := make([]*model.Repo, 0, perPage)
	sess := s.engine.Table("repos").
		Join("INNER", "perms", "perms.perm_repo_id = repos.repo_id").
		Where("perms.perm_user_id = ?", user.ID)
	if owned {
		sess = sess.And(builder.Eq{"perms.perm_push": true}.Or(builder.Eq{"perms.perm_admin": true}))
	}
	err := sess.
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
  AND (perms.perm_push = 1 OR perms.perm_admin = 1)
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
	return s.engine.Where("build_repo_id = ?", config.RepoID).
		And(builder.In("build_id", builder.Expr( // TODO: use JOIN
			`SELECT build_id
  FROM build_config
  WHERE build_config.config_id = ?`, config.ID))).
		And(builder.In("build_status", "blocked", "pending")).
		Exist(new(model.Build))
}

func (s storage) ConfigCreate(config *model.Config) error {
	_, err := s.engine.InsertOne(config)
	return err
}

func (s storage) BuildConfigCreate(config *model.BuildConfig) error {
	_, err := s.engine.InsertOne(config)
	return err
}

func (s storage) SenderFind(repo *model.Repo, login string) (*model.Sender, error) {
	sender := &model.Sender{
		RepoID: repo.ID,
		Login:  login,
	}
	if err := wrapGet(s.engine.Get(sender)); err != nil {
		return nil, err
	}
	return sender, nil
}

func (s storage) SenderList(repo *model.Repo) ([]*model.Sender, error) {
	senders := make([]*model.Sender, 0, perPage)
	err := s.engine.Where("sender_repo_id = ?", repo.ID).Find(&senders)
	return senders, err
}

func (s storage) SenderCreate(sender *model.Sender) error {
	_, err := s.engine.InsertOne(sender)
	return err
}

func (s storage) SenderUpdate(sender *model.Sender) error {
	_, err := s.engine.ID(sender.ID).Update(sender)
	return err
}

func (s storage) SenderDelete(sender *model.Sender) error {
	_, err := s.engine.ID(sender.ID).Delete(new(model.Sender))
	return err
}

func (s storage) SecretFind(repo *model.Repo, name string) (*model.Secret, error) {
	secret := &model.Secret{
		RepoID: repo.ID,
		Name:   name,
	}
	if err := wrapGet(s.engine.Get(secret)); err != nil {
		return nil, err
	}
	return secret, nil
}

func (s storage) SecretList(repo *model.Repo) ([]*model.Secret, error) {
	secrets := make([]*model.Secret, 0, perPage)
	err := s.engine.Where("secret_repo_id = ?", repo.ID).Find(&secrets)
	return secrets, err
}

func (s storage) SecretCreate(secret *model.Secret) error {
	_, err := s.engine.InsertOne(secret)
	return err
}

func (s storage) SecretUpdate(secret *model.Secret) error {
	_, err := s.engine.ID(secret.ID).Update(&secret)
	return err
}

func (s storage) SecretDelete(secret *model.Secret) error {
	_, err := s.engine.ID(secret.ID).Delete(new(model.Secret))
	return err
}

func (s storage) RegistryFind(repo *model.Repo, addr string) (*model.Registry, error) {
	reg := &model.Registry{
		RepoID:  repo.ID,
		Address: addr,
	}
	if err := wrapGet(s.engine.Get(reg)); err != nil {
		return nil, err
	}
	return reg, nil
}

func (s storage) RegistryList(repo *model.Repo) ([]*model.Registry, error) {
	regs := make([]*model.Registry, 0, perPage)
	err := s.engine.Where("registry_repo_id = ?", repo.ID).Find(&regs)
	return regs, err
}

func (s storage) RegistryCreate(registry *model.Registry) error {
	_, err := s.engine.InsertOne(registry)
	return err
}

func (s storage) RegistryUpdate(registry *model.Registry) error {
	_, err := s.engine.ID(registry.ID).Update(registry)
	return err
}

func (s storage) RegistryDelete(registry *model.Registry) error {
	_, err := s.engine.ID(registry.ID).Delete(new(model.Registry))
	return err
}

func (s storage) ProcLoad(id int64) (*model.Proc, error) {
	proc := new(model.Proc)
	if err := wrapGet(s.engine.ID(id).Get(proc)); err != nil {
		return nil, err
	}
	return proc, nil
}

func (s storage) ProcFind(build *model.Build, pid int) (*model.Proc, error) {
	proc := &model.Proc{
		BuildID: build.ID,
		PID:     pid,
	}
	if err := wrapGet(s.engine.Get(proc)); err != nil {
		return nil, err
	}
	return proc, nil
}

func (s storage) ProcChild(build *model.Build, ppid int, child string) (*model.Proc, error) {
	proc := &model.Proc{
		BuildID: build.ID,
		PPID:    ppid,
		Name:    child,
	}
	if err := wrapGet(s.engine.Get(proc)); err != nil {
		return nil, err
	}
	return proc, nil
}

func (s storage) ProcList(build *model.Build) ([]*model.Proc, error) {
	procList := make([]*model.Proc, 0, perPage)
	err := s.engine.Where("proc_build_id = ?", build.ID).Find(&procList)
	return procList, err
}

func (s storage) ProcCreate(procs []*model.Proc) error {
	_, err := s.engine.Insert(procs)
	return err
}

func (s storage) ProcUpdate(proc *model.Proc) error {
	_, err := s.engine.ID(proc.ID).Update(proc)
	return err
}

func (s storage) ProcClear(build *model.Build) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.Where("file_build_id = ?", build.ID).Delete(new(model.File)); err != nil {
		return err
	}

	if _, err := sess.Where("proc_build_id = ?", build.ID).Delete(new(model.Proc)); err != nil {
		return err
	}

	return sess.Commit()
}

func (s storage) LogFind(proc *model.Proc) (io.ReadCloser, error) {
	logs := &model.Logs{
		ProcID: proc.ID,
	}
	if err := wrapGet(s.engine.Get(logs)); err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(logs.Data)
	return ioutil.NopCloser(buf), nil
}

func (s storage) LogSave(proc *model.Proc, reader io.Reader) error {
	data, _ := ioutil.ReadAll(reader)

	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	logs := new(model.Logs)
	exist, err := sess.Where("log_job_id = ?", proc.ID).Get(logs)
	if err != nil {
		return err
	}

	if exist {
		if _, err := sess.ID(logs.ID).Cols("log_data").Update(&model.Logs{Data: data}); err != nil {
			return err
		}
	} else {
		if _, err := sess.InsertOne(&model.Logs{
			ProcID: proc.ID,
			Data:   data,
		}); err != nil {
			return err
		}
	}

	return sess.Commit()
}

func (s storage) FileList(build *model.Build) ([]*model.File, error) {
	files := make([]*model.File, 0, perPage)
	err := s.engine.Where("file_build_id = ?", build.ID).Find(&files)
	return files, err
}

func (s storage) FileFind(proc *model.Proc, name string) (*model.File, error) {
	file := &model.File{
		ProcID: proc.ID,
		Name:   name,
	}
	if err := wrapGet(s.engine.Get(file)); err != nil {
		return nil, err
	}
	return file, nil
}

func (s storage) FileRead(proc *model.Proc, name string) (io.ReadCloser, error) {
	file, err := s.FileFind(proc, name)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(file.Data)
	return ioutil.NopCloser(buf), err
}

func (s storage) FileCreate(file *model.File, reader io.Reader) error {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	file.Data = data
	_, err = s.engine.InsertOne(file)
	return err
}

func (s storage) TaskList() ([]*model.Task, error) {
	tasks := make([]*model.Task, 0, perPage)
	err := s.engine.Find(&tasks)
	return tasks, err
}

func (s storage) TaskInsert(task *model.Task) error {
	_, err := s.engine.InsertOne(task)
	return err
}

func (s storage) TaskDelete(id string) error {
	_, err := s.engine.Where("task_id = ?", id).Delete(new(model.Task))
	return err
}

func (s storage) Ping() error {
	return s.engine.Ping()
}
