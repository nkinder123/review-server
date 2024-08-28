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
	CreateAppeal(context.Context, *model.ReviewAppealInfo) error
	FindAppealInfoByReviewId(context.Context, int64) (*model.ReviewAppealInfo, error)
	UpdateAppealInfo(context.Context, *model.ReviewAppealInfo) error
	FindAppealInfoByAppealId(context.Context, int64) (*model.ReviewAppealInfo, error)
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

// 商家appeal
func (uc *ReviewUsecase) CreateAppeal(ctx context.Context, info *model.ReviewAppealInfo) error {
	//判断之前的是否有过appeal
	review_id := info.ReviewID
	appealinfo, err := uc.reviewRepo.FindAppealInfoByReviewId(ctx, review_id)
	if err != nil {
		uc.Log.Errorf("[biz]find the appeal by review_id has error")
		return nil
	}
	//是否水平越权
	review, err := uc.reviewRepo.SearchReview(ctx, info.ReviewID)
	if err != nil {
		uc.Log.Errorf("[biz]search reveiw info has error")
		return nil
	}
	if review.StoreID != info.StoreID {
		uc.Log.Errorf("[biz]storId不一致，发生水平越权")
		return errors.New("发生水平越权")
	}
	if appealinfo != nil {
		//1。有，判断status是否>10
		if appealinfo.Status == 10 {
			uc.Log.Info("[biz]the appeal itme is updating ")
			info.AppealID = pkg.Gen()
			err = uc.reviewRepo.UpdateAppealInfo(ctx, info)
			if err != nil {
				uc.Log.Errorf("[biz]update appeal has error")
				return err
			}
		} else {
			uc.Log.Errorf("[biz]the appeal has op,don't repeated op")
			return errors.New("appeal has oped")
		}
	} else {
		//2 没有，直接create
		uc.Log.Info("[biz]create appeal item")
		if err := uc.reviewRepo.CreateAppeal(ctx, info); err != nil {
			uc.Log.Errorf("[biz]create appeal item has error")
			return err
		}
	}
	return nil
}

// 评价
func (uc *ReviewUsecase) OpReAppeal(ctx context.Context, info *model.ReviewAppealInfo) (*model.ReviewAppealInfo, error) {
	appealinfo, err := uc.reviewRepo.FindAppealInfoByAppealId(ctx, info.AppealID)
	if err != nil {
		uc.Log.Errorf("[biz]find appeal by appealid has error")
		return nil, err
	}
	if appealinfo.Status > 10 {
		uc.Log.Errorf("[biz]the appeal has done")
		return nil, errors.New("the appeal has done")
	}
	err = uc.reviewRepo.UpdateAppealInfo(ctx, info)
	if err != nil {
		uc.Log.Errorf("[biz]update op item  has error")
		return nil, err
	}
	return info, nil
}
