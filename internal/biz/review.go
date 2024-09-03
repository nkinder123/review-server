package biz

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"review-server/internal/data/model"
	"review-server/pkg"
	"strings"
	"time"
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
	UpdateAppealInfoOp(context.Context, *model.ReviewAppealInfo) error
	Getdata(context.Context, *FindStruct) ([]*MyReviewInfo, error)
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
	if err != nil && err.Error() != "record not found" {
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
			uc.Log.Info("[biz]the appeal item is updating ")
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
		info.AppealID = pkg.Gen()
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
	err = uc.reviewRepo.UpdateAppealInfoOp(ctx, info)
	if err != nil {
		uc.Log.Errorf("[biz]update op item  has error")
		return nil, err
	}
	return info, nil
}

type FindStruct struct {
	StoreId int64
	Page    int32
	Limit   int32
}

func (uc *ReviewUsecase) FindReveiwBySotre(ctx context.Context, storeId int64, page int32, limite int32) ([]*MyReviewInfo, error) {
	if page < 0 {
		return nil, errors.New("page is invalidate")
	}
	dataReq := &FindStruct{
		StoreId: storeId,
		Page:    page,
		Limit:   limite,
	}
	reviewes, err := uc.reviewRepo.Getdata(ctx, dataReq)
	if err != nil {
		uc.Log.Errorf("[biz]-->[data] find review item by storeId has error")
		return nil, err
	}
	for _, index := range reviewes {
		fmt.Printf("[biz]marshal review info:%#v\n", index)
	}
	return reviewes, nil
}

// 传进来的是"2024-08-31 09:33:41",在marshal到reviewinfo的时候是timeTime结构
// 和go的的不一样
type MyReviewInfo struct {
	*model.ReviewInfo
	ID           int64  `json:"id,string"`
	OrderID      int64  `json:"order_id,string"`
	ReviewID     int64  `json:"review_id,string"`
	SkuID        int64  `json:"sku_id,string"`
	SpuID        int64  `json:"spu_id,string"`
	StoreID      int64  `json:"store_id,string"`
	UserID       int64  `json:"user_id,string"`
	CreateAt     MyTime `json:"create_at"`
	UpdateAt     MyTime `json:"update_at"`
	Version      int32  `json:"version,string"`
	Score        int32  `json:"score,string"`
	ServiceScore int32  `json:"service_score,string"`
	ExpressScore int32  `json:"express_score,string"`
	HasMedia     int32  `json:"has_media,string"`
	Anonymous    int32  `json:"anonymous,string"`
	Status       int32  `json:"status,string"`
	IsDefault    int32  `json:"is_default,string"`
	HasReply     int32  `json:"has_reply,string"`
}

type MyTime time.Time

func (t *MyTime) UnmarshalJSON(d []byte) error {
	trim := strings.Trim(string(d), `"`)
	parse, err := time.Parse(time.DateTime, trim)
	if err != nil {
		return err
	}
	*t = MyTime(parse)
	return nil
}

//CMD ["go run main.go", "-conf", "/data/conf"]
