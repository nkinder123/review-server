package data

import (
	"context"
	"fmt"
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

func (br *reviewRepo) CreateAppeal(ctx context.Context, info *model.ReviewAppealInfo) error {
	if err := br.data.query.WithContext(ctx).ReviewAppealInfo.Save(info); err != nil {
		return err
	}
	return nil
}

func (br *reviewRepo) FindAppealInfoByReviewId(ctx context.Context, reviewId int64) (*model.ReviewAppealInfo, error) {
	reveiwinfo, err := br.data.query.WithContext(ctx).ReviewAppealInfo.Where(br.data.query.ReviewAppealInfo.ReviewID.Eq(reviewId)).First()
	return reveiwinfo, err
}

func (br *reviewRepo) UpdateAppealInfo(ctx context.Context, info *model.ReviewAppealInfo) error {
	br.log.Info("[data update appeal info]")
	_, err := br.data.query.WithContext(ctx).ReviewAppealInfo.Where(br.data.query.ReviewAppealInfo.ReviewID.Eq(info.ReviewID)).Updates(
		map[string]interface{}{
			"reason":  info.Reason,
			"content": info.Content,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (br *reviewRepo) UpdateAppealInfoOp(ctx context.Context, info *model.ReviewAppealInfo) error {
	fmt.Printf("info:", info.Status)
	_, err := br.data.query.WithContext(ctx).ReviewAppealInfo.Where(br.data.query.ReviewAppealInfo.AppealID.Eq(info.AppealID)).Updates(
		map[string]interface{}{
			"status":     info.Status,
			"op_remarks": info.OpRemarks,
			"op_user":    info.OpUser,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (br *reviewRepo) FindAppealInfoByAppealId(ctx context.Context, appealId int64) (*model.ReviewAppealInfo, error) {
	reveiwinfo, err := br.data.query.WithContext(ctx).ReviewAppealInfo.Where(br.data.query.ReviewAppealInfo.AppealID.Eq(appealId)).First()
	if err != nil {
		return nil, err
	}
	return reveiwinfo, nil
}
