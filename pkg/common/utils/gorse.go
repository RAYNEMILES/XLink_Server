package utils

import (
	"Open_IM/pkg/common/config"
	"context"
	"encoding/json"
	"fmt"
	"github.com/zhenghaoz/gorse/client"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"io"
	"net/http"
	"strings"
	"time"
)

var gorse *client.GorseClient

func init() {
	gorse = client.NewGorseClient(config.Config.Gorse.Url, config.Config.Gorse.Token)
}

func InsertUser(userId string, interestIdList, followUserIdList []string) bool {
	user, err := gorse.GetUser(context.Background(), userId)
	if err == nil || user.UserId != "" {
		return true
	}

	_, err = gorse.InsertUser(context.Background(), client.User{
		UserId:    userId,
		Labels:    interestIdList,
		Subscribe: followUserIdList,
	})
	if err != nil {
		return false
	}
	return true
}

func DeleteUser(userId string) bool {
	_, err := gorse.DeleteUser(context.Background(), userId)
	if err != nil {
		return false
	}
	return true
}

func PatchUser(userId string, interestIdList, followUserIdList []string) bool {
	user, err := gorse.GetUser(context.Background(), userId)
	if err != nil || user.UserId == "" {
		return InsertUser(userId, interestIdList, followUserIdList)
	}

	body := client.User{
		UserId:    userId,
		Labels:    interestIdList,
		Subscribe: followUserIdList,
	}

	bodyByte, marshalErr := json.Marshal(body)
	if marshalErr != nil {
		return false
	}

	var req *http.Request
	var result any

	c := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	req, err = http.NewRequestWithContext(context.Background(), "PATCH", config.Config.Gorse.Url+fmt.Sprintf("/api/user/%s", userId), strings.NewReader(string(bodyByte)))
	if err != nil {
		return false
	}

	req.Header.Set("X-API-Key", config.Config.Gorse.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)

	if err != nil {
		return false
	}
	defer resp.Body.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	err = json.Unmarshal([]byte(buf.String()), &result)
	if err != nil {
		return false
	}
	return true
}

func InsertItem(itemId, categories, time string, interestId []string) bool {
	_, err := gorse.InsertItem(context.Background(), client.Item{
		ItemId:     itemId,
		Categories: []string{categories},
		Timestamp:  time,
		Labels:     interestId,
	})
	if err != nil {
		return false
	}
	return true
}

func UpdateItem(itemId string, isHidden bool, categories, desc string, interestId []string) bool {
	_, err := gorse.UpdateItem(context.Background(), itemId, client.ItemPatch{
		IsHidden:   &isHidden,
		Categories: []string{categories},
		Labels:     interestId,
		Comment:    &desc,
	})
	if err != nil {
		return false
	}
	return true
}

func InsertFeedback(feedBackType, userId, itemId string) bool {
	timeStamp := time.Now().UTC().Format(time.RFC3339)
	_, err := gorse.InsertFeedback(context.Background(), []client.Feedback{
		{FeedbackType: feedBackType, UserId: userId, ItemId: itemId, Timestamp: timeStamp},
	})
	if err != nil {
		return false
	}
	return true
}

func DeleteItem(fileId string) bool {
	_, err := gorse.DeleteItem(context.Background(), fileId)
	if err != nil {
		return false
	}
	return true
}

func GetRecommend(userId, categories string, size int) []string {
	result, err := gorse.GetRecommend(context.Background(), userId, categories, size)
	if err != nil {
		return []string{}
	}
	return result
}
