package environment

import (
	"context"
	"time"

	"github.com/fabric8-services/fabric8-common/gormsupport"
	"github.com/fabric8-services/fabric8-common/log"
	"github.com/goadesign/goa"
	"github.com/jinzhu/gorm"
	errs "github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type Environment struct {
	gormsupport.Lifecycle
	ID            *uuid.UUID `sql:"type:uuid default uuid_generate_v4()" gorm:"primary_key"`
	Name          *string
	Type          *string
	SpaceID       *uuid.UUID `sql:"type:uuid"`
	NamespaceName *string
	ClusterURL    *string
}

func (e Environment) TableName() string {
	return "environments"
}

type Repository interface {
	Create(ctx context.Context, env *Environment) (*Environment, error)
	List(ctx context.Context, spaceID uuid.UUID) ([]*Environment, error)
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
	defer goa.MeasureSince([]string{"goa", "db", "environment", "create"}, time.Now())

	err := r.db.Create(env).Error
	if err != nil {
		log.Error(ctx, map[string]interface{}{"err": err},
			"unable to create the environment")
		return nil, errs.WithStack(err)
	}

	return env, nil
}

func (r *GormRepository) List(ctx context.Context, spaceID uuid.UUID) ([]*Environment, error) {
	var rows []*Environment

	err := r.db.Model(&Environment{}).Where("space_id = ?", spaceID).Find(&rows).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(ctx, map[string]interface{}{"space_id": spaceID, "err": err},
			"unable to list the environments")
		return nil, errs.WithStack(err)
	}

	return rows, nil
}
