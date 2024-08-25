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
	SearchReview(context.Context, int64) (*model.ReviewInfo, error)
	CreateReply(context.Context, *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error)
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
	order, err := rc.reviewRepo.FindByOrder(ctx, info.OrderID)
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

func (uc *ReviewUsecase) CreateReply(ctx context.Context, info *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error) {
	//参数校验
	//1.1 是否已经有回复了
	//1.2 水平越权
	review, err := uc.reviewRepo.SearchReview(ctx, info.ReviewID)
	if err != nil {
		return nil, err
	}
	if review.HasReply == 1 {
		return nil, errors.New("the business has reply,deny repeated reply ")
	}
	if review.StoreID != info.StoreID {
		return nil, errors.New("the replyId  is not the direct business")
	}
	info.ReplyID = pkg.Gen()
	reply, err := uc.reviewRepo.CreateReply(ctx, info)
	if err != nil {
		return nil, err
	}
	return reply, err
}

func (uc *ReviewUsecase) SearchReview(ctx context.Context, info *model.ReviewReplyInfo) (*model.ReviewInfo, error) {
	if info == nil {
		return nil, errors.New("[biz] reviewReply is null ")
	}
	review, err := uc.reviewRepo.SearchReview(ctx, info.ReviewID)
	if err != nil {
		return nil, errors.New("[biz] search review info has error")
	}
	return review, err

}
