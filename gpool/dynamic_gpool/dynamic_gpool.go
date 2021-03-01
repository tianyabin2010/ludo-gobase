package dynamic_gpool

import (
	"container/list"
	"github.com/rs/zerolog/log"
	"github.com/tianyabin2010/ludo-gobase/util"
	"time"
)

type Job func()

type Pool interface {
	Post(Job)
}

type worker struct {
	id          int
	JobBus      chan Job
	stopChan    chan struct{}
	lastUseTime *time.Time
}

func (w *worker) getLastUseTime() *time.Time {
	ret := w.lastUseTime
	return ret
}

func (w *worker) stop() {
	w.stopChan <- struct{}{}
}

func (w *worker) run(p *gpool) {
	defer util.BtRecover("worker.Run")
	for {
		select {
		case <-w.stopChan:
			//TODO 从worker列表删除
			log.Info().Str("gpool name", p.Name).
				Int("id", w.id).
				Msgf("worker exit")
			return
		default:
			select {
			case p.Workers <- w:
				job := <-w.JobBus
				if nil != job {
					job()
				}
				now := time.Now()
				w.lastUseTime = &now
			case <-w.stopChan:
				//TODO 从worker列表删除
				log.Info().Str("gpool name", p.Name).
					Int("id", w.id).
					Msgf("worker exit")
				return
			}
		}
	}
}

func newWorker(id int, p *gpool) *worker {
	now := time.Now()
	w := &worker{
		id:          id,
		JobBus:      make(chan Job, 1),
		stopChan:    make(chan struct{}, 1),
		lastUseTime: &now,
	}
	go w.run(p)
	return w
}

type gpool struct {
	Name        string
	Num         int
	MaxNum      int
	RecycleTime int
	Workers     chan *worker
	WorkerList  []*worker
	JobCacheBus chan Job
	JobCache    list.List
	incrId      int
}

func (p *gpool) jobWrapper(job Job) Job {
	return func() {
		defer util.BtRecover("jobRecoverWrapper")
		job()
	}
}

func (p *gpool) run() {
	defer util.BtRecover("gpool.run")
	ticker_30 := time.NewTicker(time.Second * 30)
	for {
		if p.JobCache.Len() <= 0 {
			select {
			case job := <-p.JobCacheBus:
				p.JobCache.PushBack(job)
				log.Error().Msgf("gpool push back: %v, cache len: %v",
					p.Name, p.JobCache.Len())
			case now := <- ticker_30.C:
				if len(p.WorkerList) > p.Num {
					count := len(p.WorkerList) - p.Num
					for i, w := range p.WorkerList {
						if nil != w {
							last := w.getLastUseTime()
							if last != nil {
								if now.Sub(*last) > time.Second * time.Duration(p.RecycleTime) {
									w.stop()
									p.WorkerList = append(p.WorkerList[:i], p.WorkerList[i+1:]...)
									count--
									if 0 >= count {
										break
									}
								}
							}
						}
					}
				}
			}
		} else {
			//TODO 这里做动态扩容
			select {
			case job := <-p.JobCacheBus:
				p.JobCache.PushBack(job)
				log.Error().Msgf("gpool push back: %v, cache len: %v",
					p.Name, p.JobCache.Len())
			case w := <-p.Workers:
				j := p.JobCache.Front()
				p.JobCache.Remove(j)
				job, ok := j.Value.(Job)
				if ok {
					w.JobBus <- p.jobWrapper(job)
				}
				log.Error().Msgf("gpool remove: %v, cache len: %v",
					p.Name, p.JobCache.Len())
			}
		}
	}
}

func (p *gpool) Post(job Job) {
	select {
	case w := <-p.Workers:
		w.JobBus <- p.jobWrapper(job)
	default:
		p.JobCacheBus <- job
		log.Error().Msgf("gpool %v block, cache job", p.Name)
	}
}

func NewGpool(name string, num, maxNum, recycleTime int) Pool {
	pool := &gpool{
		Name:        name,
		Num:         num,
		MaxNum:      maxNum,
		RecycleTime: recycleTime,
		Workers:     make(chan *worker),
		WorkerList:  make([]*worker, 0),
		JobCacheBus: make(chan Job),
		incrId:      0,
	}
	for ; pool.incrId < num; pool.incrId++ {
		w := newWorker(pool.incrId, pool)
		pool.WorkerList = append(pool.WorkerList, w)
	}
	go pool.run()
	return pool
}