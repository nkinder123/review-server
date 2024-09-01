package service

import (
	"context"
	"errors"
	"fmt"
	"review-server/internal/biz"
	"review-server/internal/data/model"

	pb "review-server/api/review/v1"
)

type ReviewService struct {
	pb.UnimplementedReviewServer
	uc *biz.ReviewUsecase
}

func NewReviewService(reviewUc *biz.ReviewUsecase) *ReviewService {
	return &ReviewService{
		uc: reviewUc,
	}
}

func (s *ReviewService) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.CreateReviewReply, error) {
	reqModel := &model.ReviewInfo{
		UserID:       req.UserID,
		OrderID:      req.OrderID,
		Score:        req.Score,
		ServiceScore: req.ServiceScore,
		ExpressScore: req.ExpressScore,
		Content:      req.Content,
		PicInfo:      req.PicInfo,
		VideoInfo:    req.VideoInfo,
		Anonymous:    req.Anonymous,
	}
	review, err := s.uc.CreateReview(ctx, reqModel)
	if err != nil {
		s.uc.Log.Errorf("[service] the create review has error:%s", err)
		if errors.Is(err, biz.ERRORS_ORDER_IS_REVIEWED) {
			return nil, pb.ErrorOrderHasReview("order has create")
		}
		return nil, pb.ErrorCreateReviewHasError("create review has error")
	}
	return &pb.CreateReviewReply{ReviewID: review.ReviewID}, nil
}
func (s *ReviewService) UpdateReview(ctx context.Context, req *pb.UpdateReviewRequest) (*pb.UpdateReviewReply, error) {
	return &pb.UpdateReviewReply{}, nil
}
func (s *ReviewService) DeleteReview(ctx context.Context, req *pb.DeleteReviewRequest) (*pb.DeleteReviewReply, error) {
	return &pb.DeleteReviewReply{}, nil
}
func (s *ReviewService) GetReview(ctx context.Context, req *pb.GetReviewRequest) (*pb.GetReviewReply, error) {
	return &pb.GetReviewReply{}, nil
}
func (s *ReviewService) ListReview(ctx context.Context, req *pb.ListReviewRequest) (*pb.ListReviewReply, error) {
	return &pb.ListReviewReply{}, nil
}
func (s *ReviewService) ReplyReview(ctx context.Context, req *pb.ReplyReviewReq) (*pb.ReplyReviewReply, error) {
	reqs := &model.ReviewReplyInfo{
		ReviewID:  req.ReviewId,
		StoreID:   req.StoreId,
		Content:   req.Content,
		PicInfo:   req.PicInfo,
		VideoInfo: req.VideoInfo,
	}
	reply, err := s.uc.CreateReply(ctx, reqs)
	if err != nil {
		return nil, err
	}
	return &pb.ReplyReviewReply{ReplyId: reply.ReplyID}, nil
}

func (s *ReviewService) CreateAppeal(ctx context.Context, req *pb.CreateAppealRequest) (*pb.CreateAppealReply, error) {
	s.uc.Log.Info("[rpc create appeal info has arrive]")
	reqs := &model.ReviewAppealInfo{
		AppealID:  req.AppealId,
		ReviewID:  req.ReviewId,
		StoreID:   req.StoreId,
		Reason:    req.Reason,
		Content:   req.Content,
		PicInfo:   req.PicInfo,
		VideoInfo: req.VideoInfo,
		Status:    10,
	}

	if err := s.uc.CreateAppeal(ctx, reqs); err != nil {
		return nil, err
	}
	return &pb.CreateAppealReply{AppealId: reqs.AppealID}, nil
}

// 评价
func (s *ReviewService) OpReAppeal(ctx context.Context, req *pb.OpCreateAppealRequest) (*pb.OpCreateAppealReply, error) {
	reqs := &model.ReviewAppealInfo{
		AppealID:  req.AppealId,
		Status:    req.Status,
		OpUser:    req.OpUser,
		OpRemarks: req.OpRemark,
	}
	fmt.Printf("info:%#v", reqs.Status)
	appeal, err := s.uc.OpReAppeal(ctx, reqs)
	if err != nil {
		return nil, err
	}
	return &pb.OpCreateAppealReply{AppealId: appeal.AppealID}, nil
}

// elastcisearch find review item by storeId
func (s *ReviewService) SearchReviewByStoreId(ctx context.Context, req *pb.SearchReviewRequest) (*pb.SearchReveiwReply, error) {
	_, err := s.uc.FindReveiwBySotre(ctx, req.StoreId, req.Page, req.Limit)
	if err != nil {
		fmt.Printf("[service]-->[biz]find Review BY storeId has error")
		return nil, nil
	}
	return &pb.SearchReveiwReply{}, nil
}
