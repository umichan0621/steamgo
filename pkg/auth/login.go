package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"mime/multipart"
	"net/http"
	pb "steam/pkg/proto"
	"steam/pkg/utils"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

func (mgr *Core) Login() error {
	log.Info("Connecting to steam server...")
	// Get session id from steam store
	err := mgr.getSessionId()
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond * time.Duration(utils.RandRange(120, 300)))

	// Get RSA public key by proto message
	rsaRes := pb.CAuthentication_GetPasswordRSAPublicKey_Response{}
	err = mgr.getPasswordRSAPublicKey(&rsaRes)
	if err != nil {
		return err
	}
	encryptedPassword, err := mgr.encryptPassword(rsaRes.PublickeyMod, rsaRes.PublickeyExp)
	if err != nil {
		return err
	}

	time.Sleep(time.Millisecond * time.Duration(utils.RandRange(120, 300)))

	log.Infof("Try login as user: %s...", mgr.loginInfo.UserName)
	// Try begin auth
	beginAuthRes := pb.CAuthentication_BeginAuthSessionViaCredentials_Response{}
	err = mgr.beginAuthSessionViaCredentials(encryptedPassword, rsaRes.Timestamp,
		&beginAuthRes)
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond * time.Duration(utils.RandRange(120, 300)))

	// Handle confirmation if exist
	confirmationType := beginAuthRes.AllowedConfirmations.ConfirmationType
	if confirmationType != pb.EAuthSessionGuardType_k_EAuthSessionGuardType_None {
		log.Info("Need authentication...")
		updateAuthRes := pb.CAuthentication_UpdateAuthSessionWithSteamGuardCode_Response{}
		mgr.updateAuthSessionWithSteamGuardCode(beginAuthRes.ClientId, beginAuthRes.SteamId, confirmationType,
			&updateAuthRes)
	}

	log.Info("Logging in...")
	pollAuthRes := pb.CAuthentication_PollAuthSessionStatus_Response{}
	err = mgr.pollAuthSessionStatus(beginAuthRes.ClientId, beginAuthRes.RequestId, &pollAuthRes)
	if err != nil {
		return err
	}
	err = mgr.finalizeLogin(pollAuthRes.RefreshToken)
	if err != nil {
		return err
	}

	return nil
}

func (mgr *Core) getSessionId() error {
	httpReq, err := http.NewRequest("GET", kURI_STEAM_STROE, nil)
	if err != nil {
		return err
	}
	res, err := mgr.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("fail to connect steam store, status code = %d", res.StatusCode)
	}
	keyWord := "sessionid="
	mgr.sessionId = ""
	for _, cookie := range res.Cookies() {
		cookieStr := cookie.String()
		index := strings.Index(cookieStr, keyWord)
		if index == -1 {
			continue
		}
		cookieStr = cookieStr[index+len(keyWord):]
		index = strings.Index(cookieStr, ";")
		mgr.sessionId = cookieStr[:index]
	}
	if mgr.sessionId == "" {
		return fmt.Errorf("fail to get session id")
	}
	return nil
}

func (mgr *Core) getPasswordRSAPublicKey(rsaRes *pb.CAuthentication_GetPasswordRSAPublicKey_Response) error {
	pbReq := pb.CAuthentication_GetPasswordRSAPublicKey_Request{
		AccountName: mgr.loginInfo.UserName,
	}
	marshalData, err := proto.Marshal(&pbReq)
	if err != nil {
		return err
	}
	protoEncoded := base64.StdEncoding.EncodeToString(marshalData)

	reqUrl := fmt.Sprintf("%s/IAuthenticationService/GetPasswordRSAPublicKey/v1?input_protobuf_encoded=%s", kURI_STEAM_API, protoEncoded)
	httpReq, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Referer", kURI_STEAM_STROE+"/")
	httpReq.Header.Set("Origin", kURI_STEAM_STROE)

	res, err := mgr.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("fail to request GetPasswordRSAPublicKey, status code = %d", res.StatusCode)
	}

	err = proto.Unmarshal(data, rsaRes)
	if err != nil {
		return err
	}
	return nil
}

func (mgr *Core) beginAuthSessionViaCredentials(encryptedPassword string, rsaTimestamp uint64,
	beginAuthRes *pb.CAuthentication_BeginAuthSessionViaCredentials_Response) error {
	pbReq := pb.CAuthentication_BeginAuthSessionViaCredentials_Request{
		AccountName:         mgr.loginInfo.UserName,
		EncryptedPassword:   encryptedPassword,
		EncryptionTimestamp: rsaTimestamp,
		RememberLogin:       true,
		DeviceDetails: &pb.CAuthentication_DeviceDetails{
			DeviceFriendlyName: "Mozilla/5.0 (X11; Linux x86_64; rv:1.9.5.20) Gecko/2812-12-10 04:56:28 Firefox/3.8",
			PlatformType:       pb.EAuthTokenPlatformType_k_EAuthTokenPlatformType_MobileApp,
		},
		Persistence: pb.ESessionPersistence_k_ESessionPersistence_Persistent,
		WebsiteId:   "Community",
		Language:    6,
	}

	marshalData, err := proto.Marshal(&pbReq)
	if err != nil {
		return err
	}
	protoEncoded := base64.StdEncoding.EncodeToString(marshalData)
	reqUrl := fmt.Sprintf("%s/IAuthenticationService/BeginAuthSessionViaCredentials/v1", kURI_STEAM_API)

	res, err := mgr.loginAuthPost(reqUrl, protoEncoded)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("fail to request BeginAuthSessionViaCredentials, status code = %d", res.StatusCode)
	}

	xEresult, err := strconv.Atoi(res.Header.Get("X-Eresult"))
	if err != nil {
		return err
	}
	if xEresult != 1 {
		if xEresult == 5 {
			return fmt.Errorf("fail to login, error: Invaild username/password")
		} else {
			return fmt.Errorf("fail to login, X-Eresult: %d", xEresult)
		}
	}

	err = proto.Unmarshal(data, beginAuthRes)
	if err != nil {
		return err
	}
	if beginAuthRes.AllowedConfirmations == nil {
		return fmt.Errorf("fail to login, AllowedConfirmations is nil")
	}
	return nil
}

func (mgr *Core) updateAuthSessionWithSteamGuardCode(clientId, steamId uint64, guardType pb.EAuthSessionGuardType,
	updateAuthRes *pb.CAuthentication_UpdateAuthSessionWithSteamGuardCode_Response) error {
	log.Errorln("clientId =", clientId)
	log.Errorln("SteamId =", steamId)
	log.Errorln("GuardType =", guardType)

	code := ""
	if guardType == pb.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode ||
		guardType == pb.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceConfirmation {
		guardType = pb.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode
		log.Info("Please input 2FA(Two-Factor Authentication) code:")
		fmt.Scanf("%s", &code)
		code = strings.ToUpper(code)
		log.Infof("The input 2FA code is: %s", code)
	} else if guardType == pb.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode ||
		guardType == pb.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailConfirmation {
		guardType = pb.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode
		log.Info("Please input E-mail verification code:")
		fmt.Scanf("%s", &code)
		code = strings.ToUpper(code)
		log.Infof("The input E-mail verification code is: %s", code)
	} else {
		return fmt.Errorf("fail, guardType = %d", guardType)
	}
	pbReq := pb.CAuthentication_UpdateAuthSessionWithSteamGuardCode_Request{
		ClientId: clientId,
		SteamId:  steamId,
		Code:     code,
		CodeType: guardType,
	}

	marshalData, err := proto.Marshal(&pbReq)
	if err != nil {
		return err
	}
	protoEncoded := base64.StdEncoding.EncodeToString(marshalData)
	reqUrl := fmt.Sprintf("%s/IAuthenticationService/UpdateAuthSessionWithSteamGuardCode/v1", kURI_STEAM_API)
	res, err := mgr.loginAuthPost(reqUrl, protoEncoded)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("fail to request PollAuthSessionStatus, status code = %d", res.StatusCode)
	}
	log.Error("updateAuthReq header:", res.Header)
	return proto.Unmarshal(data, updateAuthRes)
}

func (mgr *Core) pollAuthSessionStatus(clientId uint64, requestId []byte,
	pollAuthRes *pb.CAuthentication_PollAuthSessionStatus_Response) error {
	pbReq := pb.CAuthentication_PollAuthSessionStatus_Request{
		ClientId:  clientId,
		RequestId: requestId,
	}

	marshalData, err := proto.Marshal(&pbReq)
	if err != nil {
		return err
	}
	protoEncoded := base64.StdEncoding.EncodeToString(marshalData)
	reqUrl := fmt.Sprintf("%s/IAuthenticationService/PollAuthSessionStatus/v1", kURI_STEAM_API)
	res, err := mgr.loginAuthPost(reqUrl, protoEncoded)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("fail to request PollAuthSessionStatus, status code = %d", res.StatusCode)
	}

	xEresult, err := strconv.Atoi(res.Header.Get("X-Eresult"))
	if err != nil {
		return err
	}
	if xEresult != 1 {
		return fmt.Errorf("fail to login, X-Eresult: %d", xEresult)
	}
	return proto.Unmarshal(data, pollAuthRes)
}

func (mgr *Core) finalizeLogin(refreshToken string) error {
	reqBody := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(reqBody)
	multipartWriter.WriteField("nonce", refreshToken)
	multipartWriter.WriteField("sessionid", mgr.sessionId)
	multipartWriter.WriteField("redir", fmt.Sprintf("%s/login/home/?goto=", kURI_STEAM_STROE))
	multipartWriter.Close()

	reqUrl := "https://login.steampowered.com/jwt/finalizelogin"
	httpReq, err := http.NewRequest("POST", reqUrl, reqBody)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Referer", kURI_STEAM_STROE+"/")
	httpReq.Header.Set("Origin", kURI_STEAM_STROE)
	httpReq.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	res, err := mgr.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("fail to post finalizeLogin, status code = %d", res.StatusCode)
	}
	fmt.Println(string(data))
	// return proto.Unmarshal(data, pollAuthRes)

	return nil
}

func (mgr *Core) encryptPassword(publicKeyMod, publicKeyExp string) (string, error) {
	modules, ret := new(big.Int).SetString(publicKeyMod, 16)
	if !ret {
		return "", fmt.Errorf("fail to generate publicKeyMod, type = big.Int, publicKeyMod = %s", publicKeyMod)
	}
	exp, err := strconv.ParseInt(publicKeyExp, 16, 32)
	if err != nil {
		return "", err
	}

	publicKey := rsa.PublicKey{}
	publicKey.N = modules
	publicKey.E = int(exp)

	passwordData := []byte(mgr.loginInfo.Password)
	encryptedPassword, err := rsa.EncryptPKCS1v15(rand.Reader, &publicKey, passwordData)
	if err != nil {
		return "", err
	}
	encodedPassword := base64.StdEncoding.EncodeToString(encryptedPassword)
	return encodedPassword, nil
}

func (mgr *Core) loginAuthPost(reqUrl, postData string) (*http.Response, error) {
	reqBody := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(reqBody)
	multipartWriter.WriteField("input_protobuf_encoded", postData)
	multipartWriter.Close()

	httpReq, err := http.NewRequest("POST", reqUrl, reqBody)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Referer", kURI_STEAM_STROE+"/")
	httpReq.Header.Set("Origin", kURI_STEAM_STROE)
	httpReq.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	return mgr.httpClient.Do(httpReq)
}
