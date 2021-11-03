package datastore_xorm

import (
	"bytes"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"io"
	"io/ioutil"

	"xorm.io/builder"
)

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
