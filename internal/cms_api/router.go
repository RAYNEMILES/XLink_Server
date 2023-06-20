package cms_api

import (
	"Open_IM/internal/api/domain"
	"Open_IM/internal/cms_api/admin"
	"Open_IM/internal/cms_api/appversion"
	"Open_IM/internal/cms_api/blacklist"
	"Open_IM/internal/cms_api/communication"
	"Open_IM/internal/cms_api/discover"
	"Open_IM/internal/cms_api/favorites"
	"Open_IM/internal/cms_api/game_store"
	"Open_IM/internal/cms_api/group"
	"Open_IM/internal/cms_api/guest"
	"Open_IM/internal/cms_api/interest"
	inviteChannelCode "Open_IM/internal/cms_api/invite_channel_code"
	inviteCode "Open_IM/internal/cms_api/invite_code"
	"Open_IM/internal/cms_api/me_page"
	messageCMS "Open_IM/internal/cms_api/message_cms"
	"Open_IM/internal/cms_api/middleware"
	"Open_IM/internal/cms_api/moments"
	"Open_IM/internal/cms_api/news"
	"Open_IM/internal/cms_api/shortVideo"
	"Open_IM/internal/cms_api/statistics"
	"Open_IM/internal/cms_api/user"

	"github.com/gin-gonic/gin"
)

func NewGinRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	baseRouter := gin.Default()
	baseRouter.Use(middleware.CorsHandler())
	router := baseRouter.Group("/cms")
	router.Use(middleware.CorsHandler())
	adminRouterGroup := router.Group("/admin")
	{
		//adminRouterGroup.POST("/login", admin.AdminLogin)

		adminRouterGroup.POST("/admin_login", admin.AdminLogin_v2)
		adminRouterGroup.POST("/admin_verify_totp", admin.VerifyTOTPAdminUser)
		adminRouterGroup.POST("/admin_change_password", middleware.JWTAuth(), admin.ChangePasswordTOTPAdminUser)
		adminRouterGroup.GET("/permissions", middleware.JWTAuth(), admin.GetPermissionsOfAdminUser)
		adminRouterGroup.POST("/permissions_by_admin_id", middleware.JWTAuth(), admin.GetPermissionsOfAdminUserByID)

		adminRouterGroup.POST("/add_admin_user", middleware.JWTAuth(), admin.AddAdminUser)
		adminRouterGroup.POST("/delete_admin_user", middleware.JWTAuth(), admin.DeleteAdminUser)
		adminRouterGroup.POST("/alter_admin_user", middleware.JWTAuth(), admin.AlterAdminUser)
		adminRouterGroup.GET("/get_admin_users", middleware.JWTAuth(), admin.GetAdminUsers)
		adminRouterGroup.POST("/admin_switch_status", middleware.JWTAuth(), admin.SwitchAdminUserStatus)
		adminRouterGroup.POST("/search_users", middleware.JWTAuth(), admin.SearchAdminUsers)

		adminRouterGroup.POST("/add_admin_role", middleware.JWTAuth(), admin.AddAdminRole)
		adminRouterGroup.POST("/alter_admin_role", middleware.JWTAuth(), admin.AlterAdminRole)
		adminRouterGroup.POST("/delete_admin_role", middleware.JWTAuth(), admin.DeleteAdminRole)
		adminRouterGroup.GET("/all_admin_roles", middleware.JWTAuth(), admin.GetAllAdminRoles)
		adminRouterGroup.POST("/search_admin_roles", middleware.JWTAuth(), admin.SearchAminRoles)

		adminRouterGroup.POST("/add_api_admin_role", middleware.JWTAuth(), admin.AddApiAdminRole)
		adminRouterGroup.POST("/alter_api_admin_role", middleware.JWTAuth(), admin.AlterApiAdminRole)
		adminRouterGroup.POST("/enb_disable_api_admin_role", middleware.JWTAuth(), admin.EnDisableApiAdminRole)
		adminRouterGroup.POST("/delete_api_admin_role", middleware.JWTAuth(), admin.DeleteApiAdminRole)
		adminRouterGroup.GET("/all_api_admin_roles", middleware.JWTAuth(), admin.GetAllApiAdminRoles)
		adminRouterGroup.POST("/search_api_admin_roles", middleware.JWTAuth(), admin.SearchApiAdminRoles)

		adminRouterGroup.POST("/add_page_admin_role", middleware.JWTAuth(), admin.AddPageAdminRole)
		adminRouterGroup.POST("/alter_page_admin_role", middleware.JWTAuth(), admin.AlterPageAdminRole)
		adminRouterGroup.POST("/delete_page_admin_role", middleware.JWTAuth(), admin.DeletePageAdminRole)
		adminRouterGroup.GET("/all_page_admin_roles", middleware.JWTAuth(), admin.GetAllPageAdminRoles)
		adminRouterGroup.POST("/search_page_admin_roles", middleware.JWTAuth(), admin.SearchPageAdminRoles)

		adminRouterGroup.POST("/search_operation_logs", middleware.JWTAuth(), admin.SearchOperationLogs)

		// adminRouterGroup.POST("/add_admin_action", middleware.JWTAuth(), admin.AddAdminAction)
		// adminRouterGroup.POST("/alter_admin_action", middleware.JWTAuth(), admin.AlterAdminAction)
		// adminRouterGroup.POST("/delete_admin_action", middleware.JWTAuth(), admin.DeleteAdminAction)
		// adminRouterGroup.GET("/all_admin_action", middleware.JWTAuth(), admin.GetAllAdminAction)

		adminRouterGroup.POST("/alterGAuthStatus", middleware.JWTAuth(), admin.AlterGAuthStatus)
		adminRouterGroup.POST("/getGAuthQrCode", middleware.JWTAuth(), admin.GetgAuthQrCode)
	}

	r2 := router.Group("")
	r2.Use(middleware.JWTAuth())

	discoverRouterGroup := r2.Group("/discover")
	{
		discoverRouterGroup.POST("/get_discover", discover.GetDiscoverUrl)
		discoverRouterGroup.POST("/save_discover_url", discover.SaveDiscoverUrl)
		discoverRouterGroup.POST("/switch_discover", discover.SwitchDiscoverPage)
	}

	mePageGroup := r2.Group("/me_page")
	{
		mePageGroup.GET("/get_url", me_page.GetURL)
		mePageGroup.POST("/save_url", me_page.SaveUrl)
		mePageGroup.POST("/switch_status", me_page.SwitchStatus)
	}

	// version update api
	versionUpdateGroup := r2.Group("/version")
	{
		versionUpdateGroup.POST("/get", appversion.GetAppVersionByID)
		versionUpdateGroup.POST("/search", appversion.GetAppVersions)
		versionUpdateGroup.POST("/add", appversion.AddAppVersion)
		versionUpdateGroup.POST("/edit", appversion.EditAppVersion)
		versionUpdateGroup.POST("/remove", appversion.DeleteAppVersion)
	}

	statisticsRouterGroup := r2.Group("/statistics")
	{
		statisticsRouterGroup.GET("/get_messages_statistics", statistics.GetMessagesStatistics)
		statisticsRouterGroup.GET("/get_user_statistics", statistics.GetUserStatistics)
		statisticsRouterGroup.GET("/get_group_statistics", statistics.GetGroupStatistics)
		statisticsRouterGroup.GET("/get_active_user", statistics.GetActiveUser)
		statisticsRouterGroup.GET("/get_active_group", statistics.GetActiveGroup)
		statisticsRouterGroup.GET("/get_game_statistics", middleware.JWTAuth(), statistics.GetGameStatistics)
	}
	// organizationRouterGroup := r2.Group("/organization")
	// {
	//	organizationRouterGroup.GET("/get_staffs", organization.GetStaffs)
	//	organizationRouterGroup.GET("/get_organizations", organization.GetOrganizations)
	//	organizationRouterGroup.GET("/get_squad", organization.GetSquads)
	//	organizationRouterGroup.POST("/add_organization", organization.AddOrganization)
	//	organizationRouterGroup.POST("/alter_staff", organization.AlterStaff)
	//	organizationRouterGroup.GET("/inquire_organization", organization.InquireOrganization)
	//	organizationRouterGroup.POST("/alter_organization", organization.AlterOrganization)
	//	organizationRouterGroup.POST("/delete_organization", organization.DeleteOrganization)
	//	organizationRouterGroup.POST("/get_organization_squad", organization.GetOrganizationSquads)
	//	organizationRouterGroup.PATCH("/alter_corps_info", organization.AlterStaffsInfo)
	//	organizationRouterGroup.POST("/add_child_org", organization.AddChildOrganization)
	// }
	domainGroup := r2.Group("/domain")
	{
		domainGroup.POST("/get_all_domains", domain.GetAllDomains)
		domainGroup.POST("/save_domains", domain.SaveAllDomains)
	}
	groupRouterGroup := r2.Group("/group")
	{
		groupRouterGroup.GET("/get_group_by_id", group.GetGroupById)
		groupRouterGroup.GET("/get_groups", group.GetGroups)
		groupRouterGroup.GET("/get_group_by_name", group.GetGroupByName)
		groupRouterGroup.GET("/get_group_members", group.GetGroupMembers)
		groupRouterGroup.POST("/create_group", group.CreateGroup)
		groupRouterGroup.POST("/add_members", group.AddGroupMembers)
		groupRouterGroup.POST("/remove_members", group.RemoveGroupMembers)
		groupRouterGroup.POST("/ban_group_private_chat", group.BanPrivateChat)
		groupRouterGroup.POST("/open_group_private_chat", group.OpenPrivateChat)
		groupRouterGroup.POST("/ban_group_chat", group.BanGroupChat)
		groupRouterGroup.POST("/open_group_chat", group.OpenGroupChat)
		groupRouterGroup.POST("/delete_group", group.DeleteGroup)
		groupRouterGroup.POST("/get_members_in_group", group.GetGroupMembers)
		groupRouterGroup.POST("/set_group_master", group.SetGroupMaster)
		groupRouterGroup.POST("/set_group_ordinary_user", group.SetGroupOrdinaryUsers)
		groupRouterGroup.POST("/set_group_admin", group.SetGroupAdmin)
		groupRouterGroup.POST("/alter_group_info", group.AlterGroupInfo)
		groupRouterGroup.POST("/mute_group_member", group.MuteGroupMember)
		groupRouterGroup.POST("/cancel_mute_group_member", group.CancelMuteGroupMember)
		groupRouterGroup.POST("/dismiss_group", group.DismissGroup)

		groupRouterGroup.POST("/set_video_audio_status", group.SetVideoAudioStatus)
		groupRouterGroup.POST("/set_user_video_audio_status", group.SetUserVideoAudioStatus)

		groupRouterGroup.GET("/get_users_by_group", group.GetUsersByGroup)

	}
	userRouterGroup := r2.Group("/user")
	{
		userRouterGroup.POST("/resign", user.ResignUser)
		userRouterGroup.GET("/get_user", user.GetUserById)
		userRouterGroup.POST("/alter_user", user.AlterUser)
		userRouterGroup.GET("/get_users", user.GetUsers)
		userRouterGroup.GET("/get_users_third_info", user.GetUsersThirdInfo)
		userRouterGroup.POST("/add_user", user.AddUser)
		userRouterGroup.POST("/multi_add_user", user.MultiAddUser)
		userRouterGroup.POST("/user_exists", user.Exists)
		userRouterGroup.POST("/unblock_user", user.UnblockUser)
		userRouterGroup.POST("/block_user", user.BlockUser)
		userRouterGroup.GET("/get_block_users", user.GetBlockUsers)
		userRouterGroup.GET("/get_block_user", user.GetBlockUserById)
		userRouterGroup.POST("/delete_user", user.DeleteUser)
		userRouterGroup.GET("/get_users_by_name", user.GetUsersByName)
		userRouterGroup.POST("/switch_status", user.SwitchStatus)
		// userRouterGroup.POST("/alter_add_friend_status", user.AlterAddFriendStatus)

		userRouterGroup.GET("/get_deleted_users", user.GetDeletedUsers)
	}
	friendRouterGroup := r2.Group("/friend")
	{
		friendRouterGroup.POST("/get_friends_by_id")
		friendRouterGroup.POST("/set_friend")
		friendRouterGroup.POST("/remove_friend")
	}

	guestRoutGroup := r2.Group("/guest")
	{
		guestRoutGroup.POST("/get_guest_status", guest.GetGuestStatus)
		guestRoutGroup.POST("/switch_guest_status", guest.SwitchGuestStatus)

		guestRoutGroup.POST("/get_guest_limit", guest.GetGuestLimit)
		guestRoutGroup.POST("/switch_guest_limit", guest.SwitchGuestLimit)
	}

	messageCMSRouterGroup := r2.Group("/message")
	{
		messageCMSRouterGroup.GET("/get_chat_logs", messageCMS.GetChatLogs)
		messageCMSRouterGroup.GET("/v1/get_chat_logs", messageCMS.GetChatLogsV1)
		messageCMSRouterGroup.POST("/broadcast_message", messageCMS.BroadcastMessage)
		messageCMSRouterGroup.POST("/mass_send_message", messageCMS.MassSendMassage)
		messageCMSRouterGroup.POST("/withdraw_message", messageCMS.WithdrawMessage)
	}
	inviteCodeRouterGroup := r2.Group("/invite_code")
	{
		inviteCodeRouterGroup.POST("/get_base_link", inviteCode.GetBaseLink)
		inviteCodeRouterGroup.POST("/set_base_link", inviteCode.SetBaseLink)

		inviteCodeRouterGroup.POST("/switch", inviteCode.Switch)
		inviteCodeRouterGroup.POST("/limit", inviteCode.Limit)
		inviteCodeRouterGroup.POST("/multi_delete", inviteCode.MultiDelete)

		inviteCodeRouterGroup.POST("/add_code", inviteCode.AddCode)
		inviteCodeRouterGroup.POST("/edit_code", inviteCode.EditCode)
		inviteCodeRouterGroup.POST("/switch_code_state", inviteCode.SwitchCodeState)
		inviteCodeRouterGroup.GET("/list", inviteCode.GetCodeList)
	}
	inviteChannelCodeRouterGroup := r2.Group("/invite_channel_code")
	{
		inviteChannelCodeRouterGroup.POST("/switch", inviteChannelCode.Switch)
		inviteChannelCodeRouterGroup.POST("/limit", inviteChannelCode.Limit)
		inviteChannelCodeRouterGroup.POST("/multi_delete", inviteChannelCode.MultiDelete)

		inviteChannelCodeRouterGroup.POST("/add_code", inviteChannelCode.AddCode)
		inviteChannelCodeRouterGroup.POST("/edit_code", inviteChannelCode.EditCode)
		inviteChannelCodeRouterGroup.POST("/switch_code_state", inviteChannelCode.SwitchCodeState)
		inviteChannelCodeRouterGroup.GET("/list", inviteChannelCode.GetCodeList)
	}
	interestRouterGroup := r2.Group("/interest")
	{
		interestRouterGroup.GET("/get_interests", interest.GetInterests)
		interestRouterGroup.POST("/delete_interests", interest.DeleteInterests)
		interestRouterGroup.POST("/alter_interest", interest.AlterInterests)
		interestRouterGroup.POST("/change_interest_status", interest.ChangeInterestStatus)
		interestRouterGroup.POST("/add_interests", interest.AddInterests)
		interestRouterGroup.POST("/add_one_interest", interest.AddOneInterest)

		interestRouterGroup.GET("/get_user_interests", interest.GetUserInterests)
		interestRouterGroup.POST("/alter_user_interests", interest.AlterUserInterests)
		interestRouterGroup.POST("/delete_user_interests", interest.DeleteUserInterests)

		// group name/id interest name
		interestRouterGroup.GET("/get_group_interests", interest.GetGroupInterests)
		interestRouterGroup.POST("/alter_group_interests", interest.AlterGroupInterests)
	}

	newsRouterGroup := r2.Group("/news")
	{
		newsRouterGroup.GET("/get_news", news.GetNews)
		newsRouterGroup.POST("/delete_news", news.DeleteNews)
		newsRouterGroup.POST("/alter_news", news.AlterNews)
		newsRouterGroup.POST("/change_privacy", news.ChangePrivacy)

		newsRouterGroup.GET("/get_news_comments", news.GetNewsComments)
		newsRouterGroup.POST("/remove_news_comments", news.RemoveNewsComments)
		newsRouterGroup.POST("/alter_news_comment", news.AlterNewsComment)
		newsRouterGroup.POST("/change_news_comment_status", news.ChangeNewsCommentStatus)

		newsRouterGroup.GET("/get_news_likes", news.GetNewsLikes)
		newsRouterGroup.POST("/remove_news_likes", news.RemoveNewsLikes)
		newsRouterGroup.POST("/change_news_like_status", news.ChangeNewsLikeStatus)

		newsRouterGroup.GET("/get_repost_articles", news.GetRepostArticles)
		newsRouterGroup.POST("/change_repost_privacy", news.ChangeRepostPrivacy)
		newsRouterGroup.POST("/delete_reposts", news.DeleteReposts)

		newsRouterGroup.GET("/get_official_accounts", news.GetOfficialAccounts)
		newsRouterGroup.POST("/delete_official_accounts", news.DeleteOfficialAccounts)
		newsRouterGroup.POST("/alter_official_account", news.AlterOfficialAccount)
		newsRouterGroup.POST("/add_official_account", news.AddOfficialAccount)
		newsRouterGroup.POST("/process", news.Process)

		newsRouterGroup.GET("/get_official_followers", news.GetOfficialFollowers)
		newsRouterGroup.POST("/mute_follower", news.MuteFollower)
		newsRouterGroup.POST("/block_follower", news.BlockFollower)
		newsRouterGroup.POST("/remove_followers", news.RemoveFollowers)
	}

	momentsRouterGroup := r2.Group("/moments")
	{
		momentsRouterGroup.GET("/get_moments", moments.GetMoments)
		momentsRouterGroup.POST("/delete_moments", moments.DeleteMoments)
		momentsRouterGroup.POST("/alter_moment", moments.AlterMoment)
		momentsRouterGroup.POST("/change_moment_status", moments.ChangeMomentStatus)
		momentsRouterGroup.POST("/modify_visibility", moments.ModifyVisibility)

		momentsRouterGroup.GET("/get_moment_details", moments.GetMomentDetails)
		momentsRouterGroup.POST("/ctl_comment", moments.CtlMomentComment)

		momentsRouterGroup.GET("/get_comments", moments.GetComments)
		momentsRouterGroup.POST("/remove_comments", moments.RemoveComments)
		momentsRouterGroup.POST("/alter_comment", moments.AlterComment)
		momentsRouterGroup.POST("/switch_comment_hide_state", moments.SwitchCommentHideState)

		momentsRouterGroup.GET("/get_likes", moments.GetLikes)
		momentsRouterGroup.POST("/remove_likes", moments.RemoveLikes)
		momentsRouterGroup.POST("/switch_like_hide_state", moments.SwitchLikeHideState)

		momentsRouterGroup.GET("/get_replay_comments", moments.GetReplayComments)
	}

	favoritesRouter := r2.Group("/favorites")
	{
		favoritesRouter.GET("/get_favorites", favorites.GetFavorites)
		favoritesRouter.POST("/delete_favorites", favorites.DeleteFavorites)
		favoritesRouter.POST("/alter_favorites", favorites.AlterFavorites)
	}

	communicationRouter := r2.Group("/communication")
	{
		communicationRouter.GET("/get_communications", communication.GetCommunications)
		communicationRouter.POST("/delete_communications", communication.DeleteCommunications)
		communicationRouter.POST("/interrupt_communications", communication.InterruptPersonalCommunications)
		communicationRouter.POST("/set_remark", communication.SetRemark)
	}

	gameStoreRouter := r2.Group("/game_store")
	{
		// game list
		gameStoreRouter.GET("/get_game_list", game_store.GetGameList)
		gameStoreRouter.POST("/edit_game", game_store.EditGame)
		gameStoreRouter.POST("/add_game", game_store.AddGame)
		gameStoreRouter.POST("/delete_games", game_store.DeleteGames)

		// categories
		gameStoreRouter.GET("/get_category", game_store.GetCategory)
		gameStoreRouter.POST("/add_category", game_store.AddCategory)
		gameStoreRouter.POST("/edit_category", game_store.EditCategory)
		gameStoreRouter.POST("/set_category_status", game_store.SetCategoryStatus)
		gameStoreRouter.POST("/delete_category", game_store.DeleteCategory)
	}

	blackListRouter := r2.Group("/blacks")
	{
		blackListRouter.GET("/get_blacks", blacklist.GetBlacks)
		blackListRouter.POST("/remove_black", blacklist.RemoveBlack)
		blackListRouter.POST("/alter_remark", blacklist.AlterRemark)
	}

	shortVideoRouter := r2.Group("/short_video")
	{
		// short video
		shortVideoRouter.GET("/get_short_video_list", shortVideo.GetShortVideoList)
		shortVideoRouter.POST("/delete_short_video", shortVideo.DeleteShortVideo)
		shortVideoRouter.POST("/alter_short_video", shortVideo.AlterShortVideo)

		// short video like
		shortVideoRouter.GET("/get_short_video_like_list", shortVideo.GetShortVideoLikeList)
		shortVideoRouter.POST("/delete_short_video_like", shortVideo.DeleteShortVideoLike)

		// short video comment
		shortVideoRouter.GET("/get_short_video_comment_list", shortVideo.GetShortVideoCommentList)
		shortVideoRouter.POST("/delete_short_video_comment", shortVideo.DeleteShortVideoComment)
		shortVideoRouter.POST("/alter_short_video_comment", shortVideo.AlterShortVideoComment)

		// short video interest label
		shortVideoRouter.GET("/get_short_video_interest_label_list", shortVideo.GetShortVideoInterestLabelList)
		shortVideoRouter.POST("/alter_short_video_interest_label", shortVideo.AlterShortVideoInterestLabel)

		shortVideoRouter.GET("/comment_replies", shortVideo.GetShortVideoCommentReplies)
		shortVideoRouter.POST("/alter_reply", shortVideo.AlterReply)
		shortVideoRouter.POST("/delete_replies", shortVideo.DeleteReplies)

		shortVideoRouter.GET("/comment_likes", shortVideo.GetShortVideoCommentLikes)
		shortVideoRouter.POST("/alter_like", shortVideo.AlterLike)
		shortVideoRouter.POST("/delete_likes", shortVideo.DeleteLikes)

		shortVideoRouter.GET("/get_followers", shortVideo.GetFollowers)
		shortVideoRouter.POST("/alter_follower", shortVideo.AlterFollower)
		shortVideoRouter.POST("/delete_followers", shortVideo.DeleteFollowers)

	}

	return baseRouter

}
