package api

import (
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const testArchiveName = "libsacloud_test_archive"

//func TestGetCentOSArchiveByName(t *testing.T) {
//	archiveAPI := client.Archive
//
//	res, err := archiveAPI.Reset().WithNameLike("CentOS 7.2 64bit").Find()
//	assert.NoError(t, err)
//	assert.NotEmpty(t, res)
//	assert.Equal(t, len(res.Archives), 1)
//}

func TestGetArchiveWithLimitOffset(t *testing.T) {
	archiveAPI := client.Archive
	res, err := archiveAPI.Reset().Limit(2).Offset(1).Include("Name").Include("CreatedAt").Find()
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.Equal(t, len(res.Archives), 2)
	assert.Equal(t, res.From, 1)
	assert.Equal(t, res.Count, 2)
	assert.True(t, res.Total > 2)
}

func TestFindState(t *testing.T) {
	api := client.Archive

	api.Reset().WithNameLike("hoge").FilterBy("Fuga", "fuga").Limit(10).Offset(1).Include("inc").Exclude("enc")

	state := api.state

	assert.NotEmpty(t, state)
	assert.Equal(t, state.Filter["Name"], "hoge")
	assert.Equal(t, state.Filter["Fuga"], "fuga")
	assert.Equal(t, state.Count, 10)
	assert.Equal(t, state.From, 1)
	assert.Equal(t, state.Include[0], "inc")
	assert.Equal(t, state.Exclude[0], "enc")

	//clear state
	api.Reset()
	state = api.state
	assert.Empty(t, state.Filter)
	assert.Empty(t, state.Count)
	assert.Empty(t, state.From)
	assert.Empty(t, state.Include)
	assert.Empty(t, state.Exclude)

	res, err := api.withNameLike("CentOS").limit(1).Find()

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.Equal(t, res.Count, 1)
	assert.Contains(t, res.Archives[0].Name, "CentOS")
}

func TestFindStateWithSetter(t *testing.T) {
	api := client.Archive

	// set parameters by setter method
	api.SetEmpty()
	api.SetNameLike("hoge")
	api.SetFilterBy("Fuga", "fuga")
	api.SetLimit(10)
	api.SetOffset(1)
	api.SetInclude("inc")
	api.SetExclude("enc")

	state := api.state

	assert.NotEmpty(t, state)
	assert.Equal(t, state.Filter["Name"], "hoge")
	assert.Equal(t, state.Filter["Fuga"], "fuga")
	assert.Equal(t, state.Count, 10)
	assert.Equal(t, state.From, 1)
	assert.Equal(t, state.Include[0], "inc")
	assert.Equal(t, state.Exclude[0], "enc")

	//clear state
	api.SetEmpty()
	state = api.state
	assert.Empty(t, state.Filter)
	assert.Empty(t, state.Count)
	assert.Empty(t, state.From)
	assert.Empty(t, state.Include)
	assert.Empty(t, state.Exclude)

	api.SetNameLike("CentOS")
	api.SetLimit(1)

	res, err := api.Find()

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.Equal(t, res.Count, 1)
	assert.Contains(t, res.Archives[0].Name, "CentOS")
}

func TestArchiveCRUDAndFTP(t *testing.T) {
	api := client.Archive

	// get icon ID
	icons, err := client.Icon.Reset().WithSharedScope().Include("ID").Find()

	if !assert.NoError(t, err) || !assert.True(t, len(icons.Icons) > 0) {
		return
	}
	icon := icons.Icons[0]

	//CREATE
	newArchive := api.New()
	newArchive.Name = testArchiveName
	newArchive.SetDescription("hoge")
	newArchive.AppendTag("hoge")
	newArchive.SetIcon(&icon)
	newArchive.SetSizeGB(20)

	archive, err := api.Create(newArchive)

	assert.NoError(t, err)
	assert.NotEmpty(t, archive)

	assert.Equal(t, archive.Description, "hoge")
	assert.Len(t, archive.Tags, 1)
	assert.Equal(t, archive.Tags[0], "hoge")
	assert.Equal(t, archive.Icon.ID, icon.ID)

	archiveID := archive.ID

	//READ
	archive, err = api.Read(archiveID)
	assert.NoError(t, err)
	assert.NotEmpty(t, archive)

	// Update
	archive.SetDescription("")
	archive.ClearTags()
	archive.ClearIcon()

	archive, err = api.Update(archive.ID, archive)

	assert.Equal(t, archive.Description, "")
	assert.Len(t, archive.Tags, 0)
	assert.Nil(t, archive.Icon)

	//Open
	ftpServer, err := api.OpenFTP(archive.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, ftpServer.Password)

	password := ftpServer.Password

	////Close
	//res, err := api.CloseFTP(archiveID)
	//assert.NoError(t, err)
	//assert.True(t, res)

	//Re-Open(password not changed)
	//ftpServer, err = api.OpenFTP(archive.ID, false)
	//assert.NoError(t, err)
	//assert.Equal(t, ftpServer.Password, password)

	//Close
	api.CloseFTP(archiveID)

	//Re-Open(will password change)
	ftpServer, err = api.OpenFTP(archive.ID)
	assert.NoError(t, err)
	assert.NotEqual(t, ftpServer.Password, password)

	//Delete
	_, err = api.Delete(archiveID)
	assert.NoError(t, err)

}

func TestCreateAndWait(t *testing.T) {

	archiveAPI := client.Archive
	src, err := archiveAPI.FindLatestStableCentOS()

	assert.NoError(t, err)
	id := src.ID
	assert.NotEmpty(t, id)

	//CREATE
	newArchive := archiveAPI.New()
	newArchive.Name = testArchiveName
	newArchive.Description = "hoge"
	newArchive.SetSourceArchive(id)

	archive, err := archiveAPI.Create(newArchive)

	assert.NoError(t, err)
	assert.NotEmpty(t, archive)

	err = archiveAPI.SleepWhileCopying(archive.ID, 180*time.Second)
	assert.NoError(t, err)

	archiveAPI.Delete(archive.ID)

}

func TestCreateAndAsyncWait(t *testing.T) {

	archiveAPI := client.Archive
	src, err := archiveAPI.FindLatestStableCentOS()

	assert.NoError(t, err)
	id := src.ID
	assert.NotEmpty(t, id)

	//CREATE
	newArchive := archiveAPI.New()
	newArchive.Name = testArchiveName
	newArchive.Description = "hoge"
	newArchive.SetSourceArchive(id)

	archive, err := archiveAPI.Create(newArchive)

	assert.NoError(t, err)
	assert.NotEmpty(t, archive)
	defer func() {
		archiveAPI.Delete(archive.ID)
	}()

	complete, progress, errChan := archiveAPI.AsyncSleepWhileCopying(archive.ID, client.DefaultTimeoutDuration)

	for {
		select {
		case a := <-progress:
			t.Logf("Copying...\t %d MB / %d MB", a.GetMigratedMB(), a.GetSizeMB())
		case a := <-complete:
			t.Logf("Done...\t %d MB / %d MB", a.GetMigratedMB(), a.GetSizeMB())
			//t.Logf("Trace:%#v", a)
			return
		case e := <-errChan:
			assert.Fail(t, e.Error(), nil)
			return
		case <-time.After(20 * time.Minute):
			assert.Fail(t, "Timeout: AsyncSleepWhileCopying: Disk -> %d", archive.ID)
			return
		}
	}
}

func TestArchiveAPI_FindStableOSs(t *testing.T) {

	api := client.Archive
	type target struct {
		label string
		f     func() (*sacloud.Archive, error)
	}

	targets := []target{
		{label: "CentOS", f: api.FindLatestStableCentOS},
		{label: "Debian", f: api.FindLatestStableDebian},
		{label: "Ubuntu", f: api.FindLatestStableUbuntu},
		{label: "VyOS", f: api.FindLatestStableVyOS},
		{label: "CoreOS", f: api.FindLatestStableCoreOS},
		{label: "Kusanagi", f: api.FindLatestStableKusanagi},
	}

	for _, ts := range targets {
		res, err := ts.f()
		assert.NoError(t, err)
		assert.NotNil(t, res)
		t.Logf("Zone:%s / Current Stable %s: %#v", client.Zone, ts.label, res.Resource)

	}

}

func init() {
	testSetupHandlers = append(testSetupHandlers, cleanupArchive)
	testTearDownHandlers = append(testTearDownHandlers, cleanupArchive)
}

func cleanupArchive() {
	items, _ := client.Archive.Reset().WithNameLike(testArchiveName).Find()
	for _, item := range items.Archives {
		client.Archive.Delete(item.ID)
	}
}
