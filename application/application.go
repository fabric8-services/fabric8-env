package application

import "github.com/fabric8-services/fabric8-env/environment"

type Application interface {
	Environments() environment.Repository
}

type Transaction interface {
	Application
	Commit() error
	Rollback() error
}

type DB interface {
	Application
	BeginTransaction() (Transaction, error)
}
