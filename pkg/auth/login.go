package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	errcode "steam/pkg/err"
	pb "steam/pkg/proto"
	"steam/pkg/utils"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/proto"
)

func (core *Core) Login() error {
	log.Info("Connecting to steam server...")
	// Get RSA public key by proto message
	rsaRes := pb.CAuthentication_GetPasswordRSAPublicKey_Response{}
	err := core.getPasswordRSAPublicKey(&rsaRes)
	if err != nil {
		return err
	}
	encryptedPassword, err := core.encryptPassword(rsaRes.PublickeyMod, rsaRes.PublickeyExp)
	if err != nil {
		return err
	}

	time.Sleep(time.Millisecond * time.Duration(utils.RandRange(120, 300)))

	log.Infof("Try login as user: %s...", core.loginInfo.UserName)
	// Try begin auth
	beginAuthRes := pb.CAuthentication_BeginAuthSessionViaCredentials_Response{}
	err = core.beginAuthSessionViaCredentials(encryptedPassword, rsaRes.Timestamp,
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
		core.updateAuthSessionWithSteamGuardCode(beginAuthRes.ClientId, beginAuthRes.SteamId, confirmationType,
			&updateAuthRes)
	}

	log.Info("Logging in...")
	pollAuthRes := pb.CAuthentication_PollAuthSessionStatus_Response{}
	err = core.pollAuthSessionStatus(beginAuthRes.ClientId, beginAuthRes.RequestId, &pollAuthRes)
	if err != nil {
		return err
	}
	err = core.finalizeLogin(pollAuthRes.RefreshToken)
	if err != nil {
		return err
	}

	return nil
}

func (core *Core) getPasswordRSAPublicKey(rsaRes *pb.CAuthentication_GetPasswordRSAPublicKey_Response) error {
	pbReq := pb.CAuthentication_GetPasswordRSAPublicKey_Request{
		AccountName: core.loginInfo.UserName,
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

	res, err := core.httpClient.Do(httpReq)
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

func (core *Core) beginAuthSessionViaCredentials(encryptedPassword string, rsaTimestamp uint64,
	beginAuthRes *pb.CAuthentication_BeginAuthSessionViaCredentials_Response) error {
	pbReq := pb.CAuthentication_BeginAuthSessionViaCredentials_Request{
		DeviceFriendlyName:  "Galaxy S22",
		AccountName:         core.loginInfo.UserName,
		EncryptedPassword:   encryptedPassword,
		EncryptionTimestamp: rsaTimestamp,
		RememberLogin:       true,
		Persistence:         pb.ESessionPersistence_k_ESessionPersistence_Persistent,
		WebsiteId:           "Mobile",
		Language:            6,
		DeviceDetails: &pb.CAuthentication_DeviceDetails{
			DeviceFriendlyName: "Galaxy S22",
			PlatformType:       pb.EAuthTokenPlatformType_k_EAuthTokenPlatformType_MobileApp,
		},
	}

	marshalData, err := proto.Marshal(&pbReq)
	if err != nil {
		return err
	}
	protoEncoded := base64.StdEncoding.EncodeToString(marshalData)
	reqUrl := fmt.Sprintf("%s/IAuthenticationService/BeginAuthSessionViaCredentials/v1", kURI_STEAM_API)

	res, err := core.loginAuthPost(reqUrl, protoEncoded)
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
	err = errcode.CheckHeader(&res.Header)
	if err != nil {
		return err
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

func (core *Core) updateAuthSessionWithSteamGuardCode(clientId, steamId uint64, guardType pb.EAuthSessionGuardType,
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
	res, err := core.loginAuthPost(reqUrl, protoEncoded)
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

func (core *Core) pollAuthSessionStatus(clientId uint64, requestId []byte,
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
	res, err := core.loginAuthPost(reqUrl, protoEncoded)
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
	err = errcode.CheckHeader(&res.Header)
	if err != nil {
		return err
	}
	return proto.Unmarshal(data, pollAuthRes)
}

func (core *Core) finalizeLogin(refreshToken string) error {
	randomBytes := make([]byte, 12)
	if _, err := rand.Read(randomBytes); err != nil {
		return err
	}

	sessionID := make([]byte, hex.EncodedLen(len(randomBytes)))
	hex.Encode(sessionID, randomBytes)
	core.sessionId = string(sessionID)

	reqBody := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(reqBody)
	multipartWriter.WriteField("nonce", refreshToken)
	multipartWriter.WriteField("sessionid", core.sessionId)
	multipartWriter.WriteField("redir", fmt.Sprintf("%s/login/home/?goto=", kURI_STEAM_STROE))
	multipartWriter.Close()

	reqUrl := "https://login.steampowered.com/jwt/finalizelogin"
	httpReq, err := http.NewRequest("POST", reqUrl, reqBody)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	res, err := core.httpClient.Do(httpReq)
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
	// Add cookie
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	cookieList := []*http.Cookie{}
	for _, cookie := range res.Cookies() {
		if cookie.Name == "steamRefresh_steam" {
			cookieList = append(cookieList, cookie)
			break
		}
	}
	jar.SetCookies(
		&url.URL{
			Scheme: "https",
			Host:   "login.steampowered.com",
		},
		cookieList,
	)

	// Get loginSecure
	jsonData := string(data)
	steamId := gjson.Get(jsonData, "steamID").String()
	if steamId == "" {
		return fmt.Errorf("fail to get steam Id, response data: %s", jsonData)
	}
	nonce := ""
	auth := ""
	for _, tokenData := range gjson.Get(jsonData, "transfer_info").Array() {
		reqUrl := tokenData.Get("url").String()
		if reqUrl != "https://steamcommunity.com/login/settoken" {
			continue
		}
		params := tokenData.Get("params")
		nonce = params.Get("nonce").String()
		auth = params.Get("auth").String()
		break
	}
	reqBody = new(bytes.Buffer)
	multipartWriter = multipart.NewWriter(reqBody)
	multipartWriter.WriteField("nonce", nonce)
	multipartWriter.WriteField("auth", auth)
	multipartWriter.WriteField("steamID", steamId)
	multipartWriter.Close()

	httpReq, err = http.NewRequest("POST", reqUrl, reqBody)
	if err != nil {
		return err
	}
	httpReq.AddCookie(&http.Cookie{
		Name:  "sessionid",
		Value: core.sessionId,
	})
	httpReq.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	res, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}

	cookieList = []*http.Cookie{}
	for _, cookie := range res.Cookies() {
		if cookie.Name != "steamLoginSecure" {
			continue
		}
		cookieList = append(cookieList, cookie)
		break
	}
	cookieList = append(cookieList, &http.Cookie{
		Name:     "sessionid",
		Value:    core.sessionId,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
		Path:     "/"},
	)
	jar.SetCookies(
		&url.URL{
			Scheme: "https",
			Host:   "steamcommunity.com",
		},
		cookieList,
	)

	core.SteamId, err = strconv.ParseInt(steamId, 10, 64)
	if err != nil {
		return err
	}

	cookieList = []*http.Cookie{
		{Name: "mobileClientVersion", Value: "0 (2.1.3)"},
		{Name: "mobileClient", Value: "android"},
		{Name: "steamid", Value: steamId},
		{Name: "Steam_Language", Value: "english"},
		{Name: "dob", Value: ""},
	}
	jar.SetCookies(
		&url.URL{
			Scheme: "https",
			Host:   "steamcommunity.com",
		},
		cookieList,
	)
	core.httpClient.Jar = jar
	return nil
}

func (core *Core) encryptPassword(publicKeyMod, publicKeyExp string) (string, error) {
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

	passwordData := []byte(core.loginInfo.Password)
	encryptedPassword, err := rsa.EncryptPKCS1v15(rand.Reader, &publicKey, passwordData)
	if err != nil {
		return "", err
	}
	encodedPassword := base64.StdEncoding.EncodeToString(encryptedPassword)
	return encodedPassword, nil
}

func (core *Core) loginAuthPost(reqUrl, postData string) (*http.Response, error) {
	reqBody := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(reqBody)
	multipartWriter.WriteField("input_protobuf_encoded", postData)
	multipartWriter.Close()

	httpReq, err := http.NewRequest("POST", reqUrl, reqBody)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	return core.httpClient.Do(httpReq)
}
