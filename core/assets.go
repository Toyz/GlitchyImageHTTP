package core

import (
	"bufio"
	"encoding/hex"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
)

type AssetTools struct {
	versions map[string]string
}

var AssetManager *AssetTools
var assetFolder string

func (v *AssetTools) New() {
	assetFolder = GetEnv("ASSET_FOLDER_PUBLIC", "./assets/public")

	ResetVersions()
	Render.AddFunc("_V", AssetManager.GetVersion)

	if strings.EqualFold(GetEnv("USE_ASSET_MONITOR", "true"), "true") {
		go setUpMonitor() // This is just in case we modify any JS files (if not hosted on the docker)
	}
}

func setUpMonitor() {
	w := watcher.New()

	w.SetMaxEvents(1)
	w.FilterOps(watcher.Create, watcher.Write, watcher.Remove)

	go func() {
		for {
			select {
			case event := <-w.Event:
				if !event.IsDir() {
					switch event.Op {
					case watcher.Remove:
						_, ok := AssetManager.versions[event.FileInfo.Name()]
						if ok {
							delete(AssetManager.versions, event.FileInfo.Name())
						}

						break

					case watcher.Write:
						file := event.FileInfo.Name()
						hash, _ := hashFileCrc32(event.Path, 0xedb88320)

						AssetManager.versions[file] = hash

						break

					case watcher.Create:
						file := event.FileInfo.Name()
						hash, _ := hashFileCrc32(event.Path, 0xedb88320)

						AssetManager.versions[file] = hash

						break
					}
				}
			case <-w.Error:
				return
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive(assetFolder); err != nil {
		log.Fatalln(err)
	}

	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}
}

func ResetVersions() {
	AssetManager = &AssetTools{
		versions: make(map[string]string),
	}

	fileList := make([]string, 0)
	e := filepath.Walk(assetFolder, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return err
		}

		fileList = append(fileList, path)

		return err
	})

	if e != nil {
		panic(e)
	}

	for _, file := range fileList {
		if fileExist(file) {
			fullPath := strings.Split(file, string(os.PathSeparator))
			hash, _ := hashFileCrc32(file, 0xedb88320)
			AssetManager.versions[fullPath[len(fullPath)-1]] = hash
		}
	}
}

func (v *AssetTools) GetVersion(file ...string) string {
	fullPath := strings.Join(file, "")
	fullPath = strings.TrimSpace(fullPath)

	if val, ok := v.versions[fullPath]; ok {
		return val
	}
	return "undefined"
}

func (v *AssetTools) FileContents(fp string) string {
	by, _ := ioutil.ReadFile(fp)

	return string(by)
}

// # is always a comment in all files we read by lines...
func (v *AssetTools) ReadFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if len(line) <= 0 {
			continue
		}

		if !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

func fileExist(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func hashFileCrc32(filePath string, polynomial uint32) (string, error) {
	var returnCRC32String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnCRC32String, err
	}
	defer file.Close()
	tablePolynomial := crc32.MakeTable(polynomial)
	hash := crc32.New(tablePolynomial)
	if _, err := io.Copy(hash, file); err != nil {
		return returnCRC32String, err
	}
	hashInBytes := hash.Sum(nil)[:]
	returnCRC32String = hex.EncodeToString(hashInBytes)
	return returnCRC32String, nil
}
