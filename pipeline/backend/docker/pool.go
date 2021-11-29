package docker

// import (

// )
//
// // Pool manages a pool of Docker clients.
// type Pool struct {
// 	queue chan (backend.Engine)
// }
//
// // NewPool returns a Pool.
// func NewPool(engines ...backend.Engine) *Pool {
// 	return &Pool{
// 		queue: make(chan backend.Engine, len(engines)),
// 	}
// }
//
// // Reserve requests the next available Docker client in the pool.
// func (p *Pool) Reserve(c context.Context) backend.Engine {
// 	select {
// 	case <-c.Done():
// 	case docker := <-p.queue:
// 		return docker
// 	}
// 	return nil
// }
//
// // Release releases the Docker client back to the pool.
// func (p *Pool) Release(docker backend.Engine) {
// 	p.queue <- docker
// }

// pool := docker.Pool(
//   docker.FromEnvironmentMust(),
//   docker.FromEnvironmentMust(),
//   docker.FromEnvironmentMust(),
//   docker.FromEnvironmentMust(),
// )
//
// client := pool.Reserve()
// defer pool.Release(client)
