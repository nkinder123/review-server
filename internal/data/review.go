package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"review-server/internal/biz"
	"review-server/internal/data/model"
	"review-server/internal/data/query"
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
		rp.log.Errorf("[data] FindByorder has error")
		return nil, err
	}
	return find, nil
}

func (br *reviewRepo) CreateReply(ctx context.Context, replyInfo *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error) {
	err := br.data.query.Transaction(func(tx *query.Query) error {
		if err := tx.ReviewReplyInfo.WithContext(ctx).Save(replyInfo); err != nil {
			br.log.Errorf("[data] save reviewreply has error ")
			return err
		}
		if _, err := tx.ReviewInfo.WithContext(ctx).Where(tx.ReviewInfo.ReviewID.Eq(replyInfo.ReviewID)).Update(tx.ReviewInfo.HasReply, 0); err != nil {
			br.log.Errorf("[data] update reviewinfo.hasrely is error ")
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return replyInfo, err
}

func (br *reviewRepo) SearchReview(ctx context.Context, review_id int64) (*model.ReviewInfo, error) {
	first, err := br.data.query.ReviewInfo.WithContext(ctx).Where(br.data.query.ReviewInfo.ReviewID.Eq(review_id)).First()
	if err != nil {
		br.log.Errorf("[data] search reviewinfo by id  has error ")
		return nil, err
	}
	return first, nil
}
