package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io/ioutil"
	"mihoyo-bbs-genshin-sign/internal/config"
	"mihoyo-bbs-genshin-sign/internal/model"
	"mihoyo-bbs-genshin-sign/internal/util"
	"net/http"
	"strings"
	"time"
)

func GetSignInfo(uid, cookie string) (signInfo *model.SignInfo, err error) {
	signInfo = &model.SignInfo{}
	var req *http.Request
	if req, err = http.NewRequest("GET", config.SignBaseUrl+config.SignAwardInfoUri, nil); err != nil {
		return
	}
	util.AddUrlQueryParametersFromStruct(req, model.SignUrlParam{
		ActId:  config.ActId,
		Uid:    uid,
		Region: getRegionFromUid(uid),
	}, config.HttpQueryTagName)
	util.AddHeadersFromMap(req, map[string]string{
		"Cookie":   cookie,
		"SignHost": config.SignHost,
	})

	// request and parse response
	var resp *http.Response
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	var respData = model.MihoyoResponse{Data: signInfo}
	if err = json.Unmarshal(body, &respData); err != nil {
		return
	}
	if respData.Retcode != 0 {
		return nil, fmt.Errorf("retcode: %d, message: %s", respData.Retcode, respData.Message)
	}
	return
}

func GetSignAwardList() (signAwardList *model.SignAwardList, err error) {
	signAwardList = &model.SignAwardList{}
	var req *http.Request
	if req, err = http.NewRequest("GET", config.SignBaseUrl+config.SignAwardHomeUri, nil); err != nil {
		return
	}
	util.AddUrlQueryParametersFromStruct(req, model.SignAwardsInfoReqParam{
		ActId: config.ActId,
	}, config.HttpQueryTagName)

	// request and parse response
	var resp *http.Response
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	var respData = model.MihoyoResponse{Data: signAwardList}
	if err = json.Unmarshal(body, &respData); err != nil {
		return
	}
	if respData.Retcode != 0 {
		return nil, fmt.Errorf("retcode: %d, message: %s", respData.Retcode, respData.Message)
	}
	return
}

func Sign(uid, cookie string) (err error) {
	var req *http.Request
	if req, err = http.NewRequest("POST", config.SignBaseUrl+config.SignAwardSignUri, nil); err != nil {
		return
	}

	util.AddUrlQueryParametersFromStruct(req, model.SignUrlParam{
		ActId:  config.ActId,
		Uid:    uid,
		Region: getRegionFromUid(uid),
	}, config.HttpQueryTagName)
	util.AddHeadersFromMap(req, map[string]string{
		"Cookie":            cookie,
		"SignHost":          config.SignHost,
		"x-rpc-client_type": config.XRpcClientType,
		"x-rpc-app_version": config.XRpcClientVersion,
		"x-rpc-device_id":   strings.Replace(uuid.NewString(), "-", "", -1),
		"DS":                getDs(),
	})

	// request and parse response
	var resp *http.Response
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	var respData = model.MihoyoResponse{}
	if err = json.Unmarshal(body, &respData); err != nil {
		return
	}
	if respData.Retcode != 0 && respData.Retcode != model.AlreadySigned {
		return fmt.Errorf("retcode: %d, message: %s", respData.Retcode, respData.Message)
	}
	return
}

func getDs() string {
	t := time.Now().Unix()
	r := util.GetRandString(6)
	hash := md5.Sum([]byte(fmt.Sprintf("salt=%s&t=%d&r=%s", config.DsSalt, t, r)))
	c := hex.EncodeToString(hash[:])
	return fmt.Sprintf("%d,%s,%s", t, r, c)
}

// getRegionFromUid get region according to the format of uid
func getRegionFromUid(uid string) string {
	if uid[0] == '5' {
		return config.RegionCnQd
	} else {
		return config.RegionCnGf
	}
}

func SignCronTask(db *gorm.DB) (err error) {
	var signItemList []model.SignItem
	if signItemList, err = model.FindAllSignItems(db); err != nil {
		return
	}
	for _, signItem := range signItemList {
		if err = Sign(signItem.Uid, signItem.Cookie); err != nil {
			log.Error(err)
		}
	}
	return
}
