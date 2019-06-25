package databases

import (
	"context"
	"github.com/pkg/errors"
	"github.com/yuuki0xff/clustertest/models"
	"sync"
	"time"
)

type MemTaskDB struct {
	m        sync.Mutex
	nextID   int
	waiting  map[string]models.Task
	running  map[string]struct{}
	finished map[string]models.TaskResult
}
type MemTask struct {
	Spec []byte
}
type MemTaskDetail struct {
	ID models.TaskID
	DB *MemTaskDB
}
type MemTaskResult struct {
	before models.ScriptResult
	main   models.ScriptResult
	after  models.ScriptResult
}

func NewMemTaskDB() *MemTaskDB {
	return &MemTaskDB{
		waiting:  map[string]models.Task{},
		running:  map[string]struct{}{},
		finished: map[string]models.TaskResult{},
	}
}
func (db *MemTaskDB) Create(task models.Task) (models.TaskID, error) {
	db.m.Lock()
	defer db.m.Unlock()

	id := &IntTaskID{
		ID: db.nextID,
	}
	db.nextID++

	db.waiting[id.String()] = task
	return id, nil
}
func (db *MemTaskDB) Inspect(id models.TaskID) (models.TaskDetail, error) {
	return &MemTaskDetail{
		ID: id,
		DB: db,
	}, nil
}
func (db *MemTaskDB) Wait(id models.TaskID, ctx context.Context) error {
	td, err := db.Inspect(id)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			if td.State() == "finished" {
				return nil
			}
		case <-ctx.Done():
			return errors.Errorf("context canceled")
		}
	}
}
func (db *MemTaskDB) Cancel(id models.TaskID) error {
	// todo
	panic("not impl")
}
func (db *MemTaskDB) Delete(id models.TaskID) error {
	db.m.Lock()
	defer db.m.Unlock()

	sid := id.String()

	if _, ok := db.waiting[sid]; ok {
		// Task is waiting.
		delete(db.waiting, sid)
		return nil
	}
	if _, ok := db.running[sid]; ok {
		// Cannot stop delete it because it is running.
		return errors.Errorf("failed to delete task: task(%s) is running", sid)
	}
	if _, ok := db.finished[sid]; ok {
		// Task is finished.
		delete(db.finished, sid)
		return nil
	}
	return errors.Errorf("not found task: %s", sid)
}
func (db *MemTaskDB) Consume(fn models.TaskConsumer) error {
	var sid string
	var task models.Task
	var ok bool
	// Get a task from waiting queue and move task to running.
	db.m.Lock()
	for sid, task = range db.waiting {
		ok = true
		break
	}
	if ok {
		delete(db.waiting, sid)
		db.running[sid] = struct{}{}
	}
	db.m.Unlock()

	if !ok {
		return models.QueueEmpty
	}

	// Consume a task.
	id := &StringTaskID{ID: sid}
	result, err := fn(id, task)
	if err != nil {
		// TODO
		panic("not impl")
	}

	// Move task to finished.
	db.m.Lock()
	delete(db.running, sid)
	db.finished[sid] = result
	db.m.Unlock()
	return nil
}
func (db *MemTaskDB) List() ([]models.TaskDetail, error) {
	db.m.Lock()
	defer db.m.Unlock()

	var ids []string
	for id := range db.waiting {
		ids = append(ids, id)
	}
	for id := range db.running {
		ids = append(ids, id)
	}
	for id := range db.finished {
		ids = append(ids, id)
	}

	var ds []models.TaskDetail
	for _, id := range ids {
		ds = append(ds, &MemTaskDetail{
			ID: &StringTaskID{id},
			DB: db,
		})
	}
	return ds, nil
}

func (t *MemTask) String() string {
	return "<MemTask>"
}
func (t *MemTask) SpecData() []byte {
	return t.Spec
}

func (d *MemTaskDetail) String() string {
	return "<MemTaskDetail>"
}
func (d *MemTaskDetail) TaskID() models.TaskID {
	return d.ID
}
func (d *MemTaskDetail) State() string {
	d.DB.m.Lock()
	defer d.DB.m.Unlock()

	sid := d.ID.String()
	db := d.DB
	if _, ok := db.waiting[sid]; ok {
		return "waiting"
	}
	if _, ok := db.running[sid]; ok {
		return "running"
	}
	if _, ok := db.finished[sid]; ok {
		return "finished"
	}

	err := errors.Errorf("not found task: %s", sid)
	panic(err)
}
func (d *MemTaskDetail) Result() models.TaskResult {
	d.DB.m.Lock()
	defer d.DB.m.Unlock()

	return d.DB.finished[d.ID.String()]
}

func (r *MemTaskResult) String() string {
	return "<MemTaskResult>"
}
func (r *MemTaskResult) Error() error {
	if r.before != nil && r.before.ExitCode() != 0 {
		return errors.Errorf("before script failed with exit code %d", r.before.ExitCode())
	}
	if r.main != nil && r.main.ExitCode() != 0 {
		return errors.Errorf("main script failed with exit code %d", r.main.ExitCode())
	}
	if r.after != nil && r.after.ExitCode() != 0 {
		return errors.Errorf("after script failed with exit code %d", r.after.ExitCode())
	}
	return nil
}
func (r *MemTaskResult) BeforeResult() models.ScriptResult {
	return r.before
}
func (r *MemTaskResult) ScriptResult() models.ScriptResult {
	return r.main
}
func (r *MemTaskResult) AfterResult() models.ScriptResult {
	return r.after
}
