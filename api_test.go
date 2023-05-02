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
	files, err := suite.client.FileList(100, "VNRNKzrKLmyyGDe21AoJR1VAo1", "")
	suite.NoError(err)
	suite.NotNil(files)
	suite.NotEmpty(files.Files)
	for _, f := range files.Files {
		println(fmt.Sprintf("%s %s", f.Kind, f.ID))
	}
	suite.client.GetDownloadUrl("VNTyZBgvrOil_IUi2Hooqanuo1")
}

func (suite *TestPikpakSuite) TestOfflineDownload() {
	task, err := suite.client.OfflineDownload("test", "magnet:?xt=urn:btih:B73ACC9D2266B2C134D4FE9FA913E54B5FA1447F", "")
	suite.NoError(err)
	println(task.Task.ID)
	suite.NotEmpty(task.Task.ID)
	finishedTask, err := suite.client.WaitForOfflineDownloadComplete(task.Task.ID, time.Minute*1)
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
