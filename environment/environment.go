package environment

import (
	"context"

	"github.com/fabric8-services/fabric8-common/gormsupport"
	"github.com/jinzhu/gorm"
	errs "github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type Environment struct {
	gormsupport.Lifecycle
	ID            *uuid.UUID
	Name          *string
	Type          *string
	NamespaceName *string
	ClusterURL    *string
}

type Repository interface {
	Create(ctx context.Context, env *Environment) (*Environment, error)
	List(ctx context.Context) ([]Environment, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{
		db: db,
	}
}

func (r *GormRepository) Create(ctx context.Context, env *Environment) (*Environment, error) {
	r.db.Create(env)
	return env, nil
}

func (r *GormRepository) List(ctx context.Context) ([]Environment, error) {
	var rows []Environment
	err := r.db.Model(&Environment{}).Find(&rows).Error
	if err != nil {
		return nil, errs.WithStack(err)
	}
	return rows, nil
}
