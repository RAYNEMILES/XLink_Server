package base_info

type MomentCreateRequest struct {
	OperationID string `json:"operationID"`
	CreatorID   string `json:"creator_id"`
	// MomentID            int32  `json:"moment_id"`
	MContentText          string                     `json:"m_content_text"`
	MContentImagesArray   string                     `json:"m_content_images_array"`
	MContentVideosArray   string                     `json:"m_content_videos_array"`
	MContentThumbnilArray string                     `json:"m_content_thumbnil_array"`
	ArticleID             int64                      `json:"article_id"`
	MContentImagesArrayV2 []MomentImageRequestObject `json:"m_content_images_array_v2"`
	MContentVideosArrayV2 []MomentVideoRequestObject `json:"m_content_videos_array_v2"`
	WoomFileID            string                     `json:"woom_file_id"`
	// MLikesCount         int32  `json:"m_likes_count"`
	// MCommentsCount      int32  `json:"m_comments_count"`
	// MRepostCount        int64  `json:"m_repost_count"`
	// MCreateTime         string `json:"m_create_time"`
	// MUpdateTime         string `json:"m_update_time"`
	// OrignalCreatorID    string `json:"orignal_creator_id"`
	// OrignalMID          int32  `json:"orignal_m_id"`
	// IsReposted          bool   `json:"is_reposted"`
	// DeleteTime          string `json:"delete_time"`
	// DeletedBy           string `json:"deleted_by"`
	// Status              bool   `json:"status"`
}
type MomentImageRequestObject struct {
	ImageUrl    string `json:"image_url"`
	SnapShotUrl string `json:"snap_shot_url"`
	ImageWidth  int    `json:"image_width"`
	ImageHeight int    `json:"image_height"`
}

type MomentVideoRequestObject struct {
	VideoUrl    string `json:"video_url"`
	SnapShotUrl string `json:"snap_shot_url"`
	VideoWidth  int    `json:"video_width"`
	VideoHeight int    `json:"video_height"`
}

type MomentCreatResp struct {
	CommResp
	Moment *Moment `json:"moment"`
}

type MomentLikeRequest struct {
	OperationID string `json:"operationID"`
	CreatorID   string `json:"creator_id"`
	MomentID    string `json:"moment_id"`
}

type MomentCacelLikeRequest struct {
	OperationID string `json:"operationID"`
	CreatorID   string `json:"creator_id"`
	MomentID    string `json:"moment_id"`
}

type MomentCommentCreateRequest struct {
	OperationID    string `json:"operationID"`
	CreatorID      string `json:"creator_id"`
	MomentID       string `json:"moment_id"`
	CommentContent string `json:"comment_content"`
}

type MomentCommentCreateResp struct {
	CommResp
	MomentComment *MomentCommentResp `json:"momentComment"`
}

// CreateReplyOfMomentComment
type CreateReplyOfMomentCommentRequest struct {
	OperationID    string `json:"operationID"`
	CreatorID      string `json:"creator_id"`
	MomentID       string `json:"moment_id"`
	CommentID      string `json:"comment_id"`
	CommentContent string `json:"comment_content"`
}
type CreateReplyOfMomentCommentResp struct {
	CommResp
	MomentComment *MomentCommentResp `json:"momentCommentReply"`
}

type ListHomeTimeLineOfMomentsRequest struct {
	OperationID   string `json:"operationID"`
	PageNumber    int64  `json:"page_number"`
	MomentLimit   int64  `json:"moment_limit"  binding:"required"`
	CommentsLimit int64  `json:"comments_limit" binding:"required"`
}

type ListHomeTimeLineOfMomentsResp struct {
	HomeTimeLineOfMoments []*HomeTimeLineOfMomentsResp `json:"homeTimeLineOfMoments"`
	PageNumber            int64                        `json:"page_number"`
	MomentLimit           int64                        `json:"moment_limit"`
	CommentsLimit         int64                        `json:"comments_limit"`
}

type HomeTimeLineOfMomentsResp struct {
	Moment         *Moment              `json:"moment"`
	MomentComments []*MomentCommentResp `json:"momentComments"`
	MomentLikes    []*MomentLikeResp    `json:"momentLikes"`
}

type Moment struct {
	CreatorID                 string                         `json:"creatorID"`
	MomentID                  string                         `json:"momentID"`
	MContentText              string                         `json:"mContentText"`
	MContentImagesArray       string                         `json:"mContentImagesArray"`
	MContentVideosArray       string                         `json:"mContentVideosArray"`
	MContentImagesArrayV2     []*MomentImageRequestObject    `json:"mContentImagesArrayV2"`
	MContentVideosArrayV2     []*MomentVideoRequestObject    `json:"mContentVideosArrayV2"`
	MContentThumbnilArray     string                         `json:"mContentThumbnilArray"`
	MLikesCount               int32                          `json:"mLikesCount"`
	MCommentsCount            int32                          `json:"mCommentsCount"`
	MRepostCount              int64                          `json:"mRepostCount"`
	MCreateTime               int64                          `json:"mCreateTime"`
	MUpdateTime               int64                          `json:"mUpdateTime"`
	OrignalCreatorID          string                         `json:"orignalCreatorID"`
	OriginalCreatorName       string                         `json:"originalCreatorName"`
	OriginalCreatorProfileImg string                         `json:"originalCreatorProfileImg"`
	OrignalID                 string                         `json:"orignalID"`
	IsReposted                bool                           `json:"isReposted"`
	DeleteTime                int64                          `json:"deleteTime"`
	DeletedBy                 string                         `json:"deletedBy"`
	Status                    int8                           `json:"status"`
	Privacy                   int32                          `json:"privacy"`
	UserID                    string                         `json:"userID"`
	UserName                  string                         `json:"userName"`
	UserProfileImg            string                         `json:"userProfileImg"`
	ArticleID                 int64                          `json:"articleID"`
	ArticleDetailsInMoment    *GetUserArticleByArticleIDData `json:"articleDetailsInMoment"`
	WoomDetails               *ShortVideoInfo                `json:"WoomDetails"`
}

type MomentCommentResp struct {
	MomentID         string `json:"momentID"`
	CommentID        string `json:"commentID"`
	UserID           string `json:"userID"`
	UserName         string `json:"userName"`
	UserProfileImg   string `json:"userProfileImg"`
	CommentContent   string `json:"commentContent"`
	CommentParentID  string `json:"commentParentID"`
	CPUserID         string `json:"cpUserID"`
	CPUserName       string `json:"cpUserName"`
	CPUserProfileImg string `json:"cpUserProfileImg"`
	CreateBy         string `json:"createBy"`
	CreateTime       int64  `json:"createTime"`
	UpdateBy         string `json:"updateBy"`
	UpdatedTime      int64  `json:"updatedTime"`
	DeletedBy        string `json:"deletedBy"`
	DeleteTime       int64  `json:"deleteTime"`
	Status           int8   `json:"status"`
	AccountStatus    int8   `json:"accountStatus"`
}
type MomentLikeResp struct {
	MomentID       string `json:"momentID"`
	UserID         string `json:"userID"`
	UserName       string `json:"userName"`
	UserProfileImg string `json:"userProfileImg"`
	CreateBy       string `json:"createBy"`
	CreateTime     int64  `json:"createTime"`
	UpdateBy       string `json:"updateBy"`
	UpdatedTime    int64  `json:"updatedTime"`
	DeletedBy      string `json:"deletedBy"`
	DeleteTime     int64  `json:"deleteTime"`
	Status         int8   `json:"status"`
}

type GetMomentDetailsByIDRequest struct {
	OperationID string `json:"operationID"`
	CreatorID   string `json:"creator_id"`
	MomentID    string `json:"moment_id"`
}

type GetMomentDetailsByIDResponse struct {
	Moment         *Moment              `json:"moment"`
	MomentComments []*MomentCommentResp `json:"momentComments"`
	MomentLikes    []*MomentLikeResp    `json:"momentLikes"`
	PageNumber     int64                `json:"page_number"`
	CommentsLimit  int64                `json:"comments_limit"`
}

type GetMomentCommentsByIDRequest struct {
	OperationID   string `json:"operationID"`
	CreatorID     string `json:"creator_id"`
	MomentID      string `json:"moment_id"`
	PageNumber    int64  `json:"page_number"`
	CommentsLimit int64  `json:"comments_limit" binding:"required"`
}

type GetMomentCommentsByIDResponse struct {
	MomentComments []*MomentCommentResp `json:"momentComments"`
	PageNumber     int64                `json:"page_number"`
	CommentsLimit  int64                `json:"comments_limit"`
}

type RepostAMomentRequest struct {
	OperationID string `json:"operationID"`
	CreatorID   string `json:"creator_id"`
	MomentID    string `json:"moment_id"`
}
type RepostAMomentResp struct {
	CommResp
	Moment *Moment `json:"moment"`
}

type DeleteMomentRequest struct {
	OperationID string `json:"operationID"`
	CreatorID   string `json:"creator_id"`
	MomentID    string `json:"moment_id"`
}
type DeleteMomentResp struct {
	CommResp
}

type GetAnyUserMomentsByIDRequest struct {
	PageNumber  int64  `json:"page_number"`
	ShowNumber  int64  `json:"show_number"`
	OperationID string `json:"operationID"`
	UserId      string `json:"user_id"`
}

type GetAnyUserMomentsByIDResp struct {
	CommResp
	PageNumber int       `json:"page_number"`
	ShowNumber int       `json:"show_number"`
	Moments    []*Moment `json:"moments"`
}

type GetUserMomentCountRequest struct {
	OperationID string `json:"operationID"`
	UserId      string `json:"user_id"`
}

type GetUserMomentCountResp struct {
	CommResp
	Posts int64 `json:"posts"`
}

type DeleteMomentCommentRequest struct {
	OperationID string `json:"operationID"`
	CreatorID   string `json:"creatorID"`
	CommentID   string `json:"commentID"`
}

type GlobalSearchInMomentsRequest struct {
	OperationID   string `json:"operationID"`
	SearchKeyWord string `json:"searchKeyWord" binding:"required"`
	PageNumber    int64  `json:"page_number"`
	MomentLimit   int64  `json:"moment_limit"`
	CommentsLimit int64  `json:"comments_limit"`
}
type GlobalSearchInMomentsResp struct {
	HomeTimeLineOfMoments []*HomeTimeLineOfMomentsResp `json:"searchedMoments"`
	PageNumber            int64                        `json:"page_number"`
	MomentLimit           int64                        `json:"moment_limit"`
	CommentsLimit         int64                        `json:"comments_limit"`
}

type GetMomentAnyUserMediaByIDRequest struct {
	OperationID string `json:"operation_id"`
	UserID      string `json:"user_id"`
	LastCount   int32  `json:"last_count"`
}

type GetMomentAnyUserMediaByIDResp struct {
	Pics []struct {
		URL  string `json:"url"`
		Type int8   `json:"type"`
	} `json:"pics"`
	AllMediaMomentCount int64 `json:"allMediaMomentCount"`
}
