package cms_api_struct

type GetShortVideoListRequest struct {
	OperationID string `form:"operationID"`
	UserId      string `form:"userId"`
	Status      int32  `form:"status"`
	FileId      string `form:"fileId"`
	Desc        string `form:"desc"`
	EmptyDesc   int64  `form:"emptyDesc"`
	IsBlock     int32  `form:"isBlock"`
	StartTime   int64  `form:"startTime"`
	EndTime     int64  `form:"endTime"`
	RequestPagination
}

type GetShortVideoListResponse struct {
	ResponsePagination
	ShortVideoCount int64 `json:"shortVideoCount"`
	ShortVideoList  []struct {
		Id             int64   `json:"id"`
		UserId         string  `json:"userID"`
		UserName       string  `json:"userName"`
		Status         int32   `json:"status"`
		CreateTime     int64   `json:"creatTime"`
		MediaUrl       string  `json:"mediaUrl"`
		CoverUrl       string  `json:"coverUrl"`
		Desc           string  `json:"desc"`
		LikeNum        int32   `json:"likeNum"`
		CommentNum     int32   `json:"commentNum"`
		ReplyNum       int32   `json:"replyNum"`
		CommentLikeNum int32   `json:"commentLikeNum"`
		Remark         string  `json:"remark"`
		FileId         string  `json:"fileId"`
		InterestId     string  `json:"interestId"`
		InterestArray  []int64 `json:"interestArray"`
	} `json:"shortVideoList"`
}

type DeleteShortVideoRequest struct {
	OperationID string   `json:"operationID"`
	FileId      []string `json:"fileId"`
}

type DeleteShortVideoResponse struct {
	CommResp
}

type AlterShortVideoRequest struct {
	OperationID string `json:"operationID"`
	FileId      string `json:"fileId"`
	Status      int32  `json:"status"`
	Remark      string `json:"remark"`
	Desc        string `json:"desc"`
}

type AlterShortVideoResponse struct {
	CommResp
}

type GetShortVideoLikeListRequest struct {
	OperationID string `form:"operationID"`
	FileId      string `form:"fileId"`
	UserId      string `form:"userId"`
	Status      int32  `form:"status"`
	Desc        string `form:"desc"`
	EmptyDesc   int64  `form:"emptyDesc"`
	StartTime   int64  `form:"startTime"`
	EndTime     int64  `form:"endTime"`
	LikeUserId  string `form:"likeUserId"`
	RequestPagination
}

type GetShortVideoLikeListResponse struct {
	ResponsePagination
	ShortVideoLikeCount int64 `json:"shortVideoLikeCount"`
	ShortVideoLikeList  []struct {
		Id           int64  `json:"id"`
		FileId       string `json:"fileId"`
		MediaUrl     string `json:"mediaUrl"`
		CoverUrl     string `json:"coverUrl"`
		PostUserId   string `json:"postUserId"`
		PostUserName string `json:"postUserName"`
		UserId       string `json:"userId"`
		UserName     string `json:"userName"`
		CreateTime   int64  `json:"creatTime"`
		FileStatus   int32  `json:"fileStatus"`
	} `json:"shortVideoLikeList"`
}

type DeleteShortVideoLikeRequest struct {
	OperationID string  `json:"operationID"`
	LikeIdList  []int64 `json:"likeIdList" binding:"required"`
}

type DeleteShortVideoLikeResponse struct {
	CommResp
}

type GetShortVideoCommentListRequest struct {
	OperationID   string `form:"operationID"`
	FileId        string `form:"fileId"`
	UserId        string `form:"userId"`
	CommentUserId string `form:"commentUserId"`
	Status        int32  `form:"status"`
	Desc          string `form:"desc"`
	EmptyDesc     int64  `form:"emptyDesc"`
	StartTime     int64  `form:"startTime"`
	EndTime       int64  `form:"endTime"`
	Context       string `form:"context"`
	RequestPagination
}

type GetShortVideoCommentListResponse struct {
	ResponsePagination
	ShortVideoCommentCount int64 `json:"shortVideoCommentCount"`
	ShortVideoCommentList  []struct {
		Id           int64  `json:"id"`
		FileId       string `json:"fileId"`
		Status       int32  `json:"fileStatus"`
		PostUserId   string `json:"postUserId"`
		PostUserName string `json:"postUserName"`
		MediaUrl     string `json:"mediaUrl"`
		CoverUrl     string `json:"coverUrl"`
		UserId       string `json:"userId"`
		UserName     string `json:"userName"`
		CreateTime   int64  `json:"creatTime"`
		Content      string `json:"context"`
		LikeNum      int32  `json:"likeNum"`
		ReplyNum     int32  `json:"replyNum"`
		Remark       string `json:"remark"`
		Desc         string `json:"desc"`
	} `json:"shortVideoCommentList"`
}

type DeleteShortVideoCommentRequest struct {
	OperationID   string  `json:"operationID"`
	CommentIdList []int64 `json:"commentIdList" binding:"required"`
}

type DeleteShortVideoCommentResponse struct {
	CommResp
}

type AlterShortVideoCommentRequest struct {
	OperationID string `json:"operationID"`
	CommentId   int64  `json:"commentId"`
	Content     string `json:"content"`
	Remark      string `json:"remark"`
}

type AlterShortVideoCommentResponse struct {
	CommResp
}

type GetShortVideoInterestLabelListRequest struct {
	OperationID  string `form:"operationID"`
	UserId       string `form:"userId"`
	Desc         string `form:"desc"`
	EmptyDesc    int64  `form:"emptyDesc"`
	Default      int64  `form:"default"`
	InterestName string `form:"interestName"`
	RequestPagination
}

type GetShortVideoInterestLabelListResponse struct {
	ResponsePagination
	ShortVideoCount int64 `json:"shortVideoCount"`
	ShortVideoList  []struct {
		Id                  int64    `json:"id"`
		FileId              string   `json:"fileId"`
		MediaUrl            string   `json:"mediaUrl"`
		CoverUrl            string   `json:"coverUrl"`
		UserId              string   `json:"userId"`
		UserName            string   `json:"userName"`
		InterestId          string   `json:"interestId"`
		InterestIdList      []int64  `json:"interestIdList"`
		InterestChineseName []string `json:"interestChineseName"`
		InterestEnglishName []string `json:"interestEnglishName"`
		InterestArabicName  []int64  `json:"interestArabicName"`
	} `json:"shortVideoList"`
}

type AlterShortVideoInterestLabelRequest struct {
	OperationID    string  `json:"operationID"`
	FileId         string  `json:"fileId"`
	InterestIdList []int64 `json:"interestIdList"`
}

type AlterShortVideoInterestLabelResponse struct {
	CommResp
}

type GetShortVideoCommentRepliesRequest struct {
	RequestPagination
	OperationID  string `form:"operationID"`
	CommentId    int64  `form:"comment_id"`
	Privacy      int32  `form:"privacy"`
	Content      string `form:"content"`
	IsEmpty      int8   `form:"is_empty"`
	CommentUser  string `form:"comment_user"`
	Comment      string `form:"comment"`
	ReplyUser    string `form:"reply_user"`
	ReplyContent string `form:"reply_content"`
	StartTime    string `form:"start_time"`
	EndTime      string `form:"end_time"`
	Publisher    string `form:"publisher"`
}

type GetShortVideoCommentRepliesResponse struct {
	ResponsePagination
	CommentReplies []struct {
		FileId              string `json:"file_id"`
		PublishUserID       string `json:"publisher_id"`
		PublishUser         string `json:"publish_user"`
		ShortVideoStatus    int32  `json:"short_video_status"`
		Content             string `json:"content"`
		CoverUrl            string `json:"cover_url"`
		MediaUrl            string `json:"media_url"`
		Size                int64  `json:"size"`
		Height              int64  `json:"height"`
		Width               int64  `json:"width"`
		CommentId           int64  `json:"comment_id"`
		CommentContent      string `json:"comment_content"`
		CommentStatus       int64  `json:"comment_status"`
		ReplyCommentId      int64  `json:"reply_comment_id"`
		ReplyUserName       string `json:"reply_user_name"`
		ReplyUserID         string `json:"reply_user_id"`
		ReplyCommentContent string `json:"reply_comment_content"`
		ReplyTime           int64  `json:"reply_time"`
		LikeCount           int64  `json:"like_count"`
		CommentCount        int64  `json:"comment_count"`
		Remark              string `json:"remark"`
		Status              int64  `json:"status"`
	} `json:"comment_replies"`
	RepliesCount int64 `json:"replies_count"`
	CommResp
}

type AlterReplyRequest struct {
	OperationID    string `json:"operationID"`
	ShortVideoId   string `json:"short_video_id"`
	Content        string `json:"content"`
	ReplyCommentId int64  `json:"reply_comment_id"`
	ReplyContent   string `json:"reply_content"`
	Remark         string `json:"remark"`
}

type AlterReplyResponse struct {
	CommResp
}

type DeleteRepliesRequest struct {
	OperationID string  `json:"operationID"`
	CommentIds  []int64 `json:"comment_ids"`
}

type DeleteRepliesResponse struct {
	CommResp
}

type GetShortVideoCommentLikesRequest struct {
	RequestPagination
	OperationID  string `form:"operationID"`
	CommentID    int64  `form:"comment_id"`
	Privacy      int32  `form:"privacy"`
	Content      string `form:"content"`
	IsEmpty      int32  `form:"is_empty"`
	CommentUser  string `form:"comment_user"`
	Comment      string `form:"comment"`
	ReplyUser    string `form:"reply_user"`
	ReplyContent string `form:"reply_content"`
	LikeUser     string `form:"like_user"`
	StartTime    string `form:"start_time"`
	EndTime      string `form:"end_time"`
	Publisher    string `form:"publisher"`
}

type GetShortVideoCommentLikesResponse struct {
	ResponsePagination
	CommentLikes []struct {
		FileId              string `json:"file_id"`
		ShortVideoStatus    int32  `json:"short_video_status"`
		PublishUserID       string `json:"publisher_id"`
		PublishUser         string `json:"publish_user"`
		Content             string `json:"content"`
		CoverUrl            string `json:"cover_url"`
		MediaUrl            string `json:"media_url"`
		Size                int64  `json:"size"`
		Height              int64  `json:"height"`
		Width               int64  `json:"width"`
		CommentId           int64  `json:"comment_id"`
		CommentContent      string `json:"comment_content"`
		CommentUserName     string `json:"comment_user_name"`
		CommentUserID       string `json:"comment_user_id"`
		ReplyCommentId      int64  `json:"reply_comment_id"`
		ReplyUserName       string `json:"reply_user_name"`
		ReplyUserID         string `json:"reply_user_id"`
		ReplyCommentContent string `json:"reply_comment_content"`
		LikeId              int64  `json:"like_id"`
		LikeUserName        string `json:"like_user_name"`
		LikeUserID          string `json:"like_user_id"`
		LikeTime            int64  `json:"like_time"`
		Remark              string `json:"remark"`
	} `json:"comment_likes"`
	LikesCount int64 `json:"likes_count"`
	CommResp
}

type AlterLikeRequest struct {
	OperationID  string `json:"operationID"`
	ShortVideoID string `json:"short_video_id"`
	LikeId       int64  `json:"like_id"`
	Content      string `json:"content"`
	Remark       string `json:"remark"`
}

type AlterLikeResponse struct {
	CommResp
}

type DeleteLikesRequest struct {
	OperationID string  `json:"operationID"`
	Likes       []int64 `json:"likes"`
}

type DeleteLikesResponse struct {
	CommResp
}

type GetFollowersRequest struct {
	RequestPagination
	OperationID  string `form:"operationID"`
	StartTime    string `form:"start_time"`
	EndTime      string `form:"end_time"`
	Follower     string `form:"follower"`
	FollowedUser string `form:"followed_user"`
}

type GetFollowersResponse struct {
	CommResp
	ResponsePagination
	Followers []struct {
		Id         int64  `json:"id"`
		UserId     string `json:"user_id"`
		FansId     string `json:"fans_id"`
		CreateTime int64  `json:"create_time"`
		UserFace   string `json:"user_face"`
		UserName   string `json:"user_name"`
		FansFace   string `json:"fans_face"`
		FansName   string `json:"fans_name"`
		Remark     string `json:"remark"`
	} `json:"Followers"`
	FollowersCount int64 `json:"FollowersCount"`
}

type AlterFollowerRequest struct {
	OperationID string `json:"operationID"`
	Id          int64  `json:"id"`
	Remark      string `json:"remark"`
}

type AlterFollowerResponse struct {
	CommResp
}

type DeleteFollowersRequest struct {
	OperationID string  `json:"operationID"`
	Id          []int64 `json:"id"`
}

type DeleteFollowersResponse struct {
	CommResp
}
