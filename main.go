package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	// cloudbuild "cloud.google.com/go/cloudbuild/apiv1/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/manifoldco/promptui"
	"github.com/mitchellh/go-homedir"
	"github.com/securisec/cliam/gcp"
	"github.com/securisec/cliam/shared"
	"github.com/sirupsen/logrus"
	// cloudbuildpb "google.golang.org/genproto/googleapis/devtools/cloudbuild/v1"
)

var (
	gcpServiceAccountPath string
	gcpProjectId          string
	gcpRegion             string
	gcpZone               string
	gcpAccessToken        string
)

func main() {

	showIDs := false
	startFolder := ""
	if len(os.Args) > 1 {
		startFolder = os.Args[1]
		if strings.ToLower(startFolder) == "true" {
			startFolder = ""
			showIDs = true
		}
		if len(os.Args) > 2 {
			showIDs = true
		}
	}

	// sa, projectId, _, _ := getSaAndRegion()
	sa, _, _, _ := getSaAndRegion()

	ctx := context.Background()
	accessToken, err := gcp.GetAccessToken(ctx, sa)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get access token")
	}

	var httpReq *http.Request

	// url := "https://cloudresourcemanager.googleapis.com/v2/folders?parent=organizations/266328200403"
	// url := "https://cloudresourcemanager.googleapis.com/v2/folders?parent=folders/689878017259"
	url := "https://cloudresourcemanager.googleapis.com/v1beta1/organizations"
	isGet := true
	reqMethod := "GET"
	reqBody := ""

	if isGet {
		httpReq, err = http.NewRequestWithContext(ctx, reqMethod, url, nil)
		if err != nil {
			logrus.WithError(err).Error("Failed to set request")
		}
	} else {
		o, err := json.Marshal(reqBody)
		if err != nil {
			logrus.WithError(err).Error("Failed to marshall REQ body")
		}
		httpReq, err = http.NewRequestWithContext(context.TODO(), reqMethod, url, bytes.NewBuffer(o))
		if err != nil {
			logrus.WithError(err).Error("Failed to set request")
		}
	}
	if !isGet {
		httpReq.Header.Add("Content-Type", "application/json")
	}

	// add auth bearer token
	httpReq.Header.Add("Authorization", "Bearer "+accessToken)
	httpReq.Header.Add("user-agent", "google-cloud-sdk gcloud/379.0.0")

	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to call REST")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read REST data")
	}
	res.Body.Close()

	org := GcpOrganization{}
	if err := json.Unmarshal(body, &org); err != nil {
		logrus.WithError(err).Fatal("DEBUG: failed to marshall the JSON")
	}

	for _, v := range org.Organizations {
		fmt.Printf("%s\n", v.DisplayName)
		if len(startFolder) == 0 {
			getFolders("organizations", v.OrganizationID, accessToken, "  ", true, showIDs)
		} else {
			gf := getFolders("organizations", v.OrganizationID, accessToken, "  ", false, showIDs)
			for _, g := range gf.Folders {
				if g.DisplayName == startFolder {
					if showIDs {
						fmt.Printf("  %s (%s)\n", g.DisplayName, g.Name)
					} else {
						fmt.Printf("  %s\n", g.DisplayName)
					}
					getProjects("", g.Name, accessToken, "    ", showIDs)
					getFolders("", g.Name, accessToken, "    ", true, showIDs)
				}
			}
		}
	}
}

func getFolders(req string, id string, accessToken string, indent string, display bool, showIDs bool) GcpFolders {

	var httpReq *http.Request
	var err error
	var url string
	ctx := context.Background()

	// logrus.Warnf("Req: %s\n", req)
	if len(req) > 0 {
		url = fmt.Sprintf("https://cloudresourcemanager.googleapis.com/v2/folders?parent=%s/%s", req, id)
	} else {
		url = fmt.Sprintf("https://cloudresourcemanager.googleapis.com/v2/folders?parent=%s", id)
	}
	isGet := true
	reqMethod := "GET"
	reqBody := ""

	// logrus.Warn(url)
	if isGet {
		httpReq, err = http.NewRequestWithContext(ctx, reqMethod, url, nil)
		if err != nil {
			logrus.WithError(err).Error("Failed to set request")
		}
	} else {
		o, err := json.Marshal(reqBody)
		if err != nil {
			logrus.WithError(err).Error("Failed to marshall REQ body")
		}
		httpReq, err = http.NewRequestWithContext(context.TODO(), reqMethod, url, bytes.NewBuffer(o))
		if err != nil {
			logrus.WithError(err).Error("Failed to set request")
		}
	}
	if !isGet {
		httpReq.Header.Add("Content-Type", "application/json")
	}

	// add auth bearer token
	httpReq.Header.Add("Authorization", "Bearer "+accessToken)
	httpReq.Header.Add("user-agent", "google-cloud-sdk gcloud/379.0.0")

	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to call REST")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read REST data")
	}
	res.Body.Close()

	// logrus.Warn(string(body[:]))

	gf := GcpFolders{}
	if err := json.Unmarshal(body, &gf); err != nil {
		logrus.WithError(err).Fatal("DEBUG: failed to marshall the JSON")
		return GcpFolders{}
	}
	if display {
		for _, g := range gf.Folders {
			if showIDs {
				fmt.Printf("%s%s (%s)\n", indent, g.DisplayName, g.Name)
			} else {
				fmt.Printf("%s%s\n", indent, g.DisplayName)
			}
			getProjects("", g.Name, accessToken, indent, showIDs)
			// logrus.Warnf("Getting Folders Under: %s", g.Name)
			getFolders("", g.Name, accessToken, fmt.Sprintf("%s  ", indent), true, showIDs)
		}
	}

	return gf
}

func getProjects(req string, id string, accessToken string, indent string, showIDs bool) GcpProjects {

	var httpReq *http.Request
	var err error
	var url string
	ctx := context.Background()

	// logrus.Warnf("Req: %s\n", req)
	if len(req) > 0 {
		url = fmt.Sprintf("https://cloudresourcemanager.googleapis.com/v3/projects?parent=%s/%s", req, id)
	} else {
		url = fmt.Sprintf("https://cloudresourcemanager.googleapis.com/v3/projects?parent=%s", id)
	}
	isGet := true
	reqMethod := "GET"
	reqBody := ""

	// logrus.Warn(url)
	if isGet {
		httpReq, err = http.NewRequestWithContext(ctx, reqMethod, url, nil)
		if err != nil {
			logrus.WithError(err).Error("Failed to set request")
		}
	} else {
		o, err := json.Marshal(reqBody)
		if err != nil {
			logrus.WithError(err).Error("Failed to marshall REQ body")
		}
		httpReq, err = http.NewRequestWithContext(context.TODO(), reqMethod, url, bytes.NewBuffer(o))
		if err != nil {
			logrus.WithError(err).Error("Failed to set request")
		}
	}
	if !isGet {
		httpReq.Header.Add("Content-Type", "application/json")
	}

	// add auth bearer token
	httpReq.Header.Add("Authorization", "Bearer "+accessToken)
	httpReq.Header.Add("user-agent", "google-cloud-sdk gcloud/379.0.0")

	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to call REST")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read REST data")
	}
	res.Body.Close()

	// logrus.Warn(string(body[:]))

	prj := GcpProjects{}
	if err := json.Unmarshal(body, &prj); err != nil {
		logrus.WithError(err).Fatal("DEBUG: failed to marshall the JSON")
		return GcpProjects{}
	}
	for _, g := range prj.Projects {
		if showIDs {
			fmt.Printf("%s- %s (%s)\n", indent, g.DisplayName, g.Name)
		} else {
			fmt.Printf("%s- %s\n", indent, g.DisplayName)
		}
	}

	return prj
}

func getSaAndRegion() (string, string, string, string) {
	return getSaPath(), getProjectId(), gcpRegion, gcpZone
}

func getSaPath() string {
	if gcpAccessToken != "" {
		return ""
	}
	if gcpServiceAccountPath != "" {
		return expandPath(gcpServiceAccountPath)
	}
	if k, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
		return k
	}

	ca := getCurrentConfig("account")
	if ca != "" {
		cap := getCredPath(ca)
		if cap != "" {
			return expandPath(cap)
		}
	}
	return expandPath(promptInput("GCP service account path: "))
}

func getProjectId() string {
	if gcpProjectId != "" {
		return gcpProjectId
	}
	if k, ok := os.LookupEnv("CLOUDSDK_CORE_PROJECT"); ok {
		return k
	}
	cp := getCurrentConfig("project")
	if cp != "" {
		return cp
	}
	return promptInput("GCP project id: ")
}

func expandPath(p string) string {
	h, err := homedir.Expand(p)
	if err != nil {
		logrus.WithError(err).Fatal("failed to expand path")
	}
	ps, err := filepath.Abs(h)
	if err != nil {
		logrus.WithError(err).Fatal("failed to get absolute path")
	}
	return ps
}

func printValidArgs(f func() []string) {
	fmt.Println(shared.Red("Valid arguments:"))
	for _, v := range f() {
		fmt.Println(v)
	}
}

func getRequest(url, service string) (int, error) {
	logrus.Debug("url", url)
	res, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	status := res.StatusCode
	if status == 200 {
		logrus.Info("success", service, "url", url)
		return status, nil
	} else {
		logrus.Debug("failed", status)
	}
	return status, fmt.Errorf("bad status: %d", status)
}

func templateBuilder(t string, args map[string]string) (string, error) {
	temp := template.Must(template.New("").Parse(t))
	var tpl bytes.Buffer
	err := temp.Execute(&tpl, args)
	return tpl.String(), err
}

func ValidateJwtExpiration(token string) (isValid bool) {
	j, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		logrus.WithError(err).Error("Failed to parse JWT")
	}
	// if exp field is not set, we want to return true
	return j.Claims.(jwt.MapClaims).VerifyExpiresAt(time.Now().Unix(), false)
}

// promptInput is a helper function to prompt the user for input.
func promptInput(msg string) string {
	prompt := promptui.Prompt{
		Label: msg,
	}
	p, err := prompt.Run()
	if err != nil {
		logrus.WithError(err).Error("")
		os.Exit(1)
	}
	// trim whitespace
	return strings.Trim(p, " ")
}

func getCurrentConfig(setting string) string {
	// cat ~/.config/gcloud/configurations/config_default
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	fi := fmt.Sprintf("%s/.config/gcloud/configurations/config_default", home)
	f, err := os.Open(fi)
	if err != nil {
		fmt.Printf("error opening file: %v\n", err)
		os.Exit(1)
	}
	r := bufio.NewReader(f)
	s, e := Readln(r)
	for e == nil {
		if strings.HasPrefix(s, fmt.Sprintf("%s = ", setting)) {
			return strings.TrimSpace(strings.Split(s, "=")[1])
		}
		s, e = Readln(r)
	}
	return ""
}

func getCredPath(account string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	// /Users/cmaahs/.config/gcloud/legacy_credentials/chris.maahs@evolutioniq.com/adc.json
	fileName := fmt.Sprintf("%s/.config/gcloud/legacy_credentials/%s/adc.json", home, account)
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			return ""
		}
	}
	return fileName
}

// Readln returns a single line (without the ending \n)
// from the input buffered reader.
// An error is returned iff there is an error with the
// buffered reader.
func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}
