package base_info

import (
	"Open_IM/pkg/common/config"
	"bytes"
	"encoding/json"
	"github.com/jinzhu/copier"
	"text/template"
	"time"
)

var detailsUrlTemplate = template.Must(template.New("details").Parse(config.Config.News.ArticleUrlTemplate))

type DetailsUrl int64

func (d DetailsUrl) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	if err := detailsUrlTemplate.Execute(&buf, d); err != nil {
		return nil, err
	}
	return json.Marshal(buf.String())
}

type OfficialRegisterRequest struct {
	OperationID string  `json:"operationID" binding:"required"`
	Type        int32   `json:"type" binding:"required,oneof=1 2"`
	IdType      int32   `json:"idType" binding:"required,oneof=1 2 3"`
	IdName      string  `json:"idName" binding:"required,max=255"`
	IdNumber    string  `json:"idNumber" binding:"required,max=255"`
	Nickname    string  `json:"nickname" binding:"required,max=255"`
	FaceURL     string  `json:"faceURL" binding:"max=255"`
	Bio         string  `json:"bio" binding:"required,max=512"`
	CountryCode string  `json:"countryCode" binding:"required,iso3166_1_alpha2"`
	Interests   []int64 `json:"interests" binding:"required"`
}

type GetSelfOfficialInfoRequest struct {
	OperationID string `json:"operationID" binding:"required"`
}

type GetSelfOfficialInfoResponseDataUserInfo struct {
	UserID     string `json:"userID"`
	OfficialID int64  `json:"officialID"`
	Nickname   string `json:"nickname"`
	FaceURL    string `json:"faceURL"`
}

type GetSelfOfficialInfoResponseDataOfficialInfo struct {
	Nickname            string  `json:"nickname"`
	NicknameUpdateTime  int32   `json:"nicknameUpdateTime"`
	NicknameUpdateCount int32   `json:"nicknameUpdateCount"`
	Bio                 string  `json:"bio"`
	FaceURL             string  `json:"faceURL"`
	ProcessStatus       int32   `json:"processStatus"`
	ProcessFeedback     string  `json:"processFeedback"`
	PostCounts          int64   `json:"postCounts"`
	FollowCounts        int64   `json:"followCounts"`
	LikeCounts          int64   `json:"likeCounts"`
	Interests           []int64 `json:"interests"`
}

type GetSelfOfficialInfoResponseData struct {
	UserInfo     GetSelfOfficialInfoResponseDataUserInfo      `json:"userInfo"`
	OfficialInfo *GetSelfOfficialInfoResponseDataOfficialInfo `json:"officialInfo"`
}

type GetSelfOfficialInfoResponse struct {
	CommResp
	Data GetSelfOfficialInfoResponseData `json:"data"`
}

type SetSelfOfficialInfoRequest struct {
	OperationID string  `json:"operationID" binding:"required"`
	FaceURL     string  `json:"faceURL" binding:"max=255"`
	Nickname    string  `json:"nickname" binding:"required,max=30"`
	Bio         string  `json:"bio" binding:"required,max=500"`
	Interests   []int64 `json:"interests" binding:"required,min=1,max=10"`
}

type FollowOfficialRequest struct {
	OperationID string `json:"operationID"`
	OfficialID  int64  `json:"officialID" binding:"required"`
}

type UnfollowOfficialRequest struct {
	OperationID string `json:"operationID"`
	OfficialID  int64  `json:"officialID" binding:"required"`
}

type UpdateOfficialFollowSettingsRequest struct {
	OperationID string `json:"operationID"`
	OfficialID  int64  `json:"officialID" binding:"required"`
	Muted       bool   `json:"muted"`
	Enabled     bool   `json:"enabled"`
}
type OfficialFollowSettingsByOfficialIDRequest struct {
	OperationID string `json:"operationID"`
	OfficialID  int64  `json:"officialID" binding:"required"`
}

type OfficialFollowSettingsByOfficialIDResponse struct {
	CommResp
	UserFollowedOfficialAccSetting *UserFollowedOfficialAccSetting `json:"data"`
}

type BlockOfficialFollowsRequest struct {
	OperationID string   `json:"operationID"`
	UserIDList  []string `json:"userIDList" binding:"required"`
}

type UnblockOfficialFollowsRequest struct {
	OperationID string   `json:"operationID"`
	UserIDList  []string `json:"userIDList" binding:"required"`
}

type DeleteOfficialFollowsRequest struct {
	OperationID string   `json:"operationID"`
	UserIDList  []string `json:"userIDList" binding:"required"`
}

type LikeArticleRequest struct {
	OperationID string `json:"operationID"`
	ArticleID   int64  `json:"articleID" binding:"required"`
}

type UnlikeArticleRequest struct {
	OperationID string `json:"operationID"`
	ArticleID   int64  `json:"articleID" binding:"required"`
}

type DeleteArticleLikeRequest struct {
	OperationID string `json:"operationID"`
	ArticleID   int64  `json:"articleID" binding:"required"`
	UserID      string `json:"userID" binding:"required"`
}

type AddArticleCommentRequest struct {
	OperationID     string `json:"operationID"`
	ArticleID       int64  `json:"articleID" binding:"required"`
	ParentCommentID int64  `json:"parentCommentID"`
	ReplyUserID     string `json:"replyUserID"`
	ReplyOfficialID int64  `json:"ReplyOfficialID"`
	Content         string `json:"content" binding:"required"`
}

type LikeArticleCommentRequest struct {
	OperationID string `json:"operationID"`
	CommentID   int64  `json:"commentID" binding:"required"`
}

type UnlikeArticleCommentRequest struct {
	OperationID string `json:"operationID"`
	CommentID   int64  `json:"commentID" binding:"required"`
}

type OfficialLikeArticleCommentRequest struct {
	OperationID string `json:"operationID"`
	CommentID   int64  `json:"commentID" binding:"required"`
}

type OfficialUnlikeArticleCommentRequest struct {
	OperationID string `json:"operationID"`
	CommentID   int64  `json:"commentID" binding:"required"`
}

type OfficialDeleteArticleCommentRequest struct {
	OperationID string `json:"operationID"`
	CommentID   int64  `json:"commentID" binding:"required"`
}

type OfficialHideArticleCommentRequest struct {
	OperationID string `json:"operationID"`
	CommentID   int64  `json:"commentID" binding:"required"`
}

type OfficialShowArticleCommentRequest struct {
	OperationID string `json:"operationID"`
	CommentID   int64  `json:"commentID" binding:"required"`
}

type AddArticleCommentResponse struct {
	CommResp
	Data int64 `json:"data"`
}

type AddOfficialCommentRequest struct {
	OperationID     string `json:"operationID"`
	ArticleID       int64  `json:"articleID" binding:"required"`
	ParentCommentID int64  `json:"parentCommentID"`
	ReplyUserID     string `json:"replyUserID"`
	ReplyOfficialID int64  `json:"ReplyOfficialID"`
	Content         string `json:"content" binding:"required"`
}

type AddOfficialCommentResponse struct {
	CommResp
	Data int64 `json:"data"`
}

type CreateArticleCommentResp struct {
	CommResp
}

type CreateArticleCommentReplyRequest struct {
	OperationID     string `json:"operationID"`
	CreatorID       string `json:"creator_id"`
	ArticleID       int64  `json:"article_id"`
	Content         string `json:"content"`
	ParentCommentID string `json:"parent_comment_id"`
	OfficialID      int64  `json:"official_id"`
}

type ListArticlesTimeLineRequest struct {
	OperationID string `json:"operationID"`
	Source      int32  `json:"source" binding:"oneof=0 1"`
	OfficialID  int64  `json:"officialID"`
	Offset      int64  `json:"offset"`
	Limit       int64  `json:"limit"  binding:"required,max=100"`
}

type UserArticleSummary struct {
	ArticleID        int64
	Title            string
	CoverPhoto       string
	TextContent      string
	ReadCounts       int64
	UniqueReadCounts int64
	LikeCounts       int64
	CommentCounts    int64
	RepostCounts     int64
	CreateTime       int64
}

type UserArticleSummaryJson struct {
	ArticleID        int64      `json:"articleID"`
	Title            string     `json:"title"`
	CoverPhoto       string     `json:"coverPhoto"`
	TextContent      string     `json:"textContent"`
	DetailsUrl       DetailsUrl `json:"detailsUrl"`
	ReadCounts       int64      `json:"readCounts"`
	UniqueReadCounts int64      `json:"uniqueReadCounts"`
	LikeCounts       int64      `json:"likeCounts"`
	CommentCounts    int64      `json:"commentCounts"`
	RepostCounts     int64      `json:"repostCounts"`
	CreateTime       int64      `json:"createTime"`
}

func (userArticleSummary UserArticleSummary) MarshalJSON() ([]byte, error) {
	var userArticleSummaryJson UserArticleSummaryJson
	if err := copier.Copy(&userArticleSummaryJson, userArticleSummary); err != nil {
		return nil, err
	}
	userArticleSummaryJson.DetailsUrl = DetailsUrl(userArticleSummary.ArticleID)
	return json.Marshal(userArticleSummaryJson)
}

type ListArticlesTimeLineEntry struct {
	Article  UserArticleSummary `json:"article"`
	Official UserFollow         `json:"official"`
}

type ListArticlesTimeLineData struct {
	Entries []ListArticlesTimeLineEntry `json:"entries"`
	Count   int64                       `json:"count"`
}

type ListArticlesTimeLineResponse struct {
	CommResp
	Data *ListArticlesTimeLineData `json:"data"`
}

type Article struct {
	ArticleID          int64  `json:"articleID"`
	OfficialID         int64  `json:"officialID"`
	Title              string `json:"title"`
	Content            string `json:"content"`
	TextContent        string `json:"textContent"`
	OfficialName       string `json:"officialName"`
	OfficialProfileImg string `json:"officialProfilrImg"`
	CreateBy           string `json:"createdBy"`
	CreateTime         int64  `json:"createdTime"`
	UpdateBy           string `json:"updatedBy"`
	UpdatedTime        int64  `json:"updatedTime"`
	DeletedBy          string `json:"deletedBy"`
	DeleteTime         int64  `json:"deleteTime"`
	Status             int32  `json:"status"`
	Privacy            int32  `json:"privacy"`
	CommentCounts      int64  `json:"commentCounts"`
	LikeCounts         int64  `json:"likeCounts"`
	RepostCounts       int64  `json:"repostCounts"`
	ReadCounts         int64  `json:"readCounts"`
	UniqueReadCounts   int64  `json:"uniqueReadCounts"`
}

type ArticleCommentResp struct {
	ArticleID       int64  `json:"articleID"`
	CommentID       string `json:"commentID"`
	UserID          string `json:"userID"`
	UserName        string `json:"userName"`
	UserProfileImg  string `json:"userProfileImg"`
	CommentContent  string `json:"commentContent"`
	CommentParentID string `json:"commentParentID"`
	CreateBy        string `json:"createBy"`
	CreateTime      int64  `json:"createTime"`
	UpdateBy        string `json:"updateBy"`
	UpdatedTime     int64  `json:"updatedTime"`
	DeletedBy       string `json:"deletedBy"`
	DeleteTime      int64  `json:"deleteTime"`
	Status          int32  `json:"status"`
}
type ArticleLikeResp struct {
	ArticleID      int64  `json:"articleID"`
	UserID         string `json:"userID"`
	UserName       string `json:"userName"`
	UserProfileImg string `json:"userProfileImg"`
	CreateBy       string `json:"createBy"`
	CreateTime     int64  `json:"createTime"`
	UpdateBy       string `json:"updateBy"`
	UpdatedTime    int64  `json:"updatedTime"`
	DeletedBy      string `json:"deletedBy"`
	DeleteTime     int64  `json:"deleteTime"`
	Status         int32  `json:"status"`
}

type CreateArticleReq struct {
	OperationID string `json:"operationID" binding:"required"`
	CoverPhoto  string `json:"coverPhoto"`
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content" binding:"required"`
	TextContent string `json:"textContent" binding:"required"`
}

type UpdateArticleReq struct {
	OperationID string `json:"operationID" binding:"required"`
	ArticleID   int64  `json:"articleID" binding:"required"`
	CoverPhoto  string `json:"coverPhoto"`
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content" binding:"required"`
	TextContent string `json:"textContent" binding:"required"`
}

type DeleteArticleReq struct {
	OperationID string `json:"operationID" binding:"required"`
	ArticleID   int64  `json:"articleID" binding:"required"`
}

type ListArticlesReq struct {
	OperationID   string `json:"operationID" binding:"required"`
	Offset        int64  `json:"offset"`
	Limit         int64  `json:"limit" binding:"required,max=100"`
	MinCreateTime int64  `json:"minCreateTime"`
}

type ArticleSummaryEntry struct {
	ArticleID        int64  `json:"articleID"`
	Title            string `json:"title"`
	CoverPhoto       string `json:"coverPhoto"`
	TextContent      string `json:"textContent"`
	ReadCounts       int64  `json:"readCounts"`
	UniqueReadCounts int64  `json:"uniqueReadCounts"`
	LikeCounts       int64  `json:"likeCounts"`
	CommentCounts    int64  `json:"commentCounts"`
	RepostCounts     int64  `json:"repostCounts"`
	CreateTime       int64  `json:"createTime"`
}

type ListArticlesData struct {
	Entries []ArticleSummaryEntry `json:"entries"`
	Count   int64                 `json:"count"`
}

type ListArticlesResp struct {
	CommResp
	Data ListArticlesData `json:"data"`
}

type ListOfficialSelfFollowsRequest struct {
	OperationID   string `json:"operationID" binding:"required"`
	Offset        int32  `json:"offset"`
	Limit         int32  `json:"limit" binding:"required,max=100"`
	OrderBy       int32  `json:"orderBy" binding:"oneof=0 1"`
	MinFollowTime int64  `json:"minFollowTime"`
	MinBlockTime  int64  `json:"minBlockTime"`
	BlockFilter   *bool  `json:"blockFilter"`
}

type OfficialFollowEntry struct {
	UserID     string `json:"userID"`
	FaceURL    string `json:"faceURL"`
	Nickname   string `json:"nickname"`
	Gender     int32  `json:"gender"`
	FollowTime int64  `json:"followTime"`
	BlockTime  int64  `json:"blockTime"`
}

type ListOfficialSelfFollowsData struct {
	Entries []OfficialFollowEntry `json:"entries"`
	Count   int64                 `json:"count"`
}

type ListOfficialSelfFollowsResponse struct {
	CommResp
	Data ListOfficialSelfFollowsData `json:"data"`
}

type GetOfficialArticleReq struct {
	OperationID string `json:"operationID"`
	ArticleID   int64  `json:"articleID" binding:"required"`
}

type ArticleEntry struct {
	ArticleID          int64  `json:"articleID"`
	OfficialID         int64  `json:"officialID"`
	Title              string `json:"title"`
	CoverPhoto         string `json:"coverPhoto"`
	Content            string `json:"content"`
	TextContent        string `json:"textContent"`
	OfficialName       string `json:"officialName"`
	OfficialProfileImg string `json:"officialProfilrImg"`
	CreateTime         int64  `json:"createdTime"`
	UpdatedTime        int64  `json:"updatedTime"`
	CommentCounts      int64  `json:"commentCounts"`
	LikeCounts         int64  `json:"likeCounts"`
	RepostCounts       int64  `json:"repostCounts"`
	ReadCounts         int64  `json:"readCounts"`
	UniqueReadCounts   int64  `json:"uniqueReadCounts"`
}

type GetOfficialArticleResp struct {
	CommResp
	Data *ArticleEntry `json:"data"`
}

type ListArticleLikesRequest struct {
	OperationID   string `json:"operationID" binding:"required"`
	ArticleID     int64  `json:"articleID" binding:"required"`
	Keyword       string `json:"keyword"`
	Offset        int32  `json:"offset"`
	Limit         int32  `json:"limit" binding:"required,max=100"`
	MinCreateTime int64  `json:"minCreateTime"`
}

type ArticleLikeEntry struct {
	UserID     string `json:"userID"`
	FaceURL    string `json:"faceURL"`
	Nickname   string `json:"nickname"`
	Gender     int32  `json:"gender"`
	CreateTime int64  `json:"createTime"`
}

type ListArticleLikesData struct {
	Entries []ArticleLikeEntry `json:"entries"`
	Count   int64              `json:"count"`
}

type ListArticleLikesResponse struct {
	CommResp
	Data ListArticleLikesData `json:"data"`
}

type ListArticleCommentsRequest struct {
	OperationID string `json:"operationID"`
	ArticleID   int64  `json:"articleID" binding:"required"`
	Offset      int32  `json:"offset"`
	Limit       int32  `json:"limit" binding:"required,max=100"`
	ReplyLimit  int32  `json:"ReplyLimit" binding:"required,max=100"`
}

type CommentEntry struct {
	CommentID     int64  `json:"commentID"`
	UserID        string `json:"userID"`
	OfficialID    int64  `json:"officialID"`
	Nickname      string `json:"nickname"`
	FaceURL       string `json:"faceURL"`
	ReplyNickname string `json:"replyNickname"`
	ReplyFaceURL  string `json:"replyFaceURL"`
	ReplyCounts   int64  `json:"replyCounts"`
	LikeCounts    int64  `json:"likeCounts"`
	Content       string `json:"content"`
	CreateTime    int64  `json:"createTime"`
	LikeTime      int64  `json:"likeTime"`
	Status        int32  `json:"status"`
}

type ListArticleCommentRepliesRequest struct {
	OperationID     string `json:"operationID"`
	ParentCommentID int64  `json:"parentCommentID" binding:"required"`
	Offset          int32  `json:"offset"`
	Limit           int32  `json:"limit" binding:"required,max=100"`
}

type ListArticleCommentRepliesData struct {
	Count   int64          `json:"count"`
	Entries []CommentEntry `json:"entries"`
}

type CommentEntryWithReplies struct {
	CommentEntry
	Replies ListArticleCommentRepliesData `json:"replies"`
}

type ListArticleCommentsData struct {
	Count   int64                     `json:"count"`
	Entries []CommentEntryWithReplies `json:"entries"`
}

type ListArticleCommentsResponse struct {
	CommResp
	Data ListArticleCommentsData `json:"data"`
}

type ListArticleCommentRepliesResponse struct {
	CommResp
	Data ListArticleCommentRepliesData `json:"data"`
}

type UserArticleCommentEntry struct {
	CommentID       int64  `json:"commentID"`
	ParentCommentID int64  `json:"parent_comment_id"`
	UserID          string `json:"userID"`
	OfficialID      int64  `json:"officialID"`
	Nickname        string `json:"nickname"`
	FaceURL         string `json:"faceURL"`
	ReplyNickname   string `json:"replyNickname"`
	ReplyFaceURL    string `json:"replyFaceURL"`
	ReplyCounts     int64  `json:"replyCounts"`
	LikeCounts      int64  `json:"likeCounts"`
	Content         string `json:"content"`
	CreateTime      int64  `json:"createTime"`
	LikeTime        int64  `json:"likeTime"`
}

type UserArticleCommentEntryWithTopReplies struct {
	UserArticleCommentEntry
	TopReplies []UserArticleCommentEntry `json:"topReplies"`
}

type ListUserArticleCommentsRequest struct {
	OperationID string `json:"operationID"`
	ArticleID   int64  `json:"articleID" binding:"required"`
	Offset      int64  `json:"offset"`
	Limit       int64  `json:"limit" binding:"required,max=100"`
}

type ListUserArticleCommentsData struct {
	Count   int64                                   `json:"count"`
	Entries []UserArticleCommentEntryWithTopReplies `json:"entries"`
}

type ListUserArticleCommentsResponse struct {
	CommResp
	Data ListUserArticleCommentsData `json:"data"`
}

type ListUserArticleCommentRepliesRequest struct {
	OperationID string `json:"operationID"`
	CommentID   int64  `json:"commentID" binding:"required"`
	Offset      int64  `json:"offset"`
	Limit       int64  `json:"limit" binding:"required,max=100"`
}

type ListUserArticleCommentRepliesData struct {
	Count   int64                     `json:"count"`
	Entries []UserArticleCommentEntry `json:"entries"`
}

type ListUserArticleCommentRepliesResponse struct {
	CommResp
	Data ListUserArticleCommentRepliesData `json:"data"`
}

type UserFollowListRequest struct {
	OperationID string `json:"operationID"`
	Offset      int32  `json:"offset"`
	Limit       int32  `json:"limit" binding:"required,max=100"`
	Keyword     string `json:"keyword"`
}

type UserFollow struct {
	OfficialID int64  `json:"officialID"`
	Nickname   string `json:"nickname"`
	FaceURL    string `json:"faceURL"`
	Bio        string `json:"bio"`
	Type       int32  `json:"type"`
	FollowTime int64  `json:"followTime"`
	Muted      bool   `json:"muted"`
	Enabled    bool   `json:"enabled"`
}

type UserFollowedOfficialAccSetting struct {
	OfficialID int64 `json:"officialID"`
	FollowTime int64 `json:"followTime"`
	Muted      bool  `json:"muted"`
	Enabled    bool  `json:"enabled"`
}

type UserFollowListData struct {
	Count   int64        `json:"count"`
	Entries []UserFollow `json:"entries"`
}

type UserFollowListResponse struct {
	CommResp
	Data UserFollowListData `json:"data"`
}

type GetOfficialProfileRequest struct {
	OperationID string `json:"operationID"`
	OfficialID  int64  `json:"officialID"`
}

type GetOfficialProfileResponse struct {
	CommResp
	Data *UserFollow `json:"data"`
}

type GetRecentAnalyticsRequest struct {
	OperationID string `json:"operationID"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
}

type GetRecentAnalyticsEntryGender struct {
	Unknown int64 `json:"unknown"`
	Male    int64 `json:"male"`
	Female  int64 `json:"female"`
}

type GetRecentAnalyticsEntry struct {
	LikesByGender       GetRecentAnalyticsEntryGender `json:"likesByGender"`
	CommentsByGender    GetRecentAnalyticsEntryGender `json:"commentsByGender"`
	FollowsByGender     GetRecentAnalyticsEntryGender `json:"followsByGender"`
	ReadsByGender       GetRecentAnalyticsEntryGender `json:"readsByGender"`
	UniqueReadsByGender GetRecentAnalyticsEntryGender `json:"uniqueReadsByGender"`
}

type GetRecentAnalyticsData struct {
	Current  GetRecentAnalyticsEntry `json:"current"`
	Previous GetRecentAnalyticsEntry `json:"previous"`
}

type GetRecentAnalyticsResponse struct {
	CommResp
	Data GetRecentAnalyticsData `json:"data"`
}

type GetAnalyticsByDayRequest struct {
	OperationID string    `json:"operationID"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
}

type GetAnalyticsByDayEntry struct {
	Date        string `json:"date"`
	Likes       int64  `json:"likes"`
	Comments    int64  `json:"comments"`
	Follows     int64  `json:"follows"`
	Reads       int64  `json:"reads"`
	UniqueReads int64  `json:"uniqueReads"`
}

type GetAnalyticsByDayResponse struct {
	CommResp
	Data []GetAnalyticsByDayEntry `json:"data"`
}

type SearchOfficialAccountsRequest struct {
	OperationID string `json:"operationID"`
	Keyword     string `json:"keyword"`
	Offset      int32  `json:"offset"`
	Limit       int32  `json:"limit" binding:"required,max=100"`
}

type SearchOfficialAccountsData struct {
	Entries []UserFollow `json:"entries"`
	Count   int64        `json:"count"`
}

type SearchOfficialAccountsResponse struct {
	CommResp
	Data SearchOfficialAccountsData `json:"data"`
}

type SearchArticlesRequest struct {
	OperationID   string `json:"operationID"`
	Keyword       string `json:"keyword"`
	OfficialID    int64  `json:"officialID"`
	MinReadTime   int64  `json:"minReadTime"`
	MaxReadTime   int64  `json:"maxReadTime"`
	MinCreateTime int64  `json:"minCreateTime"`
	MaxCreateTime int64  `json:"maxCreateTime"`
	Sort          int64  `json:"sort" binding:"oneof=0 1 2 3"`
	Offset        int64  `json:"offset"`
	Limit         int64  `json:"limit" binding:"required,max=100"`
}

type SearchArticlesEntry struct {
	Article  UserArticleSummary `json:"article"`
	Official UserFollow         `json:"official"`
	ReadTime int64              `json:"readTime"`
}

type SearchArticlesData struct {
	Entries []SearchArticlesEntry `json:"entries"`
	Count   int64                 `json:"count"`
}

type SearchArticlesResponse struct {
	CommResp
	Data *SearchArticlesData `json:"data"`
}

type GetUserArticleByArticleIDRequest struct {
	OperationID string `json:"operationID"`
	ArticleID   int64  `json:"articleID"`
}

type UserArticle struct {
	ArticleID        int64
	Title            string
	CoverPhoto       string
	TextContent      string
	Content          string
	ReadCounts       int64
	UniqueReadCounts int64
	CommentCounts    int64
	RepostCounts     int64
	CreateTime       int64
	LikeTime         int64
	FavoriteTime     int64
	FavoriteID       string
}

type UserArticleJson struct {
	ArticleID        int64      `json:"articleID"`
	Title            string     `json:"title"`
	CoverPhoto       string     `json:"coverPhoto"`
	TextContent      string     `json:"textContent"`
	Content          string     `json:"content"`
	DetailsUrl       DetailsUrl `json:"detailsUrl"`
	ReadCounts       int64      `json:"readCounts"`
	UniqueReadCounts int64      `json:"uniqueReadCounts"`
	CommentCounts    int64      `json:"commentCounts"`
	RepostCounts     int64      `json:"repostCounts"`
	CreateTime       int64      `json:"createTime"`
	LikeTime         int64      `json:"likeTime"`
	FavoriteTime     int64      `json:"favoriteTime"`
	FavoriteID       string     `json:"favoriteID"`
}

func (userArticle UserArticle) MarshalJSON() ([]byte, error) {
	var userArticleJson UserArticleJson
	if err := copier.Copy(&userArticleJson, userArticle); err != nil {
		return nil, err
	}
	userArticleJson.DetailsUrl = DetailsUrl(userArticleJson.ArticleID)
	return json.Marshal(userArticleJson)
}

type GetUserArticleByArticleIDData struct {
	Article  UserArticle `json:"article"`
	Official UserFollow  `json:"official"`
}

type GetUserArticleByArticleIDResponse struct {
	CommResp
	Data *GetUserArticleByArticleIDData `json:"data"`
}

type InsertArticleReadRequest struct {
	OperationID string `json:"operationID"`
	ArticleID   int64  `json:"articleID"`
}

type ListUserArticleReadsRequest struct {
	OperationID string `json:"operationID"`
	Offset      int64  `json:"offset"`
	Limit       int64  `json:"limit" binding:"required,max=100"`
}

type ListUserArticleReadsEntry struct {
	Article  UserArticleSummary `json:"article"`
	Official UserFollow         `json:"official"`
	ReadTime int64              `json:"readTime"`
}

type ListUserArticleReadsData struct {
	Entries []ListUserArticleReadsEntry `json:"entries"`
	Count   int64                       `json:"count"`
}

type ListUserArticleReadsResponse struct {
	CommResp
	Data ListUserArticleReadsData `json:"data"`
}

type ClearUserArticleReadsRequest struct {
	OperationID string `json:"operationID"`
}

type DeleteArticleCommentRequest struct {
	OperationID string `json:"operationID"`
	CommentID   int64  `json:"commentID" binding:"required"`
}
type DeleteArticleCommentResponse struct {
	CommResp
	Data interface{} `json:"data"`
}
type FollowedOfficialConversationRequest struct {
	OperationID string `json:"operationID"`
}

type FollowedOfficialConversationResponse struct {
	CommResp
	Data []interface{} `json:"data"`
}

type GetOfficialIDNumberAvailabilityRequest struct {
	OperationID string `json:"operationID"`
	IDNumber    string `json:"id_number" binding:"required,max=255"`
	IDType      int32  `json:"id_type" binding:"required"`
}

type GetOfficialIDNumberAvailabilityResponse struct {
	CommResp
	Data GetOfficialIDNumberAvailability `json:"data"`
}

type GetOfficialIDNumberAvailability struct {
	IDNumber    string `json:"idNumber"`
	IsAvailable bool   `json:"isAvailable"`
}
