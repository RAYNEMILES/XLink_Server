package main

import (
	"Open_IM/internal/api/appversion"
	apiAuth "Open_IM/internal/api/auth"
	"Open_IM/internal/api/callback"
	"Open_IM/internal/api/channel_code"
	apiChat "Open_IM/internal/api/chat"
	"Open_IM/internal/api/conversation"
	"Open_IM/internal/api/discover"
	"Open_IM/internal/api/familiar"
	"Open_IM/internal/api/favorite"
	"Open_IM/internal/api/friend"
	gameStore "Open_IM/internal/api/game_store"
	"Open_IM/internal/api/group"
	"Open_IM/internal/api/invite_code"
	"Open_IM/internal/api/manage"
	"Open_IM/internal/api/me_page_url"
	"Open_IM/internal/api/middleware"
	"Open_IM/internal/api/privacy"
	"Open_IM/internal/api/qr_login"
	"Open_IM/internal/api/short_video"

	"Open_IM/internal/api/interest"
	"Open_IM/internal/api/moments"
	"Open_IM/internal/api/news"

	"Open_IM/internal/api/oauth"
	apiThird "Open_IM/internal/api/third"
	"Open_IM/internal/api/user"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	// "syscall"
	userMiddleWere "Open_IM/internal/api/middleware"
	"Open_IM/pkg/common/constant"
)

func main() {
	//runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪
	//runtime.SetBlockProfileRate(1)     // 开启对阻塞操作的跟踪
	//
	//go func() {
	//	// 启动一个自定义mux的http服务器
	//	mux := http.NewServeMux()
	//	mux.HandleFunc("/debug/pprof/", pprof.Index)
	//	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	//	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	//	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	//	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	//
	//	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
	//		w.Write([]byte("hello"))
	//	})
	//	// 启动一个 http server，注意 pprof 相关的 handler 已经自动注册过了
	//	if err := http.ListenAndServe(":6061", mux); err != nil {
	//		log1.Fatal(err)
	//		log.NewError("", utils.GetSelfFuncName(), "启动pprof报错：", err.Error())
	//	}
	//	os.Exit(0)
	//}()

	log.NewPrivateLog(constant.OpenImApiLog)
	gin.SetMode(gin.ReleaseMode)
	f, _ := os.Create("../logs/api.log")
	gin.DefaultWriter = io.MultiWriter(f)
	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	r.Use(utils.CorsHandler())

	log.Info("load  config: ", config.Config)
	// user routing group, which handles user registration and login services
	userRouterGroup := r.Group("/user")
	{
		userRouterGroup.POST("/update_user_info", middleware.JWTAuth(), user.UpdateUserInfo)       // 1
		userRouterGroup.POST("/remove_user_faceurl", middleware.JWTAuth(), user.RemoveUserFaceUrl) // 1
		userRouterGroup.POST("/set_global_msg_recv_opt", middleware.JWTAuth(), user.SetGlobalRecvMessageOpt)
		userRouterGroup.POST("/get_users_info", middleware.JWTAuth(), user.GetUsersInfo)                  // 1
		userRouterGroup.POST("/get_self_user_info", middleware.JWTAuth(), user.GetSelfUserInfo)           // 1
		userRouterGroup.POST("/get_users_online_status", middleware.JWTAuth(), user.GetUsersOnlineStatus) // 1
		userRouterGroup.POST("/get_users_info_from_cache", middleware.JWTAuth(), user.GetUsersInfoFromCache)
		userRouterGroup.POST("/get_user_friend_from_cache", middleware.JWTAuth(), user.GetFriendIDListFromCache)
		userRouterGroup.POST("/get_black_list_from_cache", middleware.JWTAuth(), user.GetBlackIDListFromCache)
		userRouterGroup.POST("/delete_self_user", middleware.JWTAuth(), user.DeleteUser)
		userRouterGroup.POST("/starting_welcome_messages", channel_code.StartingWelcomeMessages)
		userRouterGroup.POST("/account_check")

		userRouterGroup.POST("/search_user", middleware.JWTAuth(), user.SearchUser) // 1

	}
	// friend routing group
	friendRouterGroup := r.Group("/friend")
	{
		//	friendRouterGroup.POST("/get_friends_info", friend.GetFriendsInfo)
		friendRouterGroup.POST("/add_friend", middleware.JWTAuth(), friend.AddFriend)                              //1
		friendRouterGroup.POST("/delete_friend", middleware.JWTAuth(), friend.DeleteFriend)                        //1
		friendRouterGroup.POST("/get_friend_apply_list", middleware.JWTAuth(), friend.GetFriendApplyList)          //1
		friendRouterGroup.POST("/get_self_friend_apply_list", middleware.JWTAuth(), friend.GetSelfFriendApplyList) //1
		friendRouterGroup.POST("/get_friend_list", middleware.JWTAuth(), friend.GetFriendList)                     //1
		friendRouterGroup.POST("/get_friends", middleware.JWTAuth(), friend.GetFriendsInfo)
		friendRouterGroup.POST("/add_friend_response", middleware.JWTAuth(), friend.AddFriendResponse) //1
		friendRouterGroup.POST("/set_friend_remark", middleware.JWTAuth(), friend.SetFriendRemark)     //1
		friendRouterGroup.POST("/get_friend_remark_nick", middleware.JWTAuth(), friend.GetFriendRemarkOrNick)

		friendRouterGroup.POST("/add_black", middleware.JWTAuth(), friend.AddBlack)          // 1
		friendRouterGroup.POST("/get_black_list", middleware.JWTAuth(), friend.GetBlacklist) // 1
		friendRouterGroup.POST("/remove_black", middleware.JWTAuth(), friend.RemoveBlack)    // 1

		friendRouterGroup.POST("/import_friend", middleware.JWTAuth(), friend.ImportFriend) // 1
		friendRouterGroup.POST("/is_friend", middleware.JWTAuth(), friend.IsFriend)         // 1

		friendRouterGroup.POST("/add_black_friends", userMiddleWere.JWTAuth(), friend.AddBlackFriends)
		friendRouterGroup.POST("/get_black_friends", userMiddleWere.JWTAuth(), friend.GetBlackFriends)
		friendRouterGroup.POST("/remove_black_friends", userMiddleWere.JWTAuth(), friend.RemoveBlackFriends)

	}
	// group related routing group
	groupRouterGroup := r.Group("/group")
	{
		groupRouterGroup.POST("/create_group", middleware.JWTAuth(), group.CreateGroup)                                   // 1
		groupRouterGroup.POST("/set_group_info", middleware.JWTAuth(), group.SetGroupInfo)                                // 1
		groupRouterGroup.POST("/join_group", middleware.JWTAuth(), group.JoinGroup)                                       // 1
		groupRouterGroup.POST("/quit_group", middleware.JWTAuth(), group.QuitGroup)                                       // 1
		groupRouterGroup.POST("/group_application_response", middleware.JWTAuth(), group.ApplicationGroupResponse)        // 1
		groupRouterGroup.POST("/transfer_group", middleware.JWTAuth(), group.TransferGroupOwner)                          // 1
		groupRouterGroup.POST("/get_recv_group_applicationList", middleware.JWTAuth(), group.GetRecvGroupApplicationList) // 1
		groupRouterGroup.POST("/get_user_req_group_applicationList", middleware.JWTAuth(), group.GetUserReqGroupApplicationList)
		groupRouterGroup.POST("/get_groups_info", middleware.JWTAuth(), group.GetGroupsInfo)            // 1
		groupRouterGroup.POST("/kick_group", middleware.JWTAuth(), group.KickGroupMember)               // 1
		groupRouterGroup.POST("/get_group_member_list", middleware.JWTAuth(), group.GetGroupMemberList) // no use
		groupRouterGroup.POST("/get_group_member_list_v2", middleware.JWTAuth(), group.GetGroupMemberListV2)
		groupRouterGroup.POST("/get_group_all_member_list", middleware.JWTAuth(), group.GetGroupAllMemberList) // 1
		groupRouterGroup.POST("/get_group_members_info", middleware.JWTAuth(), group.GetGroupMembersInfo)      // 1
		groupRouterGroup.POST("/invite_user_to_group", middleware.JWTAuth(), group.InviteUserToGroup)          // 1
		groupRouterGroup.POST("/get_joined_group_list", middleware.JWTAuth(), group.GetJoinedGroupList)
		groupRouterGroup.POST("/dismiss_group", middleware.JWTAuth(), group.DismissGroup) //
		groupRouterGroup.POST("/mute_group_member", middleware.JWTAuth(), group.MuteGroupMember)
		groupRouterGroup.POST("/cancel_mute_group_member", middleware.JWTAuth(), group.CancelMuteGroupMember) // MuteGroup
		groupRouterGroup.POST("/mute_group", middleware.JWTAuth(), group.MuteGroup)
		groupRouterGroup.POST("/cancel_mute_group", middleware.JWTAuth(), group.CancelMuteGroup)
		groupRouterGroup.POST("/set_group_member_nickname", middleware.JWTAuth(), group.SetGroupMemberNickname)
		groupRouterGroup.POST("/set_group_member_info", middleware.JWTAuth(), group.SetGroupMemberInfo)
		groupRouterGroup.POST("/check_group_update_version", middleware.JWTAuth(), group.CheckGroupUpdateVersionsFromLocal)
	}
	superGroupRouterGroup := r.Group("/super_group")
	{
		superGroupRouterGroup.POST("/get_joined_group_list", middleware.JWTAuth(), group.GetJoinedSuperGroupList)
		superGroupRouterGroup.POST("/get_groups_info", middleware.JWTAuth(), group.GetSuperGroupsInfo)
	}
	// certificate
	authRouterGroup := r.Group("/auth")
	{
		authRouterGroup.POST("/user_register", apiAuth.UserRegister)                     // 1
		authRouterGroup.POST("/user_token", apiAuth.UserToken)                           // 1
		authRouterGroup.POST("/parse_token", apiAuth.ParseToken)                         // 1
		authRouterGroup.POST("/force_logout", middleware.JWTAuth(), apiAuth.ForceLogout) // 1
		authRouterGroup.POST("/updateIpUserStatus", middleware.JWTAuth(), apiAuth.UpdateUserIpLocation)
		authRouterGroup.POST("/getIpUserStatus", apiAuth.GetUserIpLocation)

		authRouterGroup.POST("/change_password", middleware.JWTAuth(), apiAuth.ChangePassword)
	}
	// Third service
	thirdGroup := r.Group("/third")
	{
		thirdGroup.POST("/tencent_cloud_storage_credential", middleware.JWTAuth(), apiThird.TencentCloudStorageCredential)
		thirdGroup.POST("/tencent_cloud_storage_upload", middleware.JWTAuth(), apiThird.TencentCloudUploadFile)
		thirdGroup.POST("/tencent_cloud_storage_upload_persistence", middleware.JWTAuth(), apiThird.TencentCloudUploadPersistentFile)
		thirdGroup.POST("/tencent_cloud_storage_multi_upload", middleware.JWTAuth(), apiThird.TencentCloudMultiUploadFile)
		thirdGroup.POST("/tencent_cloud_storage_multi_upload_persistence", middleware.JWTAuth(), apiThird.TencentCloudMultiUploadPersistentFile)
		thirdGroup.GET("/tencent_cloud_storage_multi_upload_process", apiThird.TencentCloudMultiUploadFileProcess)
		thirdGroup.POST("/ali_oss_credential", middleware.JWTAuth(), apiThird.AliOSSCredential)
		thirdGroup.POST("/minio_storage_credential", middleware.JWTAuth(), apiThird.MinioStorageCredential)
		thirdGroup.POST("/minio_upload", middleware.JWTAuth(), apiThird.MinioUploadFile)
		thirdGroup.POST("/minio_upload_persistence", middleware.JWTAuth(), apiThird.MinioUploadPersistentFile)
		thirdGroup.POST("/upload_update_app", middleware.JWTAuth(), apiThird.UploadUpdateApp)
		thirdGroup.POST("/get_download_url", middleware.JWTAuth(), apiThird.GetDownloadURL)
		thirdGroup.POST("/get_rtc_invitation_info", middleware.JWTAuth(), apiThird.GetRTCInvitationInfo)
		thirdGroup.POST("/get_rtc_invitation_start_app", middleware.JWTAuth(), apiThird.GetRTCInvitationInfoStartApp)

		thirdGroup.POST("/check_status", userMiddleWere.JWTAuth(), apiThird.CheckStatus)
		thirdGroup.POST("/start_communication", userMiddleWere.JWTAuth(), apiThird.StartCommunication)
		thirdGroup.POST("/join_communication", userMiddleWere.JWTAuth(), apiThird.JoinCommunication)
		thirdGroup.POST("/end_communication", userMiddleWere.JWTAuth(), apiThird.EndCommunication)
		thirdGroup.POST("/get_members_by_communication_id", userMiddleWere.JWTAuth(), apiThird.GetMembersByCommunicationID)
		thirdGroup.POST("/record_callback", apiThird.RecordCallBack)
	}
	// Message
	chatGroup := r.Group("/msg")
	{
		chatGroup.POST("/newest_seq", middleware.JWTAuth(), apiChat.GetSeq)
		chatGroup.POST("/get_max_seq", middleware.JWTAuth(), apiChat.GetMaxSeq)
		chatGroup.POST("/get_group_min_seq", middleware.JWTAuth(), apiChat.GetGroupMinSeq)
		chatGroup.POST("/send_msg", middleware.JWTAuth(), apiChat.SendMsg)
		chatGroup.POST("/pull_msg_by_seq", middleware.JWTAuth(), apiChat.PullMsgBySeqList)
		chatGroup.POST("/del_msg", middleware.JWTAuth(), apiChat.DelMsg)
		chatGroup.POST("/clear_msg", middleware.JWTAuth(), apiChat.ClearMsg)
		chatGroup.POST("/switch_broadcast", userMiddleWere.JWTAuth(), apiChat.SwitchBroadcast)
		chatGroup.POST("/get_broadcast_status", userMiddleWere.JWTAuth(), apiChat.GetBroadcastStatus)
	}
	// Manager
	managementGroup := r.Group("/manager")
	{
		managementGroup.POST("/delete_user", middleware.JWTAuth(), manage.DeleteUser) // 1
		managementGroup.POST("/send_msg", middleware.JWTAuth(), manage.ManagementSendMsg)
		managementGroup.POST("/batch_send_msg", middleware.JWTAuth(), manage.ManagementBatchSendMsg)
		managementGroup.POST("/batch_send_msg_v2", middleware.JWTAuth(), manage.ManagementBatchSendMsgV2)
		managementGroup.POST("/get_all_users_uid", middleware.JWTAuth(), manage.GetAllUsersUid)             // 1
		managementGroup.POST("/account_check", middleware.JWTAuth(), manage.AccountCheck)                   // 1
		managementGroup.POST("/get_users_online_status", middleware.JWTAuth(), manage.GetUsersOnlineStatus) // 1
	}
	// Conversation
	conversationGroup := r.Group("/conversation")
	{ // 1
		conversationGroup.POST("/get_all_conversations", middleware.JWTAuth(), conversation.GetAllConversations)
		conversationGroup.POST("/get_conversation", middleware.JWTAuth(), conversation.GetConversation)
		conversationGroup.POST("/get_conversations", middleware.JWTAuth(), conversation.GetConversations)
		conversationGroup.POST("/set_conversation", middleware.JWTAuth(), conversation.SetConversation)
		conversationGroup.POST("/batch_set_conversation", middleware.JWTAuth(), conversation.BatchSetConversations)
		conversationGroup.POST("/set_recv_msg_opt", middleware.JWTAuth(), conversation.SetRecvMsgOpt)
		conversationGroup.POST("/modify_conversation_field", middleware.JWTAuth(), conversation.ModifyConversationField)
	}
	// office
	// officeGroup := r.Group("/office")
	// {
	//	officeGroup.POST("/get_user_tags", office.GetUserTags)
	//	officeGroup.POST("/get_user_tag_by_id", office.GetUserTagByID)
	//	officeGroup.POST("/create_tag", office.CreateTag)
	//	officeGroup.POST("/delete_tag", office.DeleteTag)
	//	officeGroup.POST("/set_tag", office.SetTag)
	//	officeGroup.POST("/send_msg_to_tag", office.SendMsg2Tag)
	//	officeGroup.POST("/get_send_tag_log", office.GetTagSendLogs)
	//
	//	officeGroup.POST("/create_one_work_moment", office.CreateOneWorkMoment)
	//	officeGroup.POST("/delete_one_work_moment", office.DeleteOneWorkMoment)
	//	officeGroup.POST("/like_one_work_moment", office.LikeOneWorkMoment)
	//	officeGroup.POST("/comment_one_work_moment", office.CommentOneWorkMoment)
	//	officeGroup.POST("/get_work_moment_by_id", office.GetWorkMomentByID)
	//	officeGroup.POST("/get_user_work_moments", office.GetUserWorkMoments)
	//	officeGroup.POST("/get_user_friend_work_moments", office.GetUserFriendWorkMoments)
	//	officeGroup.POST("/set_user_work_moments_level", office.SetUserWorkMomentsLevel)
	//	officeGroup.POST("/delete_comment", office.DeleteComment)
	// }

	// organizationGroup := r.Group("/organization")
	// {
	//	organizationGroup.POST("/create_department", organization.CreateDepartment)
	//	organizationGroup.POST("/update_department", organization.UpdateDepartment)
	//	organizationGroup.POST("/get_sub_department", organization.GetSubDepartment)
	//	organizationGroup.POST("/delete_department", organization.DeleteDepartment)
	//	organizationGroup.POST("/get_all_department", organization.GetAllDepartment)
	//
	//	organizationGroup.POST("/create_organization_user", organization.CreateOrganizationUser)
	//	organizationGroup.POST("/update_organization_user", organization.UpdateOrganizationUser)
	//	organizationGroup.POST("/delete_organization_user", organization.DeleteOrganizationUser)
	//
	//	organizationGroup.POST("/create_department_member", organization.CreateDepartmentMember)
	//	organizationGroup.POST("/get_user_in_department", organization.GetUserInDepartment)
	//	organizationGroup.POST("/update_user_in_department", organization.UpdateUserInDepartment)
	//
	//	organizationGroup.POST("/get_department_member", organization.GetDepartmentMember)
	//	organizationGroup.POST("/delete_user_in_department", organization.DeleteUserInDepartment)
	//
	// }

	discoverGroup := r.Group("/discover")
	{
		discoverGroup.POST("/get_discover_url", discover.GetDiscoverUrl)
	}

	privacyGroup := r.Group("/privacy")
	{
		privacyGroup.POST("/get_privacy", userMiddleWere.JWTAuth(), privacy.GetPrivacy)
		privacyGroup.POST("/set_privacy", userMiddleWere.JWTAuth(), privacy.SetPrivacy)
	}

	versionUpdateGroup := r.Group("/version")
	{
		versionUpdateGroup.POST("/latest", appversion.GetLatestAppVersion)
	}

	InviteCodeGroup := r.Group("/invite_code")
	{
		InviteCodeGroup.POST("/get_link", userMiddleWere.JWTAuth(), invite_code.GetInviteCodeLink)
		InviteCodeGroup.POST("/get_total_invitation", userMiddleWere.JWTAuth(), invite_code.GetTotalInvitation)
	}

	MomentsGroup := r.Group("/moments")
	{
		MomentsGroup.POST("/create", userMiddleWere.JWTAuth(), moments.CreateMoment)
		MomentsGroup.POST("/like", userMiddleWere.JWTAuth(), moments.CreateMomentLike)
		MomentsGroup.POST("/cancelLike", userMiddleWere.JWTAuth(), moments.CancelMomentLike)
		MomentsGroup.POST("/comment", userMiddleWere.JWTAuth(), moments.CreateMomentComment)
		MomentsGroup.POST("/comment_reply", userMiddleWere.JWTAuth(), moments.CreateReplyOfMomentComment)
		MomentsGroup.POST("/timeline", userMiddleWere.JWTAuth(), moments.GetListHomeTimeLineOfMoments)
		MomentsGroup.POST("/detail_comments", userMiddleWere.JWTAuth(), moments.GetMomentDetailsByID)
		MomentsGroup.POST("/comments_paging", userMiddleWere.JWTAuth(), moments.GetMomentCommentsByID)
		MomentsGroup.POST("/repost_moment", userMiddleWere.JWTAuth(), moments.RepostAMoment)
		MomentsGroup.POST("/delete_moment", userMiddleWere.JWTAuth(), moments.DeleteMoment)
		MomentsGroup.POST("/delete_moment_comment", userMiddleWere.JWTAuth(), moments.DeleteMomentComment)
		MomentsGroup.POST("/global_search_moments", userMiddleWere.JWTAuth(), moments.GlobalSearchInMoments)
		MomentsGroup.POST("/get_moment_any_userid_have_media", userMiddleWere.JWTAuth(), moments.GetMomentAnyUserMediaByID)

		MomentsGroup.POST("/get_any_user_moments", userMiddleWere.JWTAuth(), moments.GetAnyUserMomentsByID)
		MomentsGroup.POST("/get_user_moments_count", userMiddleWere.JWTAuth(), moments.GetUserMomentCount)
	}

	InterestGroup := r.Group("/interest")
	{
		InterestGroup.POST("/set_interest", userMiddleWere.JWTAuth(), interest.SetInterest)
		InterestGroup.POST("/get_interest_group", userMiddleWere.JWTAuth(), interest.GetInterestGroup)
		InterestGroup.POST("/remove_group", userMiddleWere.JWTAuth(), interest.RemoveGroup)
	}

	OfficialGroup := r.Group("/official")
	{
		OfficialGroup.POST("/register", userMiddleWere.JWTAuth(), news.RegisterOfficial)
		OfficialGroup.POST("/get_self_info", userMiddleWere.JWTAuth(), news.GetSelfInfo)
		OfficialGroup.POST("/set_self_info", userMiddleWere.JWTAuth(), news.SetSelfInfo)
		OfficialGroup.POST("/follow", userMiddleWere.JWTAuth(), news.FollowOfficial)
		OfficialGroup.POST("/unfollow", userMiddleWere.JWTAuth(), news.UnfollowOfficial)
		OfficialGroup.POST("/update_follow_settings", userMiddleWere.JWTAuth(), news.UpdateOfficialFollowSettings)
		OfficialGroup.POST("/follow_settings_by_id", userMiddleWere.JWTAuth(), news.OfficialFollowSettingsByOfficialID)
		OfficialGroup.POST("/follow_list", userMiddleWere.JWTAuth(), news.UserFollowList)
		OfficialGroup.POST("/get_profile", userMiddleWere.JWTAuth(), news.GetOfficialProfile)
		OfficialGroup.POST("/block_follows", userMiddleWere.JWTAuth(), news.BlockOfficialFollows)
		OfficialGroup.POST("/unblock_follows", userMiddleWere.JWTAuth(), news.UnblockOfficialFollows)
		OfficialGroup.POST("/delete_follows", userMiddleWere.JWTAuth(), news.DeleteOfficialFollows)
		OfficialGroup.POST("/create_article", userMiddleWere.JWTAuth(), news.CreateArticle)
		OfficialGroup.POST("/update_article", userMiddleWere.JWTAuth(), news.UpdateArticle)
		OfficialGroup.POST("/delete_article", userMiddleWere.JWTAuth(), news.DeleteArticle)
		OfficialGroup.POST("/list_articles", userMiddleWere.JWTAuth(), news.ListArticles)
		OfficialGroup.POST("/get_article", userMiddleWere.JWTAuth(), news.GetOfficialArticle)
		OfficialGroup.POST("/list_self_follows", userMiddleWere.JWTAuth(), news.ListOfficialSelfFollows)
		OfficialGroup.POST("/list_article_likes", userMiddleWere.JWTAuth(), news.ListArticleLikes)
		OfficialGroup.POST("/delete_article_like", userMiddleWere.JWTAuth(), news.DeleteArticleLike)
		OfficialGroup.POST("/add_article_comment", userMiddleWere.JWTAuth(), news.AddOfficialArticleComment)
		OfficialGroup.POST("/list_article_comments", userMiddleWere.JWTAuth(), news.ListArticleComments)
		OfficialGroup.POST("/list_article_comment_replies", userMiddleWere.JWTAuth(), news.ListArticleCommentReplies)
		OfficialGroup.POST("/delete_article_comment", userMiddleWere.JWTAuth(), news.OfficialDeleteArticleComment)
		OfficialGroup.POST("/show_article_comment", userMiddleWere.JWTAuth(), news.OfficialShowArticleComment)
		OfficialGroup.POST("/hide_article_comment", userMiddleWere.JWTAuth(), news.OfficialHideArticleComment)
		OfficialGroup.POST("/like_article_comment", userMiddleWere.JWTAuth(), news.OfficialLikeArticleComment)
		OfficialGroup.POST("/unlike_article_comment", userMiddleWere.JWTAuth(), news.OfficialUnlikeArticleComment)
		OfficialGroup.POST("/get_recent_analytics", userMiddleWere.JWTAuth(), news.GetRecentAnalytics)
		OfficialGroup.POST("/get_analytics_by_day", userMiddleWere.JWTAuth(), news.GetAnalyticsByDay)
		OfficialGroup.POST("/search", userMiddleWere.JWTAuth(), news.SearchOfficialAccounts)
		OfficialGroup.POST("/followed_official_conv", userMiddleWere.JWTAuth(), news.GetFollowedOfficialConversation)
		OfficialGroup.POST("/check_id_number", userMiddleWere.JWTAuth(), news.GetOfficialIDNumberAvailability)
	}

	ArticleGroup := r.Group("/article")
	{
		ArticleGroup.POST("/like", userMiddleWere.JWTAuth(), news.LikeArticle)
		ArticleGroup.POST("/unlike", userMiddleWere.JWTAuth(), news.UnlikeArticle)
		ArticleGroup.POST("/comment", userMiddleWere.JWTAuth(), news.AddArticleComment)
		ArticleGroup.POST("/list_comments", userMiddleWere.JWTAuth(), news.ListUserArticleComments)
		ArticleGroup.POST("/list_comment_replies", userMiddleWere.JWTAuth(), news.ListUserArticleCommentReplies)
		ArticleGroup.POST("/like_comment", userMiddleWere.JWTAuth(), news.LikeArticleComment)
		ArticleGroup.POST("/unlike_comment", userMiddleWere.JWTAuth(), news.UnlikeArticleComment)
		ArticleGroup.POST("/timeline", userMiddleWere.JWTAuth(), news.ListArticlesTimeLine)
		ArticleGroup.POST("/search", userMiddleWere.JWTAuth(), news.SearchArticles)
		ArticleGroup.POST("/get_by_id", userMiddleWere.JWTAuth(), news.GetUserArticleByArticleID)
		ArticleGroup.POST("/insert_read", userMiddleWere.JWTAuth(), news.InsertArticleRead)
		ArticleGroup.POST("/read_list", userMiddleWere.JWTAuth(), news.ListUserArticleReads)
		ArticleGroup.POST("/clear_read_list", userMiddleWere.JWTAuth(), news.ClearUserArticleReads)
		ArticleGroup.POST("/delete_comment", userMiddleWere.JWTAuth(), news.DeleteArticleComment)
	}

	FavoriteGroup := r.Group("/favorites")
	{
		FavoriteGroup.POST("/add_favorite", userMiddleWere.JWTAuth(), favorite.AddFavorite)
		FavoriteGroup.POST("/get_favorite_list", userMiddleWere.JWTAuth(), favorite.GetFavoriteList)
		FavoriteGroup.POST("/remove_favorite", userMiddleWere.JWTAuth(), favorite.RemoveFavorite)
	}

	OauthGroup := r.Group("/oauth")
	{
		OauthGroup.POST("/binding_list", middleware.JWTAuth(), oauth.BindingList)

		OauthGroup.POST("/binding_facebook", middleware.JWTAuth(), oauth.BindingFacebook)
		OauthGroup.POST("/binding_google", middleware.JWTAuth(), oauth.BindingGoogle)
		OauthGroup.POST("/binding_apple", middleware.JWTAuth(), oauth.BindingApple)
		OauthGroup.POST("/binding_email", middleware.JWTAuth(), oauth.BindingEmail)
		OauthGroup.POST("/binding_phone", middleware.JWTAuth(), oauth.BindingPhone)

		OauthGroup.POST("/unbinding", middleware.JWTAuth(), oauth.Unbinding)
	}

	FamiliarGroup := r.Group("/familiar")
	{
		FamiliarGroup.POST("/sync_contact", middleware.JWTAuth(), familiar.SyncContact)
		FamiliarGroup.POST("/get_familiar_list", middleware.JWTAuth(), familiar.GetFamiliarList)
		FamiliarGroup.POST("/remove_user", middleware.JWTAuth(), familiar.RemoveUser)
	}

	gameStoreGroup := r.Group("/game_store")
	{
		// game list
		gameStoreGroup.POST("/banner_games", userMiddleWere.JWTAuth(), gameStore.BannerGames)
		gameStoreGroup.POST("/today_recommendations", userMiddleWere.JWTAuth(), gameStore.TodayRecommendations)
		gameStoreGroup.POST("/popular_games", userMiddleWere.JWTAuth(), gameStore.PopularGames)
		gameStoreGroup.POST("/all_games", userMiddleWere.JWTAuth(), gameStore.AllGames)
		gameStoreGroup.POST("/get_categories", userMiddleWere.JWTAuth(), gameStore.GetCategories)

		// game name
		gameStoreGroup.POST("/search_name", userMiddleWere.JWTAuth(), gameStore.SearchName)
		gameStoreGroup.POST("/search_game_list_by_name", userMiddleWere.JWTAuth(), gameStore.SearchGameListByName)

		// play game record, add click count, history
		gameStoreGroup.POST("/play_game", userMiddleWere.JWTAuth(), gameStore.PlayGame)

		// history
		gameStoreGroup.POST("/get_history", userMiddleWere.JWTAuth(), gameStore.GetGameHistory)

		// favorite
		gameStoreGroup.POST("/get_favorites", userMiddleWere.JWTAuth(), gameStore.GetGameFavorites)
		gameStoreGroup.POST("/remove_favorite", userMiddleWere.JWTAuth(), gameStore.RemoveGameFavorite)
		gameStoreGroup.POST("/add_favorite", userMiddleWere.JWTAuth(), gameStore.AddGameFavorite)

		// detail
		gameStoreGroup.POST("/game_details", userMiddleWere.JWTAuth(), gameStore.GameDetails)

	}

	mePageGroup := r.Group("/me_page")
	{
		mePageGroup.POST("/get_all_me_page_urls", userMiddleWere.JWTAuth(), me_page_url.GetMePageURLs)
	}

	otcGroup := r.Group("/otc")
	{
		otcGroup.POST("/get_otc_url", userMiddleWere.JWTAuth(), me_page_url.GetOTCUrl)
	}

	depositGroup := r.Group("/deposit")
	{
		depositGroup.POST("/get_deposit", userMiddleWere.JWTAuth(), me_page_url.DepositURL)
	}
	withdrawGroup := r.Group("/withdraw")
	{
		withdrawGroup.POST("/get_withdraw_url", userMiddleWere.JWTAuth(), me_page_url.WithdrawURL)
	}
	exchangeGroup := r.Group("/exchange")
	{
		exchangeGroup.POST("/get_exchange_url", userMiddleWere.JWTAuth(), me_page_url.ExchangeURL)
	}
	marketGroup := r.Group("/market")
	{
		marketGroup.POST("/get_market_url", userMiddleWere.JWTAuth(), me_page_url.MarketURL)
	}
	earnGroup := r.Group("/earn")
	{
		earnGroup.POST("/get_earn_url", userMiddleWere.JWTAuth(), me_page_url.EarnURL)
	}

	ShortVideoGroup := r.Group("/short_video")
	{
		ShortVideoGroup.POST("/create", userMiddleWere.JWTAuth(), short_video.CreateShortVideo)

		ShortVideoGroup.POST("/get_user_count", userMiddleWere.JWTAuth(), short_video.GetUserCount)
		ShortVideoGroup.POST("/notices", userMiddleWere.JWTAuth(), short_video.GetUserNotices)

		ShortVideoGroup.POST("/search_short_video", userMiddleWere.JWTAuth(), short_video.SearchShortVideo)

		ShortVideoGroup.POST("/get_sign", userMiddleWere.JWTAuth(), short_video.GetUpdateShortVideoSign)
		ShortVideoGroup.POST("/get_short_video", userMiddleWere.JWTAuth(), short_video.GetShortVideoByFileId)
		ShortVideoGroup.POST("/get_like_short_video_list", userMiddleWere.JWTAuth(), short_video.GetLikeShortVideoList)
		ShortVideoGroup.POST("/get_short_video_list_by_user_id", userMiddleWere.JWTAuth(), short_video.GetShortVideoListByUserId)
		ShortVideoGroup.POST("/short_video_like", userMiddleWere.JWTAuth(), short_video.ShortVideoLike)
		ShortVideoGroup.POST("/short_video_comment", userMiddleWere.JWTAuth(), short_video.ShortVideoComment)
		ShortVideoGroup.POST("/delete_short_video_comment", userMiddleWere.JWTAuth(), short_video.DeleteShortVideoComment)
		ShortVideoGroup.POST("/get_short_video_comment", userMiddleWere.JWTAuth(), short_video.GetShortVideoComment)
		ShortVideoGroup.POST("/short_video_comment_like", userMiddleWere.JWTAuth(), short_video.ShortVideoCommentLike)
		ShortVideoGroup.POST("/get_recommend", userMiddleWere.JWTAuth(), short_video.GetRecommendList)

		ShortVideoGroup.POST("/follow", userMiddleWere.JWTAuth(), short_video.ShortVideoFollow)
		ShortVideoGroup.POST("/follow_list", userMiddleWere.JWTAuth(), short_video.ShortVideoFollowList)
		ShortVideoGroup.POST("/fans_list", userMiddleWere.JWTAuth(), short_video.ShortVideoFansList)
		ShortVideoGroup.POST("/get_follow_short_video_list", userMiddleWere.JWTAuth(), short_video.GetFollowShortVideoList)

		ShortVideoGroup.POST("/comment_page", userMiddleWere.JWTAuth(), short_video.CommentPage)
		ShortVideoGroup.POST("/comment_page_reply_list", userMiddleWere.JWTAuth(), short_video.CommentPageReplyList)

		ShortVideoGroup.POST("/block_short_video", userMiddleWere.JWTAuth(), short_video.BlockShortVideo)

	}

	CallBackGroup := r.Group("/callback")
	{
		CallBackGroup.POST("/vod", callback.Vod)
	}

	QrLoginGroup := r.Group("/qr_login")
	{
		// pc
		QrLoginGroup.POST("/get_qr_code", qr_login.GetQrCode)
		QrLoginGroup.POST("/check_state", qr_login.CheckState)

		// mobile
		QrLoginGroup.POST("/push_qr_code", userMiddleWere.JWTAuth(), qr_login.PushQrCode)
		QrLoginGroup.POST("/confirm_login", userMiddleWere.JWTAuth(), qr_login.ConfirmQrCode)
	}

	go apiThird.MinioInit()
	defaultPorts := config.Config.Api.GinPort
	ginPort := flag.Int("port", defaultPorts[0], "get ginServerPort from cmd,default 10002 as port")
	flag.Parse()
	address := "0.0.0.0:" + strconv.Itoa(*ginPort)
	if config.Config.Api.ListenIP != "" {
		address = config.Config.Api.ListenIP + ":" + strconv.Itoa(*ginPort)
	}
	address = config.Config.Api.ListenIP + ":" + strconv.Itoa(*ginPort)
	fmt.Println("start api server, address: ", address)
	err := r.Run(address)
	if err != nil {
		log.Error("", "run failed ", *ginPort, err.Error())
	}
}
