package service

import (
	"context"
	"errors"
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
