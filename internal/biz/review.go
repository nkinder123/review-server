package biz

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"review-server/internal/data/model"
	"review-server/pkg"
)

type ReviewRepo interface {
	Save(context.Context, *model.ReviewInfo) (*model.ReviewInfo, error)
	FindByOrder(context.Context, int64) ([]*model.ReviewInfo, error)
}

type ReviewUsecase struct {
	reviewRepo ReviewRepo
	Log        *log.Helper
}

func NewReviewUserCase(reviewrepo ReviewRepo, logs log.Logger) *ReviewUsecase {
	return &ReviewUsecase{
		reviewRepo: reviewrepo,
		Log:        log.NewHelper(logs),
	}
}

func (rc *ReviewUsecase) CreateReview(ctx context.Context, info *model.ReviewInfo) (*model.ReviewInfo, error) {
	//判断这个评论之前有人评价过吗？
	order, err := rc.FindByOrder(ctx, info.OrderID)
	if err != nil {
		rc.Log.Errorf("[biz] find by orderId in createReview has error")
		return nil, err
	}
	if len(order) > 0 {
		rc.Log.Errorf("[biz] order has reviewed")
		return nil, ERRORS_ORDER_IS_REVIEWED
	}
	//雪花算法生成分布式id
	review_id := pkg.Gen()
	info.ReviewID = review_id
	item, err := rc.reviewRepo.Save(ctx, info)
	if err != nil {
		rc.Log.Errorf("[biz] create review item has error")
		return nil, err
	}
	return item, nil
}

func (rc *ReviewUsecase) FindByOrder(ctx context.Context, order_id int64) ([]*model.ReviewInfo, error) {
	order, err := rc.reviewRepo.FindByOrder(ctx, order_id)
	if err != nil {
		rc.Log.Errorf("[biz] find order has error")
		return nil, errors.New("find order has error")
	}
	return order, nil
}
