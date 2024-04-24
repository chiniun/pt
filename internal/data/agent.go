package data

import (
	"context"
	"pt/internal/biz/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
)

type AgentAllowed struct {
	data *Data
	log  *log.Helper
}

func NewAgentAllowed(data *Data, logger log.Logger) *AgentAllowed {
	return &AgentAllowed{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (o *AgentAllowed) Create(ctx context.Context, allowAgent *model.AgentAllowedFamily) error {
	err := o.data.DB.WithContext(ctx).Create(allowAgent).Error
	if err != nil {
		return errors.Wrap(err, "Create")
	}
	return nil
}

func (o *AgentAllowed) Get(ctx context.Context) (*model.AgentAllowedFamily, error) {

	allowAgent := new(model.AgentAllowedFamily)
	err := o.data.DB.WithContext(ctx).First(allowAgent).Error
	if err != nil {
		return nil, errors.Wrap(err, "Create")
	}
	return allowAgent, nil
}

func (o *AgentAllowed) GetList(ctx context.Context) ([]*model.AgentAllowedFamily, error) {

	allowAgentList := make([]*model.AgentAllowedFamily, 0)
	err := o.data.DB.WithContext(ctx).Find(&allowAgentList).Error
	if err != nil {
		return nil, errors.Wrap(err, "GetList")
	}
	return allowAgentList, nil
}
