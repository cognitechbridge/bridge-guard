package e2e_test

import (
	"bytes"
	"context"
	"ctb-cli/cmd"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/rand"
)

type TestSuite struct {
	suite.Suite
	repoAddress string
	mountPoint  string
	key         string
	cancel      context.CancelFunc
}

func (suite *TestSuite) SetupSuite() {
	if runtime.GOOS == "linux" {
		suite.mountPoint = "/mnt/bridge_guard"
	} else {
		suite.mountPoint = "Z:\\"
	}
	suite.key = "79dvjtK2jcPpfXi1HsKa2S9GV5qjhbKgJHQyoWevg6ZQ"
	tempDir := os.TempDir()
	suite.repoAddress = filepath.Join(tempDir, "bridge_guard_temp_mount")
	os.RemoveAll(suite.repoAddress)
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

func (suite *TestSuite) exec(args ...string) (context.CancelFunc, string, error) {
	root := cmd.RootCmd
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs(args)
	ctx, cancel := context.WithCancel(context.Background())
	_, err := root.ExecuteContextC(ctx)
	return cancel, buf.String(), err
}

func (suite *TestSuite) initRepo() {
	_, _, err := suite.exec("init", "-k", suite.key, "-p", suite.repoAddress)
	assert.NoError(suite.T(), err)
}

func removeTrailingBackslash(path string) string {
	if runtime.GOOS == "windows" && strings.HasSuffix(path, "\\") {
		path = strings.TrimSuffix(path, "\\")
	}
	return path
}

func (suite *TestSuite) mountRepo() {
	go func() {
		suite.cancel, _, _ = suite.exec("mount", "-k", suite.key, "-p", suite.repoAddress, "-m", removeTrailingBackslash((suite.mountPoint)))
	}()
	time.Sleep(2 * time.Second)
}

func (suite *TestSuite) unmountRepo() {
	if suite.cancel != nil {
		suite.cancel()
	}
}

func (suite *TestSuite) writeFile(fileName string, content []byte) {
	filePath := filepath.Join(suite.mountPoint, fileName)
	err := os.WriteFile(filePath, content, 0644)
	assert.NoError(suite.T(), err)
}

func (suite *TestSuite) appendToFile(fileName string, content []byte) {
	filePath := filepath.Join(suite.mountPoint, fileName)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	assert.NoError(suite.T(), err)
	defer file.Close()
	_, err = file.Write(content)
	assert.NoError(suite.T(), err)
}

func (suite *TestSuite) readFile(fileName string) []byte {
	filePath := filepath.Join(suite.mountPoint, fileName)
	content, err := os.ReadFile(filePath)
	assert.NoError(suite.T(), err)
	return content
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

func (suite *TestSuite) moveFile(srcDir, destDir, fileName string) {
	srcPath := filepath.Join(suite.mountPoint, srcDir, fileName)
	destPath := filepath.Join(suite.mountPoint, destDir, fileName)
	err := os.Rename(srcPath, destPath)
	assert.NoError(suite.T(), err)
}

func (suite *TestSuite) TestWriteAndReadFile() {
	fileName := "test.txt"
	content := "Hello World!"

	suite.writeFile(fileName, []byte(content))
	readContent := suite.readFile(fileName)

	filePath := filepath.Join(suite.mountPoint, fileName)
	fileInfo, err := os.Stat(filePath)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(len(content)), fileInfo.Size())

	assert.Equal(suite.T(), content, string(readContent))
}

func (suite *TestSuite) TestWriteAndReadDirectory() {
	dirName := "test_dir"

	suite.writeDirectory(dirName)
	entries := suite.readDirectory(dirName)

	assert.Equal(suite.T(), 0, len(entries))
}

func (suite *TestSuite) TestMoveFile() {
	srcDir := "src_dir"
	destDir := "dest_dir"
	fileName := "test.txt"
	content := "Hello World!"

	suite.writeDirectory(srcDir)
	suite.writeDirectory(destDir)
	suite.writeFile(filepath.Join(srcDir, fileName), []byte(content))

	suite.moveFile(srcDir, destDir, fileName)

	srcEntries := suite.readDirectory(srcDir)
	destEntries := suite.readDirectory(destDir)
	readContent := suite.readFile(filepath.Join(destDir, fileName))

	assert.Equal(suite.T(), 0, len(srcEntries))
	assert.Equal(suite.T(), 1, len(destEntries))
	assert.Equal(suite.T(), content, string(readContent))
}

func (suite *TestSuite) TestMoveFileSameDir() {
	dir := "src_dir"
	fileName := "test.txt"
	content := "Hello World!"

	suite.writeDirectory(dir)
	suite.writeFile(filepath.Join(dir, fileName), []byte(content))

	suite.moveFile(dir, dir, fileName)

	entries := suite.readDirectory(dir)
	readContent := suite.readFile(filepath.Join(dir, fileName))

	assert.Equal(suite.T(), 1, len(entries))
	assert.Equal(suite.T(), content, string(readContent))
}

func (suite *TestSuite) TestWriteAndReadLargeFile() {
	fileName := "large_test.txt"
	content := make([]byte, 100*1024*1024) // 100 MB of data
	for i := range content {
		content[i] = byte(i % 256)
	}

	suite.writeFile(fileName, content)
	readContent := suite.readFile(fileName)

	assert.Equal(suite.T(), len(content), len(readContent))
	assert.Equal(suite.T(), content, readContent)
}

func (suite *TestSuite) TestWriteEmptyThenAppendFile() {
	fileName := "empty_then_append.txt"
	appendContent := "This is some appended text."

	// Write an empty file
	suite.writeFile(fileName, []byte{})
	filePath := filepath.Join(suite.mountPoint, fileName)
	fileInfo, err := os.Stat(filePath)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), fileInfo.Size())

	// Append content to the file
	suite.appendToFile(fileName, []byte(appendContent))

	// Read the file and verify the contents
	expectedContent := appendContent
	readContent := suite.readFile(fileName)

	assert.Equal(suite.T(), expectedContent, string(readContent))
}

func (suite *TestSuite) TestWriteThenAppendFile() {
	fileName := "empty_then_append.txt"
	initialContent := "This is some initial text."
	appendContent := "This is some appended text."

	// Write an empty file
	suite.writeFile(fileName, []byte(initialContent))
	filePath := filepath.Join(suite.mountPoint, fileName)
	fileInfo, err := os.Stat(filePath)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(len(initialContent)), fileInfo.Size())

	// Append content to the file
	suite.appendToFile(fileName, []byte(appendContent))

	// Read the file and verify the contents
	expectedContent := initialContent + appendContent
	readContent := suite.readFile(fileName)

	assert.Equal(suite.T(), expectedContent, string(readContent))
}

// Generates a random file name
func randomFileName() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b) + ".txt"
}

// Monkey test implementation with assertions
func (suite *TestSuite) TestMonkey() {
	rand.Seed(uint64(time.Now().UnixNano()))

	operations := []func(){
		func() {
			fileName := randomFileName()
			content := []byte("Monkey testing!")
			suite.writeFile(fileName, content)
			readContent := suite.readFile(fileName)
			assert.Equal(suite.T(), content, readContent)
		},
		func() {
			fileName := randomFileName()
			initialContent := []byte("Initial content.")
			appendContent := []byte(" Appending more data.")
			suite.writeFile(fileName, initialContent)
			suite.appendToFile(fileName, appendContent)
			expectedContent := append(initialContent, appendContent...)
			readContent := suite.readFile(fileName)
			assert.Equal(suite.T(), expectedContent, readContent)
		},
		func() {
			fileName := randomFileName()
			content := []byte("Reading file content.")
			suite.writeFile(fileName, content)
			readContent := suite.readFile(fileName)
			assert.Equal(suite.T(), content, readContent)
		},
		func() {
			dirName := "monkey_dir_" + randomFileName()
			suite.writeDirectory(dirName)
			entries := suite.readDirectory(dirName)
			assert.Equal(suite.T(), 0, len(entries))
		},
		func() {
			srcDir := "src_dir_" + randomFileName()
			destDir := "dest_dir_" + randomFileName()
			fileName := randomFileName()
			content := []byte("Moving file content.")
			suite.writeDirectory(srcDir)
			suite.writeDirectory(destDir)
			suite.writeFile(filepath.Join(srcDir, fileName), content)
			suite.moveFile(srcDir, destDir, fileName)
			srcEntries := suite.readDirectory(srcDir)
			destEntries := suite.readDirectory(destDir)
			readContent := suite.readFile(filepath.Join(destDir, fileName))
			assert.Equal(suite.T(), 0, len(srcEntries))
			assert.Equal(suite.T(), 1, len(destEntries))
			assert.Equal(suite.T(), content, readContent)
		},
		func() {
			fileName := randomFileName()
			content := make([]byte, 10*1024*1024) // 10 MB of data
			for i := range content {
				content[i] = byte(i % 256)
			}
			suite.writeFile(fileName, content)
			readContent := suite.readFile(fileName)
			assert.Equal(suite.T(), content, readContent)
		},
	}

	for i := 0; i < 100; i++ { // Execute 100 random operations
		operation := operations[rand.Intn(len(operations))]
		operation()
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond) // Random delay between 0 and 1000 milliseconds
	}
}

func TestBridgeGuardTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
