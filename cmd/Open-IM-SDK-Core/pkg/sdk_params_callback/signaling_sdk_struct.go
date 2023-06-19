package sdk_params_callback

import (
	sdk "Open_IM/pkg/proto/sdk_ws"
)

type InviteCallback *sdk.SignalInviteReply

type InviteInGroupCallback *sdk.SignalInviteInGroupReply

type CancelCallback *sdk.SignalCancelReply

type RejectCallback *sdk.SignalRejectReply

type AcceptCallback *sdk.SignalAcceptReply

type HungUpCallback *sdk.SignalHungUpReply
