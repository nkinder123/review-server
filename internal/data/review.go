package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"review-server/internal/biz"
	"review-server/internal/data/model"
	"review-server/internal/data/query"
	"strconv"
	"strings"
	"time"
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

func (br *reviewRepo) Save(ctx context.Context, info *model.ReviewInfo) (*model.ReviewInfo, error) {
	if err := br.data.query.WithContext(ctx).ReviewInfo.Save(info); err != nil {
		br.log.Fatal("[data]create review item is error ")
		return nil, err
	}
	return info, nil
}

func (br *reviewRepo) FindByOrder(ctx context.Context, order_id int64) ([]*model.ReviewInfo, error) {
	q := br.data.query
	find, err := q.WithContext(ctx).ReviewInfo.Where(q.ReviewInfo.OrderID.Eq(order_id)).Find()
	if err != nil {
		br.log.Errorf("[data] FindByorder has error")
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
	fmt.Printf("info:%v", info.Status)
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

func (br *reviewRepo) FindReviewBydStoreId(ctx context.Context, info *biz.FindStruct) ([]*biz.MyReviewInfo, error) {
	//happened during the Search query execution: http: no Host in request URL
	reply, err := br.data.elasticClient.Search().
		Index("review").
		From(int(info.Page)).Size(int(info.Limit)).
		Query(&types.Query{
			Bool: &types.BoolQuery{
				Filter: []types.Query{
					{Term: map[string]types.TermQuery{
						"store_id": {Value: info.StoreId},
					}},
				},
			},
		}).Do(ctx)
	if err != nil {
		br.log.Errorf("[data]elastic search review by storeId has error:%#v", err.Error())
		return nil, err
	}
	myinfo := make([]*biz.MyReviewInfo, 0, reply.Hits.Total.Value)
	for _, item := range reply.Hits.Hits {
		tmp_item := &biz.MyReviewInfo{}
		if err = json.Unmarshal(item.Source_, tmp_item); err != nil {
			br.log.Errorf("unmrshal has error:%#v", err)
			continue
		}
		myinfo = append(myinfo, tmp_item)
	}
	return myinfo, nil
}

func (br *reviewRepo) Getdata(ctx context.Context, info *biz.FindStruct) ([]*biz.MyReviewInfo, error) {
	single, err := br.GetdataBySingle(ctx, info)
	reply := new(types.HitsMetadata)
	if err = json.Unmarshal(single, reply); err != nil {
		br.log.Errorf("unmarshal hits has error")
		return nil, err
	}
	myinfo := make([]*biz.MyReviewInfo, 0, reply.Total.Value)
	for _, item := range reply.Hits {
		tmp_item := &biz.MyReviewInfo{}
		if err = json.Unmarshal(item.Source_, tmp_item); err != nil {
			br.log.Errorf("unmrshal has error:%#v", err)
			continue
		}
		myinfo = append(myinfo, tmp_item)
	}
	return myinfo, nil
}

func (br *reviewRepo) GetdataBySingle(ctx context.Context, info *biz.FindStruct) ([]byte, error) {
	//封装key--+---使用singleflight
	key := fmt.Sprintf("review:%d:%d:%d", info.StoreId, info.Page, info.Limit)
	g := new(singleflight.Group)
	v, err, _ := g.Do(key, func() (interface{}, error) {
		data, err := br.GetRedisByKey(ctx, key)
		if err == nil {
			br.log.Info("redis has data")
			return data, nil
		}
		if errors.Is(err, redis.Nil) {
			//查询elasticsearch并将elasticsearch的data存入redis中
			data, err = br.searchElastic(ctx, key)
			if err == nil {
				fmt.Printf("hello")
				if err = br.SetRedisKey(ctx, key, data); err != nil {
					fmt.Printf("redis set has error---------->")
					br.log.Errorf("redis set key-value has error")
					return nil, err
				}
				return data, nil
			}
			return nil, err
		}
		return nil, errors.New("redis connect has error")
	})
	//fmt.Printf("v:%v,err:%v,shared:%v\n", v, err, shared)
	if err != nil {
		br.log.Errorf("single has error")
		return nil, err
	}
	return v.([]byte), nil
}

func (br *reviewRepo) GetRedisByKey(ctx context.Context, key string) ([]byte, error) {
	//查找数据并by key
	br.log.Info("[data]get redis by key ")
	return br.data.rdb.Get(ctx, key).Bytes()
}

func (br *reviewRepo) SetRedisKey(ctx context.Context, key string, data []byte) error {
	//设置
	fmt.Printf("__--------data:%v", data)
	fmt.Printf("[data]set key is here:key:%s\n", key)
	set := br.data.rdb.Set(ctx, key, data, time.Hour*12)
	fmt.Printf("------set-------:%v", set)
	return set.Err()
}

func (br *reviewRepo) searchElastic(ctx context.Context, key string) ([]byte, error) {
	split := strings.Split(key, ":")
	if len(split) < 4 {
		br.log.Errorf("redis key is invalidate")
		return nil, errors.New("the redis key is invalidate")
	}
	re_index, store_id := split[0], split[1]
	page, err := strconv.Atoi(split[2])
	if err != nil {
		return nil, errors.New("change from str to int has error")
	}
	limit, err := strconv.Atoi(split[3])
	if err != nil {
		return nil, errors.New("change from str to int has error")
	}
	do, err := br.data.elasticClient.Search().
		Index(re_index).From(page).Size(limit).
		Query(&types.Query{
			Bool: &types.BoolQuery{
				Filter: []types.Query{
					{Term: map[string]types.TermQuery{
						"store_id": {Value: store_id},
					}},
				},
			},
		}).Do(ctx)
	if err != nil {
		br.log.Errorf("search review by stored Id has error")
		return nil, err
	}
	return json.Marshal(do.Hits)
}
