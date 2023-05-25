package pikpakgo_test

import (
	"fmt"
	"testing"
	"time"

	pikpakgo "github.com/lyqingye/pikpak-go"

	"github.com/stretchr/testify/suite"
)

type TestPikpakSuite struct {
	suite.Suite
	client *pikpakgo.PikPakClient
}

func TestPikpakAPI(t *testing.T) {
	suite.Run(t, new(TestPikpakSuite))
}

func (suite *TestPikpakSuite) SetupTest() {
	client, err := pikpakgo.NewPikPakClient("", "")
	suite.NoError(err)
	err = client.Login()
	suite.NoError(err)
	suite.client = client
}

func (suite *TestPikpakSuite) TestListFile() {
	files, err := suite.client.FileList(100, "", "")
	suite.NoError(err)
	suite.NotEmpty(files)
}

func (suite *TestPikpakSuite) TestGetDownloadUrl() {
	files, err := suite.client.FileList(100, "", "")
	suite.NoError(err)
	suite.NotNil(files)
	suite.NotEmpty(files.Files)
	for _, f := range files.Files {
		println(fmt.Sprintf("%s %s", f.Kind, f.ID))
	}
	url, err := suite.client.GetDownloadUrl("VNWHVrXMz4En2yoLs_x-Uf_Ko1")
	suite.NoError(err)
	println(url)
}

func (suite *TestPikpakSuite) TestOfflineDownload() {
	task, err := suite.client.OfflineDownload("test", "magnet:?xt=urn:btih:bce204d9d53d7c843856b0b17c1d5dc1478d1cd5&tr=http%3a%2f%2ft.nyaatracker.com%2fannounce&tr=http%3a%2f%2ftracker.kamigami.org%3a2710%2fannounce&tr=http%3a%2f%2fshare.camoe.cn%3a8080%2fannounce&tr=http%3a%2f%2fopentracker.acgnx.se%2fannounce&tr=http%3a%2f%2fanidex.moe%3a6969%2fannounce&tr=http%3a%2f%2ft.acg.rip%3a6699%2fannounce&tr=https%3a%2f%2ftr.bangumi.moe%3a9696%2fannounce&tr=udp%3a%2f%2ftr.bangumi.moe%3a6969%2fannounce&tr=http%3a%2f%2fopen.acgtracker.com%3a1096%2fannounce&tr=udp%3a%2f%2ftracker.opentrackr.org%3a1337%2fannounce", "")
	suite.NoError(err)
	println(task.Task.ID)
	suite.NotEmpty(task.Task.ID)
	finishedTask, err := suite.client.WaitForOfflineDownloadComplete(task.Task.ID, time.Minute*1, nil)
	suite.NoError(err)
	println(finishedTask)
	uri, err := suite.client.GetDownloadUrl(finishedTask.FileID)
	suite.NoError(err)
	println(uri)
	uri2, err := suite.client.GetDownloadUrl("VNUHa9_xcQJe5gcxb7tZOpefo1")
	suite.NoError(err)
	println(uri2)
	files, err := suite.client.FileList(100, "", "")
	suite.NoError(err)
	suite.NotNil(files)
}

func (suite *TestPikpakSuite) TestOfflineList() {
	tasks, err := suite.client.OfflineList(100, "")
	suite.NoError(err)
	for _, f := range tasks.Tasks {
		println(fmt.Sprintf("%s %s", f.Kind, f.ID))
	}
}

func (suite *TestPikpakSuite) TestEmptyTrash() {
	err := suite.client.EmptyTrash()
	suite.NoError(err)
}

func (suite *TestPikpakSuite) TestTaskRetry() {
	tasks, err := suite.client.OfflineList(100, "")
	suite.NoError(err)
	suite.NotNil(tasks)
	for _, task := range tasks.Tasks {
		err = suite.client.OfflineRetry(task.ID)
		suite.NoError(err)
	}
}

func (suite *TestPikpakSuite) TestBatchTrashFiles() {
	err := suite.client.BatchTrashFiles([]string{
		"VNV9ua9L2OQzryfULN72j50to1",
		"VNVDm8wqQjBlpj7t6p3E9wsMo1",
	})
	suite.NoError(err)
}

func (suite *TestPikpakSuite) TestBatchDeleteFiles() {
	err := suite.client.BatchDeleteFiles([]string{
		"VNV9ua9L2OQzryfULN72j50to1",
	})
	suite.NoError(err)
}

func (suite *TestPikpakSuite) TestGetAbout() {
	info, err := suite.client.About()
	suite.NoError(err)
	suite.NotNil(info)
}

func (suite *TestPikpakSuite) TestGetMe() {
	info, err := suite.client.Me()
	suite.NoError(err)
	suite.NotNil(info)
}

func (suite *TestPikpakSuite) TestGetInviteInfo() {
	info, err := suite.client.InviteInfo()
	suite.NoError(err)
	suite.NotNil(info)
}
