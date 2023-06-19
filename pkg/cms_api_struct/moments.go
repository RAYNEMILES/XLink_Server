package cms_api_struct

import "Open_IM/pkg/base_info"

type Moment struct {
	MomentID              string                                `json:"moment_id"`
	MCreateTime           int64                                 `json:"m_create_time"`
	MContentText          string                                `json:"m_content_text"`
	MContentImagesArray   string                                `json:"m_content_images_array"`
	MContentVideosArray   string                                `json:"m_content_videos_array"`
	MContentImagesArrayV2 []*base_info.MomentImageRequestObject `json:"mContentImagesArrayV2"`
	MContentVideosArrayV2 []*base_info.MomentVideoRequestObject `json:"mContentVideosArrayV2"`
	MContentThumbnilArray string                                `json:"m_content_thumbnil_array"`
	UserID                string                                `json:"user_id"`
	UserName              string                                `json:"user_name"`
}

type GetMomentRes struct {
	Moment
	ArticleID           int64           `json:"article_id"`
	OrignalID           string          `json:"orignal_id"`
	OriginalCreatorID   string          `json:"original_creator_id"`
	OriginalCreatorName string          `json:"original_creator_name"`
	MomentType          int8            `json:"moment_type"`
	Privacy             int8            `json:"privacy"`
	Status              int8            `json:"status"`
	Interests           []*InterestType `json:"interests"`
}

type MomentDetailRes struct {
	Moment
	OriginalCreatorName string `json:"original_creator_name"`
	OrignalCreatorID    string `json:"orignal_creator_id"`
	OrignalID           string `json:"orignal_id"`
	Privacy             int8   `json:"privacy"`
	CommentCtl          int32  `json:"comment_ctl"`
	LastLoginIp         string `json:"last_login_ip"`
	MLikesCount         int32  `json:"m_likes_count"`
	MCommentsCount      int32  `json:"m_comments_count"`
	MRepostCount        int64  `json:"m_repost_count"`
}

type MomentComment struct {
	Moment
	CommentID         string `json:"comment_id"`
	PublishAccount    string `json:"publish_account"`
	PublishName       string `json:"publish_name"`
	CommentContent    string `json:"comment_content"`
	CreateBy          string `json:"create_by"`
	CreateTime        int64  `json:"create_time"`
	CommentReplies    int64  `json:"comment_replies"`
	LikeCounts        int64  `json:"like"`
	Privacy           int32  `json:"privacy"`
	Status            int8   `json:"status"`
	ReplyCommentId    string `json:"reply_comment_id"`
	CommentParentId   string `json:"comment_parent_id"`
	CommentedUseID    string `json:"commented_use_id"`
	CommentedUserName string `json:"commented_user_name"`
}

type MomentsLike struct {
	Moment
	Privacy         int32  `json:"privacy"`
	Status          int8   `json:"status"`
	Account         string `json:"account"`
	AccountNickname string `json:"account_nickname"`
	CreateTime      int64  `json:"create_time"`
}

type RepostMoment struct {
}

type GetMomentsRequest struct {
	RequestPagination
	Account      string `form:"account" binding:"omitempty"`
	Privacy      int8   `form:"privacy" binding:"omitempty,numeric"`
	ContentType  int8   `form:"content_type" binding:"omitempty,numeric"`
	Content      string `form:"content" binding:"omitempty"`
	MediaType    int8   `form:"media_type" binding:"omitempty,numeric"`
	IsReposted   int8   `form:"is_reposted" binding:"omitempty,numeric"`
	OriginalUser string `form:"original_user" binding:"omitempty"`
	IsBlocked    int8   `form:"is_blocked" binding:"omitempty,numeric"`
	StartTime    string `form:"start_time" binding:"omitempty,numeric"`
	EndTime      string `form:"end_time" binding:"omitempty,numeric"`
	OrderBy      string `form:"order_by" binding:"omitempty,oneof=start_time:asc start_time:desc"`
}

type GetMomentsResponse struct {
	ResponsePagination
	Moments     []*GetMomentRes `json:"moments"`
	MomentsNums int64           `json:"moments_nums"`
}

type DeleteMomentsRequest struct {
	Moments    []string `json:"moments"`
	ArticleIDs []int64  `json:"article_ids"`
}

type AlterMomentRequest struct {
	MomentId              string                               `json:"moment_id"`
	Privacy               int32                                `json:"privacy"`
	IsReposted            bool                                 `json:"is_reposted"`
	Content               string                               `json:"content"`
	MContentImagesArrayV2 []base_info.MomentImageRequestObject `json:"m_content_images_array_v2"`
	MContentVideosArrayV2 []base_info.MomentVideoRequestObject `json:"m_content_videos_array_v2"`
}

type AlterMomentResponse struct {
}

type ChangeMomentStatusRequest struct {
	MomentIds []string `json:"moment_ids"`
	Status    int8     `json:"status"`
}

type ChangeMomentStatusResponse struct {
}

type ModifyVisibilityRequest struct {
	MomentIds []string `json:"moment_ids"`
	Privacy   int32    `json:"privacy"`
}

type ModifyVisibilityResponse struct {
}

type GetMomentDetailsRequest struct {
	RequestPagination
	Account     string `form:"account" binding:"omitempty"`
	Privacy     int8   `form:"privacy" binding:"omitempty,numeric"`
	MediaType   int8   `form:"media_type" binding:"omitempty,numeric"`
	ContentType int8   `form:"content_type" binding:"omitempty,numeric"`
	Content     string `form:"content" binding:"omitempty"`
	MomentID    string `form:"moment_id" binding:"omitempty"`
	OriginalID  string `form:"moment_id" binding:"omitempty"`
	RepostUser  string `form:"repost_user" binding:"omitempty"`
	StartTime   string `form:"start_time" binding:"omitempty,numeric"`
	EndTime     string `form:"end_time" binding:"omitempty,numeric"`
	OrderBy     string `form:"order_by" binding:"omitempty,oneof=start_time:asc start_time:desc"`
}

type GetMomentDetailsResponse struct {
	ResponsePagination
	MomentDetail []*MomentDetailRes `json:"moment_detail"`
	MomentsNums  int64              `json:"moments_nums"`
}

type CtlMomentCommentRequest struct {
	MomentId   string `json:"moment_id"`
	CommentCtl int32  `json:"comment_ctl"`
}

type CtlMomentCommentResponse struct {
}

type GetCommentsRequest struct {
	RequestPagination
	MomentId        string `form:"moment_id" binding:"omitempty"`
	Privacy         int8   `form:"privacy" binding:"omitempty,numeric"`
	PublishUser     string `form:"publish_user" binding:"omitempty"`
	MediaType       int8   `form:"media_type" binding:"omitempty,numeric"`
	ContentType     int8   `form:"content_type" binding:"omitempty"`
	MContentText    string `form:"m_content_text" binding:"omitempty"`
	ParentCommentId string `form:"parent_comment_id" binding:"omitempty"`
	ReplyCommentId  string `form:"reply_comment_id" binding:"omitempty"`
	CommentType     string `form:"comment_type" binding:"omitempty,numeric,oneof=0 1 2"`
	CommentUser     string `form:"comment_user" binding:"omitempty"`
	CommentContent  string `form:"comment_content" binding:"omitempty"`
	TimeType        int32  `form:"time_type" binding:"omitempty,numeric"`
	StartTime       string `form:"start_time" binding:"omitempty,numeric"`
	EndTime         string `form:"end_time" binding:"omitempty,numeric"`
	OrderBy         string `form:"order_by" binding:"omitempty,oneof=start_time:asc start_time:desc"`
}

type GetCommentsResponse struct {
	ResponsePagination
	Comments     []*MomentComment `json:"comments"`
	CommentsNums int64            `json:"comments_nums"`
}

type GetReplayCommentsRequest struct {
	RequestPagination
	MomentId        string `form:"moment_id" binding:"omitempty"`
	MediaType       int8   `form:"media_type" binding:"omitempty,numeric"`
	ContentType     int8   `form:"content_type" binding:"omitempty"`
	MContentText    string `form:"m_content_text" binding:"omitempty"`
	ParentCommentId string `form:"parent_comment_id" binding:"omitempty"`
	CommentUser     string `form:"comment_user" binding:"omitempty"`
	CommentContent  string `form:"comment_content" binding:"omitempty"`
	TimeType        int32  `form:"time_type" binding:"omitempty,numeric"`
	StartTime       string `form:"start_time" binding:"omitempty,numeric"`
	EndTime         string `form:"end_time" binding:"omitempty,numeric"`
	OrderBy         string `form:"order_by" binding:"omitempty,oneof=start_time:asc start_time:desc"`
	CommentedUser   string `form:"commented_user" binding:"omitempty,numeric"`
}

type GetReplayCommentsResponse struct {
	ResponsePagination
	Comments     []*MomentComment `json:"comments"`
	CommentsNums int64            `json:"comments_nums"`
}

type RemoveCommentsRequest struct {
	CommentIds []string `json:"comment_ids"`
	MomentIds  []string `json:"moment_ids"`
	ReplyIds   []string `json:"reply_ids"`
	ParentIds  []string `json:"parent_ids"`
}

type AlterCommentRequest struct {
	CommentId string `json:"comment_id"`
	Content   string `json:"content"`
}

type SwitchCommentHideStateRequest struct {
	CommentId string `json:"comment_id"`
	Status    int8   `json:"status"`
}

type GetLikesRequest struct {
	RequestPagination
	MomentId     string `form:"moment_id" binding:"omitempty"`
	Privacy      int8   `form:"privacy" binding:"omitempty,numeric"`
	MediaType    int8   `form:"media_type" binding:"omitempty,numeric"`
	PublishUser  string `form:"publish_user" binding:"omitempty"`
	ContentType  int8   `form:"content_type" binding:"omitempty,numeric"`
	MContentText string `form:"m_content_text" binding:"omitempty"`
	LikeUser     string `form:"like_user" binding:"omitempty"`
	TimeType     int32  `form:"time_type" binding:"omitempty,numeric"`
	StartTime    string `form:"start_time" binding:"omitempty,numeric"`
	EndTime      string `form:"end_time" binding:"omitempty,numeric"`
	OrderBy      string `form:"order_by" binding:"omitempty,oneof=start_time:asc start_time:desc"`
}

type GetLikesResponse struct {
	ResponsePagination
	Likes     []*MomentsLike `json:"likes"`
	LikesNums int64          `json:"likes_nums"`
}

type RemoveLikesRequest struct {
	MomentsId []string `json:"moments_id"`
	UsersId   []string `json:"users_id"`
}

type SwitchLikeHideStateRequest struct {
	MomentId string `json:"moment_id"`
	UserId   string `json:"user_id"`
	Status   int8   `json:"status"`
}
