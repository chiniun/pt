package data

import (
	"context"
	"pt/internal/biz/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
)

type AgentDeny struct {
	data *Data
	log  *log.Helper
}

func NewAgentDeny(data *Data, logger log.Logger) *AgentDeny {
	return &AgentDeny{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (o *AgentDeny) Get(ctx context.Context) (*model.AgentAllowedException, error) {

	allowAgent := new(model.AgentAllowedException)
	err := o.data.DB.WithContext(ctx).First(allowAgent).Error
	if err != nil {
		return nil, errors.Wrap(err, "Create")
	}
	return allowAgent, nil
}

func (o *AgentDeny) GetList(ctx context.Context) ([]*model.AgentAllowedException, error) {

	allowAgentList := make([]*model.AgentAllowedException, 0)
	err := o.data.DB.WithContext(ctx).Find(&allowAgentList).Error
	if err != nil {
		return nil, errors.Wrap(err, "GetList")
	}
	return allowAgentList, nil
}
