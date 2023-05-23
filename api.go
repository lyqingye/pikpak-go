package pikpakgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	CLIENT_ID        = "YNxT9w7GMdWvEOKa"
	CLIENT_SECRET    = "dbw2OtmVEeuUvIptb1Coygx"
	DEVICE_ID        = "ed84f07adf1a442998d360933d0a080c"
	PIKPAK_USER_HOST = "https://user.mypikpak.com"
	PIKPAK_API_HOST  = "https://api-drive.mypikpak.com"
)

type PikPakClient struct {
	username     string
	password     string
	accessToken  string
	refreshToken string
	client       *resty.Client
}

func NewPikPakClient(username, password string) (*PikPakClient, error) {
	client := resty.New()
	client.EnableTrace()
	client.SetRetryCount(5)
	client.SetRetryWaitTime(5 * time.Second)
	client.SetRetryMaxWaitTime(60 * time.Second)
	client.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36")

	pikpak := PikPakClient{
		username: username,
		password: password,
		client:   client,
	}

	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		if strings.Contains(string(r.Body()), "unauthenticated") {
			return pikpak.Login() != nil
		}
		if err == nil {
			return false
		}
		if err != nil {
			return true
		}
		return false
	})

	return &pikpak, nil
}

func (c *PikPakClient) Login() error {
	req := RequestLogin{
		ClientId:     CLIENT_ID,
		ClientSecret: CLIENT_SECRET,
		Username:     c.username,
		Password:     c.password,
	}
	resp := ResponseLogin{}
	originResp, err := c.client.R().
		SetBody(&req).
		SetResult(&resp).
		Post(fmt.Sprintf("%s/v1/auth/signin", PIKPAK_USER_HOST))
	if err != nil {
		return err
	}
	if resp.AccessToken == "" {
		return errRespToError(originResp.Body())
	}
	c.accessToken = resp.AccessToken
	c.refreshToken = resp.RefreshToken
	return nil
}

func (c *PikPakClient) Logout() error {
	req := RequestLogout{
		Token: c.accessToken,
	}
	_, err := c.client.R().
		SetBody(&req).
		Post(fmt.Sprintf("%s/v1/auth/revoke", PIKPAK_USER_HOST))
	return err
}

func (c *PikPakClient) FileList(limit int, parentId string, nextPageToken string) (*FileList, error) {
	filters := Filters{
		Phase: map[string]string{
			"eq": PhaseTypeComplete,
		},
		Trashed: map[string]bool{
			"eq": false,
		},
	}
	filtersBz, err := json.Marshal(&filters)
	if err != nil {
		return nil, err
	}
	req := RequestFileList{
		ParentId:      parentId,
		ThumbnailSize: ThumbnailSizeM,
		Limit:         strconv.Itoa(limit),
		WithAudit:     strconv.FormatBool(true),
		NextPageToken: nextPageToken,
		Filters:       string(filtersBz),
	}
	bz, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	reqParams := make(map[string]string)
	err = json.Unmarshal(bz, &reqParams)
	if err != nil {
		return nil, err
	}

	result := FileList{}
	_, err = c.client.R().
		SetQueryParams(reqParams).
		SetResult(&result).
		SetAuthToken(c.accessToken).
		Get(fmt.Sprintf("%s/drive/v1/files", PIKPAK_API_HOST))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *PikPakClient) GetFile(id string) (*File, error) {
	file := File{}
	_, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetResult(&file).
		Get(fmt.Sprintf("%s/drive/v1/files/%s?usage=FETCH", PIKPAK_API_HOST, id))
	if err != nil {
		return nil, err
	}
	return &file, err
}

func (c *PikPakClient) GetDownloadUrl(id string) (string, error) {
	file := File{}
	_, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetResult(&file).
		Get(fmt.Sprintf("%s/drive/v1/files/%s?usage=FETCH", PIKPAK_API_HOST, id))
	if err != nil {
		return "", err
	}
	return file.WebContentLink, err
}

func (c *PikPakClient) OfflineDownload(name string, fileUrl string, parentId string) (*NewTask, error) {
	folderType := ""
	if parentId != "" {
		folderType = FolderTypeDownload
	}
	req := RequestNewTask{
		Kind:       KindOfFile,
		Name:       name,
		ParentID:   parentId,
		UploadType: UploadTypeURL,
		URL: &URL{
			URL: fileUrl,
		},
		FolderType: folderType,
	}
	task := NewTask{}
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetResult(&task).
		SetBody(&req).
		Post(fmt.Sprintf("%s/drive/v1/files", PIKPAK_API_HOST))
	if err != nil {
		return nil, err
	}
	return &task, errRespToError(resp.Body())
}

func (c *PikPakClient) OfflineList(limit int, nextPageToken string) (*TaskList, error) {
	filters := Filters{
		Phase: map[string]string{
			"in": strings.Join([]string{PhaseTypeRunning, PhaseTypeComplete, PhaseTypeError}, ","),
		},
	}
	filtersBz, err := json.Marshal(&filters)
	if err != nil {
		return nil, err
	}
	req := RequestTaskList{
		ThumbnailSize: ThumbnailSizeS,
		Limit:         strconv.Itoa(limit),
		NextPageToken: nextPageToken,
		Filters:       string(filtersBz),
		FileType:      FileTypeOffline,
	}
	bz, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	reqParams := make(map[string]string)
	err = json.Unmarshal(bz, &reqParams)
	if err != nil {
		return nil, err
	}

	result := TaskList{}
	resp, err := c.client.R().
		SetQueryParams(reqParams).
		SetResult(&result).
		SetAuthToken(c.accessToken).
		Get(fmt.Sprintf("%s/drive/v1/tasks", PIKPAK_API_HOST))
	if err != nil {
		return nil, err
	}
	return &result, errRespToError(resp.Body())
}

func (c *PikPakClient) OfflineRetry(taskId string) error {
	req := RequestTaskRetry{
		Id:         taskId,
		Type:       FileTypeOffline,
		CreateType: CreateTypeRetry,
	}
	bz, err := json.Marshal(&req)
	if err != nil {
		return err
	}
	reqParams := make(map[string]string)
	err = json.Unmarshal(bz, &reqParams)
	if err != nil {
		return err
	}
	resp, err := c.client.R().
		SetQueryParams(reqParams).
		SetAuthToken(c.accessToken).
		Get(fmt.Sprintf("%s/drive/v1/task", PIKPAK_API_HOST))
	if err != nil {
		return err
	}
	return errRespToError(resp.Body())
}

func (c *PikPakClient) OfflineListIterator(callback func(task *Task) bool) error {
	nextPageToken := ""
	pageSize := 100
Exit:
	for {
		taskList, err := c.OfflineList(pageSize, nextPageToken)
		if err != nil {
			return err
		}
		for _, task := range taskList.Tasks {
			if callback(task) {
				break Exit
			}
		}
		if len(taskList.Tasks) < pageSize {
			break Exit
		}
		nextPageToken = taskList.NextPageToken
	}
	return nil
}

func (c *PikPakClient) WaitForOfflineDownloadComplete(taskId string, timeout time.Duration, progressFn func(*Task)) (*Task, error) {
	finished := false
	var finishedTask *Task
	var lastErr error
	endTime := time.Now().Add(timeout)
	for {
		if finished {
			return finishedTask, nil
		}
		if time.Now().After(endTime) {
			if lastErr != nil {
				return nil, lastErr
			} else {
				return nil, errors.New("wait for offline download complete timeout")
			}
		}
		lastErr = c.OfflineListIterator(func(task *Task) bool {
			if task.ID == taskId {
				if progressFn != nil {
					progressFn(task)
				}
				if (task.Phase == PhaseTypeComplete && task.Progress == 100) || task.Phase == PhaseTypeError {
					finished = true
					finishedTask = task
					return true
				}
			}
			return false
		})
		time.Sleep(5 * time.Second)
	}
}

func (c *PikPakClient) BatchTrashFiles(ids []string) error {
	req := RequestBatch{
		Ids: ids,
	}
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetBody(&req).
		Post(fmt.Sprintf("%s/drive/v1/files:batchTrash", PIKPAK_API_HOST))
	if err != nil {
		return err
	}
	return errRespToError(resp.Body())
}

func (c *PikPakClient) BatchDeleteFiles(ids []string) error {
	req := RequestBatch{
		Ids: ids,
	}
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetBody(&req).
		Post(fmt.Sprintf("%s/drive/v1/files:batchDelete", PIKPAK_API_HOST))
	if err != nil {
		return err
	}
	return errRespToError(resp.Body())
}

func (c *PikPakClient) EmptyTrash() error {
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		Patch(fmt.Sprintf("%s/drive/v1/files/trash:empty", PIKPAK_API_HOST))
	if err != nil {
		return err
	}
	return errRespToError(resp.Body())
}

func errRespToError(body []byte) error {
	pikpakErr := Error{}
	err := json.Unmarshal(body, &pikpakErr)
	if err != nil {
		return nil
	} else if pikpakErr.Code != 0 && pikpakErr.Reason != "" {
		return errors.New(pikpakErr.Error())
	}
	return nil
}
