package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	repoAddress string
	mountPoint  string
	key         string
	program     string
	mountCmd    *exec.Cmd
}

func (suite *TestSuite) SetupSuite() {
	suite.program = "bridge_guard"
	suite.mountPoint = "Z:"
	suite.key = "79dvjtK2jcPpfXi1HsKa2S9GV5qjhbKgJHQyoWevg6ZQ"
	tempDir := os.TempDir()
	suite.repoAddress = filepath.Join(tempDir, "bridge_guard_temp_mount")
	err := os.MkdirAll(suite.repoAddress, 0755)
	assert.NoError(suite.T(), err)
	suite.initRepo()
	suite.mountRepo()
}

func (suite *TestSuite) TearDownSuite() {
	suite.unmountRepo()
	err := os.RemoveAll(suite.repoAddress)
	assert.NoError(suite.T(), err)
}

func (suite *TestSuite) initRepo() {
	cmd := exec.Command(suite.program, "init", "-k", suite.key, "-p", suite.repoAddress)
	err := cmd.Run()
	assert.NoError(suite.T(), err)
}

func (suite *TestSuite) mountRepo() {
	suite.mountCmd = exec.Command(suite.program, "mount", "-k", suite.key, "-p", suite.repoAddress, "-m", suite.mountPoint)
	err := suite.mountCmd.Start()
	assert.NoError(suite.T(), err)
	time.Sleep(2 * time.Second)
}

func (suite *TestSuite) unmountRepo() {
	if suite.mountCmd != nil && suite.mountCmd.Process != nil {
		_ = suite.mountCmd.Process.Kill()
		_ = suite.mountCmd.Wait()
	}
}

func (suite *TestSuite) writeFile(fileName, content string) {
	filePath := filepath.Join(suite.mountPoint, fileName)
	err := os.WriteFile(filePath, []byte(content), 0644)
	assert.NoError(suite.T(), err)
}

func (suite *TestSuite) readFile(fileName string) string {
	filePath := filepath.Join(suite.mountPoint, fileName)
	content, err := os.ReadFile(filePath)
	assert.NoError(suite.T(), err)
	return string(content)
}

func (suite *TestSuite) writeDirectory(dirName string) {
	dirPath := filepath.Join(suite.mountPoint, dirName)
	err := os.MkdirAll(dirPath, 0755)
	assert.NoError(suite.T(), err)
}

func (suite *TestSuite) readDirectory(dirName string) []os.DirEntry {
	dirPath := filepath.Join(suite.mountPoint, dirName)
	entries, err := os.ReadDir(dirPath)
	assert.NoError(suite.T(), err)
	return entries
}

func (suite *TestSuite) TestWriteAndReadFile() {
	fileName := "test.txt"
	content := "Hello World!"

	suite.writeFile(fileName, content)
	readContent := suite.readFile(fileName)

	assert.Equal(suite.T(), content, readContent)
}

func (suite *TestSuite) TestWriteAndReadDirectory() {
	dirName := "test_dir"

	suite.writeDirectory(dirName)
	entries := suite.readDirectory(dirName)

	assert.Equal(suite.T(), 0, len(entries))
}

func TestBridgeGuardTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
