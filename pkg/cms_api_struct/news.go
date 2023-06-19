package cms_api_struct

type Official struct {
	Id              int64  `json:"id"`
	UserID          string `json:"user_id"`
	Type            int8   `json:"type"`
	IdType          int8   `json:"id_type"`
	IdName          string `json:"id_name"`
	IdNumber        string `json:"id_number"`
	FaceURL         string `json:"face_url"`
	Nickname        string `json:"nickname"`
	Bio             string `json:"bio"`
	CountryCode     string `json:"country_code"`
	ProcessStatus   int8   `json:"process_status"`
	ProcessBy       string `json:"process_by"`
	ProcessFeedback string `json:"process_feedback"`
	CreateTime      int64  `json:"create_time"`
	ProcessTime     int64  `json:"process_time"`
	InitialNickname string `json:"initial_nickname"`
	IsSystem        int32  `json:"is_system"`
}

type OfficialResponse struct {
	Official
	Interests []InterestType
}

type GetOfficialAccountsRequest struct {
	RequestPagination
	OfficialAccount string `form:"official_account" binding:"omitempty"`
	AccountType     int8   `form:"account_type" binding:"omitempty"`
	IdType          int8   `form:"id_type" binding:"omitempty"`
	IdNumber        string `form:"id_number" binding:"omitempty"`
	ProcessStatus   int8   `form:"process_status" binding:"omitempty"`
	TagsId          string `form:"tags_id" binding:"omitempty"`
	TimeType        int8   `form:"time_type" binding:"omitempty"`
	StartTime       string `form:"start_time" binding:"omitempty,numeric"`
	EndTime         string `form:"end_time" binding:"omitempty,numeric"`
	OrderBy         string `form:"order_by" binding:"omitempty,oneof=start_time:asc start_time:desc"`
	Bio             string `form:"bio" binding:"omitempty"`
	IsSystem        int8   `form:"is_system" binding:"omitempty,numeric"`
}

type GetOfficialAccountsResponse struct {
	Official []*OfficialResponse `json:"official"`
	ResponsePagination
	OfficialNums int64 `json:"official_nums"`
	PendingNums  int64 `json:"pending_nums"`
}

type DeleteOfficialAccountsRequest struct {
	Officials []string `json:"officials"`
}

type AlterOfficialAccountRequest struct {
	Official
	Interests []int64 `json:"interests"`
}

type AlterOfficialAccountResponse struct {
	CommResp
}

type AddOfficialAccountRequest struct {
	UserID          string  `json:"user_id"`
	Nickname        string  `json:"nickname"`
	ProfilePhoto    string  `json:"profile_photo"`
	InitialNickname string  `json:"initial_nickname"`
	Type            int8    `json:"type"`
	IdType          int8    `json:"id_type"`
	IdName          string  `json:"id_name"`
	IdNumber        string  `json:"id_number"`
	Interests       []int64 `json:"interests"`
	IsSystem        int8    `json:"is_system"`
	Bio             string  `json:"bio"`
}

type AddOfficialAccountResponse struct {
	CommResp
}

type ProcessRequest struct {
	OfficialId      int64  `json:"official_id"`
	ProcessStatus   int8   `json:"process_status"`
	ProcessFeedback string `json:"process_feedback"`
}

type ProcessResponse struct {
}

type OfficialBaseInfo struct {
	OfficialName  string `json:"official_name"`
	OfficialType  int8   `json:"official_type"`
	LastLoginIp   string `json:"last_login_ip"`
	LastLoginTime string `json:"last_login_time"`
}

type Article struct {
	OfficialBaseInfo
	OfficialStatus     int64  `json:"official_status"`
	ArticleID          int64  `json:"article_id"`
	OfficialID         int64  `json:"official_id"`
	CoverPhoto         string `json:"cover_photo"`
	Title              string `json:"title"`
	Content            string `json:"content"`
	OfficialProfileImg string `json:"official_profile_img"`
	CreateTime         int64  `json:"create_time"`
	UpdatedBy          string `json:"update_by"`
	UpdateTime         int64  `json:"update_time"`
	DeletedBy          string `json:"deleted_by"`
	DeleteTime         int64  `json:"delete_time"`
	Status             int32  `json:"status"`
	Privacy            int32  `json:"privacy"`
	CommentCounts      int64  `json:"comment_counts"`
	LikeCounts         int64  `json:"like_counts"`
	RepostCounts       int64  `json:"repost_counts"`
}

type ArticleComment struct {
	OfficialBaseInfo
	CommentID         int64  `json:"comment_id"`
	ParentCommentID   int64  `json:"parent_comment_id"`
	ArticleID         int64  `json:"article_id"`
	OfficialID        int64  `json:"official_id"`
	CommentReplyCount int64  `json:"comment_reply_count"`
	Content           string `json:"content"`
	UserID            string `json:"user_id"`
	UserName          string `json:"user_name"`
	UserProfileImg    string `json:"user_profile_img"`
	CreatedBy         string `json:"create_by"`
	CreateTime        int64  `json:"create_time"`
	UpdatedBy         string `json:"update_by"`
	UpdateTime        int64  `json:"update_time"`
	DeletedBy         string `json:"deleted_by"`
	DeleteTime        int64  `json:"delete_time"`
	Status            int32  `json:"status"`
	ArticleTitle      string `json:"article_title"`
	CoverPhoto        string `json:"cover_photo"`
	PostTime          int64  `json:"post_time"`
	CommentLikes      int64  `json:"comment_likes"`
}

type ArticleLike struct {
	OfficialBaseInfo
	ArticleID      int64  `json:"article_id"`
	UserID         string `json:"user_id"`
	UserName       string `json:"user_name"`
	UserProfileImg string `json:"user_profile_img"`
	CreatedBy      string `json:"create_by"`
	CreateTime     int64  `json:"create_time"`
	UpdatedBy      string `json:"update_by"`
	UpdateTime     int64  `json:"update_time"`
	DeletedBy      string `json:"deleted_by"`
	DeleteTime     int64  `json:"delete_time"`
	Status         int32  `json:"status"`
	ArticleTitle   string `json:"article_title"`
	CoverPhoto     string `json:"cover_photo"`
	PostTime       int64  `json:"post_time"`
}

type ArticleRepost struct {
	MomentId      string `json:"moment_id"`
	ArticleId     int64  `json:"article_id"`
	ShareUser     string `json:"share_user"`
	OfficialType  int32  `json:"official_type"`
	ArticleTitle  string `json:"article_title"`
	OriginalUser  string `json:"original_user"`
	CommentCounts int64  `json:"comment_counts"`
	LikeCounts    int64  `json:"like_counts"`
	ShareTime     int64  `json:"share_time"`
	LastLoginIp   string `json:"last_login_ip"`
	Privacy       int32  `json:"privacy"`
	CoverPhoto    string `json:"cover_photo"`
	DeletedBy     string `json:"deleted_by"`
	DeleteTime    int64  `json:"delete_time"`
}

type GetNewsRequest struct {
	RequestPagination
	OfficialAccount string `form:"official_account" binding:"omitempty"`
	AccountType     int8   `form:"account_type" binding:"omitempty,numeric"`
	Ip              string `form:"ip" binding:"omitempty"`
	TimeType        int32  `form:"time_type" binding:"omitempty,numeric"`
	StartTime       string `form:"start_time" binding:"omitempty,numeric"`
	EndTime         string `form:"end_time" binding:"omitempty,numeric"`
	Title           string `form:"title" binding:"omitempty"`
	OrderBy         string `form:"order_by" binding:"omitempty,oneof=start_time:asc start_time:desc"`
}

type GetNewsResponse struct {
	ResponsePagination
	Articles []*Article `json:"articles"`
	NewsNums int64      `json:"news_nums"`
}

type DeleteNewsRequest struct {
	Articles []int64 `json:"articles"`
}

type AlterNewsRequest struct {
	ArticleId   int64  `json:"article_id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	TextContent string `json:"textContent"`
}

type AlterNewsResponse struct {
}

type ChangePrivacyRequest struct {
	ArticleId int64 `json:"article_id"`
	Privacy   int32 `json:"privacy"`
}

type ChangePrivacyResponse struct {
}

type GetNewsCommentsRequest struct {
	RequestPagination
	ArticleId       int64  `form:"article_id" binding:"omitempty,numeric"`
	OfficialAccount string `form:"official_account" binding:"omitempty"`
	AccountType     int8   `form:"account_type" binding:"omitempty,numeric"`
	Ip              string `form:"ip" binding:"omitempty"`
	TimeType        int32  `form:"time_type" binding:"omitempty,numeric"`
	StartTime       string `form:"start_time" binding:"omitempty,numeric"`
	EndTime         string `form:"end_time" binding:"omitempty,numeric"`
	Title           string `form:"title" binding:"omitempty"`
	CommentUser     string `form:"comment_user" binding:"omitempty"`
	CommentKey      string `form:"comment_key" binding:"omitempty"`
	OrderBy         string `form:"order_by" binding:"omitempty,oneof=start_time:asc start_time:desc"`
}

type GetNewsCommentsResponse struct {
	ResponsePagination
	Comments     []*ArticleComment `json:"comments"`
	CommentsNums int64             `json:"comments_nums"`
}

type RemoveNewsCommentsRequest struct {
	Comments []int64 `json:"comments"`
	Parents  []int64 `json:"parents"`
	Articles []int64 `json:"articles"`
}

type AlterNewsCommentRequest struct {
	CommentId string `json:"comment_id"`
	UserId    string `json:"comment_user"`
	Content   string `json:"content"`
}

type AlterNewsCommentResponse struct {
}

type ChangeNewsCommentStatusRequest struct {
	CommentId int64 `json:"comment_id"`
	Status    int32 `json:"status"`
}

type ChangeNewsCommentStatusResponse struct {
}

type GetNewsLikesRequest struct {
	RequestPagination
	ArticleId       int64  `form:"article_id" binding:"omitempty,numeric"`
	OfficialAccount string `form:"official_account" binding:"omitempty"`
	AccountType     int8   `form:"account_type" binding:"omitempty,numeric"`
	Ip              string `form:"ip" binding:"omitempty"`
	Title           string `form:"title" binding:"omitempty"`
	TimeType        int32  `form:"time_type" binding:"omitempty,numeric"`
	StartTime       string `form:"start_time" binding:"omitempty,numeric"`
	EndTime         string `form:"end_time" binding:"omitempty,numeric"`
	LikeUser        string `form:"like_user" binding:"omitempty"`
	OrderBy         string `form:"order_by" binding:"omitempty,oneof=start_time:asc start_time:desc"`
}

type GetNewsLikesResponse struct {
	ResponsePagination
	Likes     []*ArticleLike `json:"likes"`
	LikesNums int64          `json:"likes_nums"`
}

type RemoveNewsLikesRequest struct {
	Articles []int64  `json:"articles"`
	UserIds  []string `json:"user_ids"`
}

type ChangeNewsLikeStatusRequest struct {
	ArticleId int64  `json:"article_id"`
	UserId    string `json:"user_id"`
	Status    int32  `json:"status"`
}

type ChangeNewsLikeStatusResponse struct {
}

type GetRepostArticlesRequest struct {
	RequestPagination
	ArticleId    int64  `form:"article_id" binding:"omitempty,numeric"`
	RepostUser   string `form:"repost_user" binding:"omitempty"`
	AccountType  int32  `form:"account_type" binding:"omitempty"`
	Ip           string `form:"ip" binding:"omitempty"`
	Title        string `form:"title" binding:"omitempty"`
	TimeType     int32  `form:"time_type" binding:"omitempty,numeric"`
	StartTime    string `form:"start_time" binding:"omitempty,numeric"`
	EndTime      string `form:"end_time" binding:"omitempty,numeric"`
	OriginalUser string `form:"original_user" binding:"omitempty"`
	OrderBy      string `form:"order_by" binding:"omitempty,oneof=start_time:asc start_time:desc"`
}

type GetRepostArticlesResponse struct {
	Repost []*ArticleRepost `json:"repost"`
	ResponsePagination
	RepostNums int64 `json:"repost_nums"`
}

type ChangeRepostPrivacyRequest struct {
	MomentId string `json:"article_id"`
	Privacy  int32  `json:"privacy"`
}

type DeleteRepostsRequest struct {
	MomentIds []string `json:"moment_ids"`
	Articles  []int64  `json:"articles"`
}

type GetOfficialFollowersRequest struct {
	RequestPagination
	OperationID     string `form:"operationID"`
	StartTime       string `form:"start_time"`
	EndTime         string `form:"end_time"`
	OfficialAccount string `form:"official_account"`
	User            string `form:"user"`
	Muted           int8   `form:"muted"`
}

type GetOfficialFollowersResponse struct {
	CommResp
	ResponsePagination
	OfficialFollowers []struct {
		OfficialID   int64  `json:"official_id"`
		OfficialName string `json:"official_name"`
		UserID       string `json:"user_id"`
		Username     string `json:"username"`
		FollowTime   int64  `json:"follow_time"`
		BlockTime    int64  `json:"block_time"`
		Muted        bool   `json:"muted"`
		Enabled      bool   `json:"enabled"`
	} `json:"OfficialFollowers"`
	OfficialFollowersCount int64 `json:"OfficialFollowersCount"`
}

type BlockFollowerRequest struct {
	OperationID string `json:"operationID"`
	Block       int8   `json:"block"`
	OfficialID  int64  `json:"official_id"`
	UserID      string `json:"user_id"`
}

type BlockFollowerResponse struct {
	CommResp
}

type MuteFollowerRequest struct {
	OperationID string `json:"operationID"`
	Mute        int8   `json:"mute"`
	OfficialID  int64  `json:"official_id"`
	UserID      string `json:"user_id"`
}

type MuteFollowerResponse struct {
	CommResp
}

type RemoveFollowersRequest struct {
	OperationID string `json:"operationID"`
	Users       []struct {
		OfficialID int64  `json:"official_id"`
		UserID     string `json:"user_id"`
	} `json:"users"`
}

type RemoveFollowersResponse struct {
	CommResp
}
