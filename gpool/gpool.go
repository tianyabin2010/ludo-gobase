package gpool

import (
	"phabricator.ushow.media/source/ludo-gobase/util"
	"container/list"
	"github.com/rs/zerolog/log"
)

type Job func()

type worker struct {
	JobBus chan Job
}

func (w *worker) run() {
	defer util.BtRecover("worker.Run")
	for {
		select {
		case job := <-w.JobBus:
			job()
		}
	}
}

func newWorker() *worker {
	w := &worker{
		JobBus: make(chan Job, 1),
	}
	go w.run()
	return w
}

type gpool struct {
	Name        string
	Num         int
	Workers     chan *worker
	JobCacheBus chan Job
	JobCache    list.List
}

func (p *gpool) jobWrapper(w *worker, job Job) Job {
	return func() {
		defer util.BtRecover("jobRecoverWrapper")
		job()
		p.Workers <- w
	}
}

func (p *gpool) run() {
	defer util.BtRecover("gpool.run")
	for {
		if p.JobCache.Len() <= 0 {
			select {
			case job := <-p.JobCacheBus:
				p.JobCache.PushBack(job)
				log.Error().Msgf("gpool push back: %v, cache len: %v",
					p.Name, p.JobCache.Len())
			}
		} else {
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
					w.JobBus <- p.jobWrapper(w, job)
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
		w.JobBus <- p.jobWrapper(w, job)
	default:
		p.JobCacheBus <- job
		log.Error().Msgf("gpool %v block, cache job", p.Name)
	}
}

func NewGpool(name string, num int) *gpool {
	pool := &gpool{
		Name:        name,
		Num:         num,
		Workers:     make(chan *worker, num),
		JobCacheBus: make(chan Job),
	}
	for i := 0; i < num; i++ {
		w := newWorker()
		pool.Workers <- w
	}
	go pool.run()
	return pool
}
