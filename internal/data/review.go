package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"review-server/internal/biz"
	"review-server/internal/data/model"
)

type reviewRepo struct {
	data *Data
	log  *log.Helper
}

func NewReviewRepo(data *Data, logs log.Logger) biz.ReviewRepo {
	return &reviewRepo{
		data: data,
		log:  log.NewHelper(logs),
	}
}

func (rp *reviewRepo) Save(ctx context.Context, info *model.ReviewInfo) (*model.ReviewInfo, error) {
	if err := rp.data.query.WithContext(ctx).ReviewInfo.Save(info); err != nil {
		rp.log.Fatal("[data]create review item is error ")
		return nil, err
	}
	return info, nil
}

func (rp *reviewRepo) FindByOrder(ctx context.Context, order_id int64) ([]*model.ReviewInfo, error) {
	q := rp.data.query
	find, err := q.WithContext(ctx).ReviewInfo.Where(q.ReviewInfo.OrderID.Eq(order_id)).Find()
	if err != nil {
		rp.log.Fatal("[data] FindByorder has error")
		return nil, err
	}
	return find, nil
}
