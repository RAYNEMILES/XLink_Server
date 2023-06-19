package base_info

import "time"

type UserInfo struct {
	UserId    string `json:"userId"`
	Nickname  string `json:"nickName"`
	FaceURL   string `json:"faceUrl"`
	IsDeleted bool   `json:"isDeleted"`
}

type GetUpdateShortVideoSignRequest struct {
	OperationID string  `json:"operationID" binding:"required"`
	Desc        string  `json:"desc" binding:"omitempty,max=60,excludesrune=丨"`
	InterestId  []int32 `json:"interestId" binding:"required,max=10"`
}

type GetUpdateShortVideoSignResponse struct {
	CommResp
	Data struct {
		Sign string `json:"sign"`
	} `json:"data"`
}

type CreateShortVideoSignRequest struct {
	OperationID string  `json:"operationID" binding:"required"`
	Name        string  `json:"name" binding:"required,max=32"`
	FileId      string  `json:"fileId" binding:"required,alphanum,max=32"`
	Desc        string  `json:"desc" binding:"omitempty,max=60"`
	InterestId  []int64 `json:"interestId" binding:"required,min=1,max=10"`
	MediaUrl    string  `json:"mediaUrl" binding:"required,max=256"`
	CoverUrl    string  `json:"coverUrl" binding:"required,max=256"`
}

type CreateShortVideoSignResponse struct {
	CommResp
	Data struct {
		FileId string `json:"fileId"`
	}
}

type GetUserNoticesRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	Type        int32  `json:"type" binding:"omitempty"`
	State       int8   `json:"state" binding:"omitempty"`
	PageNumber  int32  `json:"pageNumber" binding:"required"`
	ShowNumber  int32  `json:"showNumber" binding:"required"`
}

type GetUserNoticesResponse struct {
	CommResp
	Data struct {
		NoticeCount int64 `json:"noticeCount"`
		CurrentPage int32 `json:"current_number" binding:"required"`
		ShowNumber  int32 `json:"show_number" binding:"required"`
		NoticeList  []struct {
			NoticeId   int64  `json:"noticeId"`
			NoticeType int32  `json:"noticeType"`
			FileId     string `json:"fileId"`
			CommentId  int64  `json:"commentId"`
			State      int32  `json:"state"`
			Context    string `json:"context"`
			CreateTime int64  `json:"createTime"`
			UpUserInfo struct {
				UserId    string `json:"userId"`
				Nickname  string `json:"nickName"`
				FaceURL   string `json:"faceUrl"`
				IsDeleted bool   `json:"isDeleted"`
			} `json:"upUserInfo"`
			SelfInfo struct {
				IsLike bool `json:"isLike"`
			} `json:"selfInfo"`
		} `json:"noticeList"`
	} `json:"data"`
}

type GetUserCountRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	UserID      string `json:"userId" binding:"required"`
}

type GetUserCountResponse struct {
	CommResp
	Data struct {
		ShortVideo struct {
			WorkNum              int64 `json:"workNum"`
			LikeNum              int64 `json:"likeNum"`
			HarvestedLikesNumber int64 `json:"harvestedLikesNumber"`
			CommentNum           int64 `json:"commentNum"`
			CommentLikeNum       int64 `json:"commentLikeNum"`
			FansNum              int64 `json:"fansNum"`
			FollowNum            int64 `json:"followNum"`
			NoticeNum            int64 `json:"noticeNum"`
		} `json:"shortVideo"`
		Moment struct {
			WorkNum int64 `json:"workNum"`
			LikeNum int64 `json:"likeNum"`
		} `json:"moment"`
	} `json:"data"`
}

type GetShortVideoByFileIdRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	FileId      string `json:"fileId" binding:"required,alphanum,max=32"`
}

type GetShortVideoByFileIdResponse struct {
	CommResp
	Data struct {
		FileId        string `json:"fileId"`
		Name          string `json:"name"`
		Desc          string `json:"desc"`
		CoverUrl      string `json:"coverUrl"`
		MediaUrl      string `json:"mediaUrl"`
		Type          string `json:"type"`
		Size          int64  `json:"size"`
		Height        int64  `json:"height"`
		Width         int64  `json:"width"`
		LikeNum       int64  `json:"likeNum"`
		CommentNum    int64  `json:"commentNum"`
		ForwardNum    int64  `json:"forwardNum"`
		CreateTime    int64  `json:"createTime"`
		SelfOperation struct {
			IsLike    bool `json:"isLike"`
			IsComment bool `json:"isComment"`
		} `json:"selfInfo"`
		UpUserInfo struct {
			UserId    string `json:"userId"`
			NickName  string `json:"nickName"`
			AvatarUrl string `json:"avatarUrl"`
			IsDeleted bool   `json:"isDeleted"`
		} `json:"upUserInfo"`
		IsUpFollow struct {
			IsFollow bool `json:"isFollow"`
		} `json:"isUpFollow"`
	} `json:"data"`
}

type SearchShortVideoRequest struct {
	OperationID string   `json:"operationID" binding:"required"`
	KeyWord     []string `json:"keyWord" binding:""`
}

type SearchShortVideoResponse struct {
	CommResp
	Data struct {
		ShortVideoList []struct {
			FileId     string `json:"fileId"`
			Name       string `json:"name"`
			Desc       string `json:"desc"`
			CoverUrl   string `json:"coverUrl"`
			MediaUrl   string `json:"mediaUrl"`
			Type       string `json:"type"`
			Size       int64  `json:"size"`
			Height     int64  `json:"height"`
			Width      int64  `json:"width"`
			LikeNum    int64  `json:"likeNum"`
			CommentNum int64  `json:"commentNum"`
			ForwardNum int64  `json:"forwardNum"`
			CreateTime int64  `json:"createTime"`
			SelfInfo   struct {
				IsLike    bool `json:"isLike"`
				IsComment bool `json:"isComment"`
			} `json:"selfInfo"`
			UpUserInfo struct {
				UserId    string `json:"userId"`
				Nickname  string `json:"nickName"`
				FaceURL   string `json:"avatarUrl"`
				IsDeleted bool   `json:"isDeleted"`
			} `json:"upUserInfo"`
			IsUpFollow struct {
				IsFollow bool `json:"isFollow"`
			} `json:"isUpFollow"`
		} `json:"shortVideoList"`
	} `json:"data"`
}

type GetLikeShortVideoRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	UserId      string `json:"userId" binding:"required,alphanum,max=32"`
	PageNumber  int32  `json:"pageNumber" binding:"required"`
	ShowNumber  int32  `json:"showNumber" binding:"required"`
}

type GetShortVideoListByUserIdRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	UserId      string `json:"userId" binding:"required,alphanum,max=32"`
	PageNumber  int32  `json:"pageNumber" binding:"required"`
	ShowNumber  int32  `json:"showNumber" binding:"required"`
}

type GetLikeShortVideoResponse struct {
	CommResp
	Data struct {
		ShortVideoCount int64 `json:"shortVideoCount"`
		CurrentPage     int32 `json:"current_number" binding:"required"`
		ShowNumber      int32 `json:"show_number" binding:"required"`
		IsUpFollow      struct {
			IsFollow bool `json:"isFollow"`
		} `json:"isUpFollow"`
		ShortVideoList []struct {
			FileId     string `json:"fileId"`
			Name       string `json:"name"`
			Desc       string `json:"desc"`
			CoverUrl   string `json:"coverUrl"`
			MediaUrl   string `json:"mediaUrl"`
			Type       string `json:"type"`
			Size       int64  `json:"size"`
			Height     int64  `json:"height"`
			Width      int64  `json:"width"`
			LikeNum    int64  `json:"likeNum"`
			CommentNum int64  `json:"commentNum"`
			ForwardNum int64  `json:"forwardNum"`
			CreateTime int64  `json:"createTime"`
			SelfInfo   struct {
				IsLike    bool `json:"isLike"`
				IsComment bool `json:"isComment"`
			} `json:"selfInfo"`
			IsUpFollow struct {
				IsFollow bool `json:"isFollow"`
			} `json:"isUpFollow"`
			UpUserInfo struct {
				UserId    string `json:"userId"`
				Nickname  string `json:"nickName"`
				FaceURL   string `json:"avatarUrl"`
				IsDeleted bool   `json:"isDeleted"`
			} `json:"upUserInfo"`
		} `json:"shortVideoList"`
	} `json:"data"`
}

type GetRecommendListRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	Size        int32  `json:"size" binding:"required,gt=0"`
}

type GetRecommendListResponse struct {
	CommResp
	Data struct {
		ShortVideoFileIdList []string `json:"shortVideoFileIdList"`
		ShortVideoInfoList   []struct {
			FileId     string `json:"fileId"`
			Name       string `json:"name"`
			Desc       string `json:"desc"`
			CoverUrl   string `json:"coverUrl"`
			MediaUrl   string `json:"mediaUrl"`
			Type       string `json:"type"`
			Size       int64  `json:"size"`
			Height     int64  `json:"height"`
			Width      int64  `json:"width"`
			LikeNum    int64  `json:"likeNum"`
			CommentNum int64  `json:"commentNum"`
			ForwardNum int64  `json:"forwardNum"`
			CreateTime int64  `json:"createTime"`
			SelfInfo   struct {
				IsLike    bool `json:"isLike"`
				IsComment bool `json:"isComment"`
			} `json:"selfInfo"`
			UpUserInfo struct {
				UserId    string `json:"userId"`
				Nickname  string `json:"nickName"`
				FaceURL   string `json:"avatarUrl"`
				IsDeleted bool   `json:"isDeleted"`
			} `json:"upUserInfo"`
			IsUpFollow struct {
				IsFollow bool `json:"isFollow"`
			} `json:"isUpFollow"`
		} `json:"shortVideoInfoList"`
	} `json:"data"`
}

type GetFollowShortVideoListRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	PageNumber  int32  `json:"pageNumber" binding:"required"`
	ShowNumber  int32  `json:"showNumber" binding:"required"`
}

type ShortVideoInfo struct {
	FileId     string `json:"fileId"`
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	CoverUrl   string `json:"coverUrl"`
	MediaUrl   string `json:"mediaUrl"`
	Type       string `json:"type"`
	Size       int64  `json:"size"`
	Height     int64  `json:"height"`
	Width      int64  `json:"width"`
	LikeNum    int64  `json:"likeNum"`
	CommentNum int64  `json:"commentNum"`
	ForwardNum int64  `json:"forwardNum"`
	CreateTime int64  `json:"createTime"`
	SelfInfo   struct {
		IsLike    bool `json:"isLike"`
		IsComment bool `json:"isComment"`
	} `json:"selfInfo"`
	UpUserInfo struct {
		UserId    string `json:"userId"`
		Nickname  string `json:"nickName"`
		FaceURL   string `json:"avatarUrl"`
		IsDeleted bool   `json:"isDeleted"`
	} `json:"upUserInfo"`
	IsUpFollow struct {
		IsFollow bool `json:"isFollow"`
	} `json:"isUpFollow"`
}

type GetFollowShortVideoListResponse struct {
	CommResp
	Data struct {
		ShortVideoCount    int64            `json:"shortVideoCount"`
		CurrentPage        int32            `json:"current_number" binding:"required"`
		ShowNumber         int32            `json:"show_number" binding:"required"`
		ShortVideoList     []string         `json:"shortVideoFileIdList"`
		ShortVideoInfoList []ShortVideoInfo `json:"shortVideoInfoList"`
	} `json:"data"`
}

type GetShortVideoCommentListRequest struct {
	OperationID     string `json:"operationID" binding:"required"`
	FileId          string `json:"fileId" binding:"required,alphanum,max=32"`
	ParentId        int64  `json:"parentId" binding:"omitempty,numeric"`
	PageNumber      int32  `json:"pageNumber" binding:"required"`
	ShowNumber      int32  `json:"showNumber" binding:"required"`
	Order           int32  `json:"order" binding:"required"`
	SourceCommentId int64  `json:"SourceCommentId" binding:"omitempty,numeric"`
}

type GetCommentPageRequest struct {
	OperationID     string `json:"operationID" binding:"required"`
	CommentId       int64  `json:"commentId" binding:"omitempty,numeric"`
	SourceCommentId int64  `json:"SourceCommentId" binding:"omitempty,numeric"`
	PageNumber      int32  `json:"pageNumber" binding:"required"`
	ShowNumber      int32  `json:"showNumber" binding:"required"`
}

type GetCommentPageResponse struct {
	CommResp
	Data struct {
		ReplyCount    int64  `json:"replyCount"`
		TotalReplyNum int64  `json:"totalReplyNum"`
		CurrentPage   int32  `json:"current_number"`
		ShowNumber    int32  `json:"show_number"`
		CreatorUserId string `json:"creatorUserId"`
		CommentInfo   struct {
			CommentId         int64  `json:"commentId"`
			FileId            string `json:"fileId"`
			ParentId          int64  `json:"parentId"`
			Content           string `json:"content"`
			CreateTime        int64  `json:"create_time"`
			CommentReplyCount int64  `json:"comment_reply_count"`
			CommentLikeCount  int64  `json:"comment_like_count"`
			Status            int32  `json:"status"`
			SelfOperation     struct {
				IsLike bool `json:"isLike"`
			} `json:"selfInfo"`
			UpOperation struct {
				IsLike bool `json:"isLike"`
			} `json:"upOperation"`
			CommentUserInfo struct {
				UserId    string `json:"userId"`
				Nickname  string `json:"nickName"`
				FaceURL   string `json:"avatarUrl"`
				IsDeleted bool   `json:"isDeleted"`
			} `json:"commentUserInfo"`
		} `json:"commentInfo"`
		ReplyList []struct {
			CommentId         int64  `json:"commentId"`
			FileId            string `json:"fileId"`
			ParentId          int64  `json:"parentId"`
			Content           string `json:"content"`
			CreateTime        int64  `json:"create_time"`
			CommentReplyCount int64  `json:"comment_reply_count"`
			CommentLikeCount  int64  `json:"comment_like_count"`
			Status            int32  `json:"status"`
			SelfOperation     struct {
				IsLike bool `json:"isLike"`
			} `json:"selfInfo"`
			UpOperation struct {
				IsLike bool `json:"isLike"`
			} `json:"upOperation"`
			CommentUserInfo struct {
				UserId    string `json:"userId"`
				Nickname  string `json:"nickName"`
				FaceURL   string `json:"avatarUrl"`
				IsDeleted bool   `json:"isDeleted"`
			} `json:"commentUserInfo"`
			ReplyComment []struct {
				CommentId         int64  `json:"commentId"`
				FileId            string `json:"fileId"`
				ParentId          int64  `json:"parentId"`
				Content           string `json:"content"`
				CreateTime        int64  `json:"create_time"`
				CommentReplyCount int64  `json:"comment_reply_count"`
				CommentLikeCount  int64  `json:"comment_like_count"`
				Status            int32  `json:"status"`
				SelfOperation     struct {
					IsLike bool `json:"isLike"`
				} `json:"selfInfo"`
				UpOperation struct {
					IsLike bool `json:"isLike"`
				} `json:"upOperation"`
				CommentUserInfo struct {
					UserId    string `json:"userId"`
					Nickname  string `json:"nickName"`
					FaceURL   string `json:"avatarUrl"`
					IsDeleted bool   `json:"isDeleted"`
				} `json:"commentUserInfo"`
				ReplyToUserInfo struct {
					UserId    string `json:"userId"`
					Nickname  string `json:"nickName"`
					FaceURL   string `json:"avatarUrl"`
					IsDeleted bool   `json:"isDeleted"`
				} `json:"replyToUserInfo"`
			} `json:"replyComment"`
		} `json:"replyList"`
	} `json:"data"`
}

type GetCommentPageReplyListRequest struct {
	OperationID     string `json:"operationID" binding:"required"`
	CommentId       int64  `json:"commentId" binding:"required"`
	SourceCommentId int64  `json:"SourceCommentId" binding:"omitempty,numeric"`
	PageNumber      int32  `json:"pageNumber" binding:"required"`
	ShowNumber      int32  `json:"showNumber" binding:"required"`
}

type GetCommentPageReplyListResponse struct {
	CommResp
	Data struct {
		ReplyCount  int64 `json:"replyCount"`
		CurrentPage int32 `json:"current_number"`
		ShowNumber  int32 `json:"show_number"`
		ReplyList   []struct {
			CommentId         int64  `json:"commentId"`
			FileId            string `json:"fileId"`
			ParentId          int64  `json:"parentId"`
			Content           string `json:"content"`
			CreateTime        int64  `json:"create_time"`
			CommentReplyCount int64  `json:"comment_reply_count"`
			CommentLikeCount  int64  `json:"comment_like_count"`
			Status            int32  `json:"status"`
			SelfInfo          struct {
				IsLike bool `json:"isLike"`
			} `json:"selfInfo"`
			CommentUserInfo struct {
				UserId    string `json:"userId"`
				Nickname  string `json:"nickName"`
				FaceURL   string `json:"avatarUrl"`
				IsDeleted bool   `json:"isDelete"`
			} `json:"commentUserInfo"`
			ReplyToUserInfo struct {
				UserId    string `json:"userId"`
				Nickname  string `json:"nickName"`
				FaceURL   string `json:"avatarUrl"`
				IsDeleted bool   `json:"isDelete"`
			} `json:"replyToUserInfo"`
			UpOperation struct{} `json:"upOperation"`
		} `json:"replyList"`
	} `json:"data"`
}

type GetShortVideoCommentListResponse struct {
	CommResp
	Data struct {
		CommentCount       int64 `json:"commentCount"`
		Level0CommentCount int64 `json:"level0CommentCount"`
		CurrentPage        int32 `json:"current_number" binding:"required"`
		ShowNumber         int32 `json:"show_number" binding:"required"`
		CommentList        []struct {
			CommentId         int64  `json:"commentId"`
			FileId            string `json:"fileId"`
			ParentId          int64  `json:"parentId"`
			Content           string `json:"content"`
			CreateTime        int64  `json:"create_time"`
			CommentReplyCount int64  `json:"comment_reply_count"`
			CommentLikeCount  int64  `json:"comment_like_count"`
			Status            int32  `json:"status"`
			SelfOperation     struct {
				IsLike bool `json:"isLike"`
			} `json:"selfInfo"`
			CommentUserInfo struct {
				UserId    string `json:"userId"`
				Nickname  string `json:"nickName"`
				FaceURL   string `json:"avatarUrl"`
				IsDeleted bool   `json:"isDeleted"`
			} `json:"commentUserInfo"`
			UpOperation struct {
				IsLike bool `json:"isLike"`
			} `json:"upOperation"`
			ReplyComment []struct {
				CommentId         int64  `json:"commentId"`
				FileId            string `json:"fileId"`
				Content           string `json:"content"`
				ParentId          int64  `json:"parentId"`
				CreateTime        int64  `json:"create_time"`
				CommentReplyCount int64  `json:"comment_reply_count"`
				CommentLikeCount  int64  `json:"comment_like_count"`
				Status            int32  `json:"status"`
				SelfOperation     struct {
					IsLike bool `json:"isLike"`
				} `json:"selfInfo"`
				CommentUserInfo struct {
					UserId    string `json:"userId"`
					Nickname  string `json:"nickName"`
					FaceURL   string `json:"avatarUrl"`
					IsDeleted bool   `json:"isDeleted"`
				} `json:"commentUserInfo"`
				UpOperation struct {
					IsLike bool `json:"isLike"`
				} `json:"upOperation"`
			} `json:"replyComment"`
		} `json:"commentList"`
	} `json:"data"`
}

type ShortVideoLikeRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	FileId      string `json:"fileId" binding:"required,alphanum,max=32"`
	Like        *bool  `json:"like" binding:"required"`
}

type ShortVideoLikeResponse struct {
	CommResp
}

type ShortVideoFollowRequest struct {
	OperationID  string `json:"operationID" binding:"required"`
	FollowUserId string `json:"followUserId" binding:"required,alphanum,max=32"`
	Follow       *bool  `json:"follow" binding:"required"`
}

type ShortVideoFollowResponse struct {
	CommResp
}

type ShortVideoFollowListRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	PageNumber  int32  `json:"pageNumber" binding:"required"`
	ShowNumber  int32  `json:"showNumber" binding:"required"`
}

type ShortVideoFollowListResponse struct {
	CommResp
	Data struct {
		FollowCount int64      `json:"followCount"`
		CurrentPage int32      `json:"current_number" binding:"required"`
		ShowNumber  int32      `json:"show_number" binding:"required"`
		FollowList  []UserInfo `json:"followList"`
	} `json:"data"`
}

type ShortVideoFansListRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	PageNumber  int32  `json:"pageNumber" binding:"required"`
	ShowNumber  int32  `json:"showNumber" binding:"required"`
}

type ShortVideoFansListResponse struct {
	CommResp
	Data struct {
		FansCount   int64      `json:"followCount"`
		CurrentPage int32      `json:"current_number" binding:"required"`
		ShowNumber  int32      `json:"show_number" binding:"required"`
		FansList    []UserInfo `json:"fansList"`
	} `json:"data"`
}

type ShortVideoCommentLikeRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	FileId      string `json:"fileId" binding:"required,alphanum,max=32"`
	CommentId   int64  `json:"commentId" binding:"required"`
	Like        *bool  `json:"like" binding:"required"`
}

type ShortVideoLikeCommentResponse struct {
	CommResp
}

type ShortVideoCommentRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	FileId      string `json:"fileId" binding:"required,alphanum,max=32"`
	ParentId    int64  `json:"parentId" binding:"omitempty,numeric"`
	Content     string `json:"content" binding:"required,max=220"`
}

type ShortVideoCommentResponse struct {
	CommResp
	Data struct {
		CommentId int64 `json:"commentId"`
	}
}

type DeleteShortVideoCommentRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	FileId      string `json:"fileId" binding:"required,alphanum,max=32"`
	CommentId   int64  `json:"commentId" binding:"required"`
}

type DeleteShortVideoCommentResponse struct {
	CommResp
}

type BlockShortVideoRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	FileId      string `json:"fileId" binding:"required,alphanum,max=32"`
}

type BlockShortVideoResponse struct {
	CommResp
}

type VodCallbackRequest struct {
	EventType                 string                    `json:"EventType" binding:"required"`
	FileDeleteEvent           FileDeleteEvent           `json:"FileDeleteEvent"`
	FileUploadEvent           FileUploadEvent           `json:"FileUploadEvent"`
	ProcedureStateChangeEvent ProcedureStateChangeEvent `json:"ProcedureStateChangeEvent"`
}

type FileDeleteEvent struct {
	FileIDSet            []string `json:"FileIdSet"`
	FileDeleteResultInfo []struct {
		FileID      string `json:"FileId"`
		DeleteParts []struct {
			Type       string `json:"Type"`
			Definition int    `json:"Definition"`
		} `json:"DeleteParts"`
	} `json:"FileDeleteResultInfo"`
}

type FileUploadEvent struct {
	FileId   string `json:"FileId"`
	MetaData struct {
		AudioDuration  float64 `json:"AudioDuration"`
		AudioStreamSet []struct {
			Bitrate      int    `json:"Bitrate"`
			Codec        string `json:"Codec"`
			SamplingRate int    `json:"SamplingRate"`
		} `json:"AudioStreamSet"`
		Bitrate        int     `json:"Bitrate"`
		Container      string  `json:"Container"`
		Duration       float64 `json:"Duration"`
		Height         int     `json:"Height"`
		Rotate         int     `json:"Rotate"`
		Size           int     `json:"Size"`
		VideoDuration  float64 `json:"VideoDuration"`
		VideoStreamSet []struct {
			Bitrate  int    `json:"Bitrate"`
			Codec    string `json:"Codec"`
			CodecTag string `json:"CodecTag"`
			Fps      int    `json:"Fps"`
			Height   int    `json:"Height"`
			Width    int    `json:"Width"`
		} `json:"VideoStreamSet"`
		Width int `json:"Width"`
	} `json:"MetaData"`
	MediaBasicInfo struct {
		Name          string        `json:"Name"`
		Description   string        `json:"Description"`
		CreateTime    string        `json:"CreateTime"`
		UpdateTime    string        `json:"UpdateTime"`
		ExpireTime    string        `json:"ExpireTime"`
		ClassID       int           `json:"ClassId"`
		ClassName     string        `json:"ClassName"`
		ClassPath     string        `json:"ClassPath"`
		CoverUrl      string        `json:"CoverUrl"`
		Type          string        `json:"Type"`
		MediaUrl      string        `json:"MediaUrl"`
		TagSet        []interface{} `json:"TagSet"`
		StorageRegion string        `json:"StorageRegion"`
		SourceInfo    struct {
			SourceType    string `json:"SourceType"`
			SourceContext string `json:"SourceContext"`
		} `json:"SourceInfo"`
		Size int    `json:"Size"`
		Vid  string `json:"Vid"`
	} `json:"MediaBasicInfo"`
	ProcedureTaskID        string `json:"ProcedureTaskId"`
	ReviewAudioVideoTaskID string `json:"ReviewAudioVideoTaskId"`
}

type ProcedureStateChangeEvent struct {
	TaskID   string `json:"TaskId"`
	Status   string `json:"Status"`
	ErrCode  int    `json:"ErrCode"`
	Message  string `json:"Message"`
	FileId   string `json:"FileId"`
	FileName string `json:"FileName"`
	FileURL  string `json:"FileUrl"`
	MetaData struct {
		AudioDuration  float64 `json:"AudioDuration"`
		AudioStreamSet []struct {
			Bitrate      int    `json:"Bitrate"`
			Codec        string `json:"Codec"`
			SamplingRate int    `json:"SamplingRate"`
		} `json:"AudioStreamSet"`
		Bitrate        int     `json:"Bitrate"`
		Container      string  `json:"Container"`
		Duration       float64 `json:"Duration"`
		Height         int     `json:"Height"`
		Md5            string  `json:"Md5"`
		Rotate         int     `json:"Rotate"`
		Size           int     `json:"Size"`
		VideoDuration  float64 `json:"VideoDuration"`
		VideoStreamSet []struct {
			Bitrate          int    `json:"Bitrate"`
			Codec            string `json:"Codec"`
			CodecTag         string `json:"CodecTag"`
			DynamicRangeInfo struct {
				HDRType string `json:"HDRType"`
				Type    string `json:"Type"`
			} `json:"DynamicRangeInfo"`
			Fps    int `json:"Fps"`
			Height int `json:"Height"`
			Width  int `json:"Width"`
		} `json:"VideoStreamSet"`
		Width int `json:"Width"`
	} `json:"MetaData"`
	MediaProcessResultSet []struct {
		Type          string `json:"Type"`
		TranscodeTask struct {
			Status           string    `json:"Status"`
			ErrCode          int       `json:"ErrCode"`
			ErrCodeExt       string    `json:"ErrCodeExt"`
			Message          string    `json:"Message"`
			Progress         int       `json:"Progress"`
			BeginProcessTime time.Time `json:"BeginProcessTime"`
			FinishTime       time.Time `json:"FinishTime"`
			Input            struct {
				Definition     int `json:"Definition"`
				TraceWatermark struct {
					Definition           int    `json:"Definition"`
					DefinitionForBStream int    `json:"DefinitionForBStream"`
					Switch               string `json:"Switch"`
				} `json:"TraceWatermark"`
				WatermarkSet    []interface{} `json:"WatermarkSet"`
				HeadTailSet     []interface{} `json:"HeadTailSet"`
				MosaicSet       []interface{} `json:"MosaicSet"`
				StartTimeOffset int           `json:"StartTimeOffset"`
				EndTimeOffset   int           `json:"EndTimeOffset"`
			} `json:"Input"`
			Output struct {
				URL            string  `json:"Url"`
				Size           int     `json:"Size"`
				Container      string  `json:"Container"`
				Height         int     `json:"Height"`
				Width          int     `json:"Width"`
				Bitrate        int     `json:"Bitrate"`
				Md5            string  `json:"Md5"`
				Duration       float64 `json:"Duration"`
				VideoStreamSet []struct {
					Bitrate          int    `json:"Bitrate"`
					Codec            string `json:"Codec"`
					CodecTag         string `json:"CodecTag"`
					DynamicRangeInfo struct {
						HDRType string `json:"HDRType"`
						Type    string `json:"Type"`
					} `json:"DynamicRangeInfo"`
					Fps    int `json:"Fps"`
					Height int `json:"Height"`
					Width  int `json:"Width"`
				} `json:"VideoStreamSet"`
				AudioStreamSet []struct {
					Bitrate      int    `json:"Bitrate"`
					Codec        string `json:"Codec"`
					SamplingRate int    `json:"SamplingRate"`
				} `json:"AudioStreamSet"`
				Definition             int    `json:"Definition"`
				DigitalWatermarkType   string `json:"DigitalWatermarkType"`
				CopyRightWatermarkText string `json:"CopyRightWatermarkText"`
			} `json:"Output"`
		} `json:"TranscodeTask"`
	} `json:"MediaProcessResultSet"`
	SessionContext  string `json:"SessionContext"`
	SessionID       string `json:"SessionId"`
	TasksPriority   int    `json:"TasksPriority"`
	TasksNotifyMode string `json:"TasksNotifyMode"`
	Operator        string `json:"Operator"`
	OperationType   string `json:"OperationType"`
}
