package main

import (
	"archive/zip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	_ "image/png"
	"io"
	"io/fs"
	"math/rand/v2"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.bug.st/serial"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetPin(serialNumber string) map[string]any {
	const possibleChars string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	indem := ""
	for range 16 {
		indem += string(possibleChars[rand.IntN(len(possibleChars))])
	}

	req, err := http.NewRequest("GET", "https://play.date/api/v2/device/register/"+serialNumber+"/get", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Idempotency-Key", indem)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var res map[string]any
	err = json.Unmarshal(body, &res)
	if err != nil {
		panic(err)
	}

	return res
}

var accessToken string = ""

func (a *App) FinishRegistration(serialNumber string) map[string]any {
	resp, err := http.Get("https://play.date/api/v2/device/register/" + serialNumber + "/complete/get")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var res map[string]any
	err = json.Unmarshal(body, &res)
	if err != nil {
		panic(err)
	}

	return res
}

var (
	pdosFilename  string
	pdkeyFilename string
)

func (a *App) DownloadPlaydateOS(accessToken string) {
	req, err := http.NewRequest("GET", "https://play.date/api/v2/firmware/?current_version=2.6.2", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Token "+accessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var res map[string]string
	err = json.Unmarshal(body, &res)
	if err != nil {
		panic(err)
	}

	decryptKey, err := base64.StdEncoding.DecodeString(res["decryption_key"])
	if err != nil {
		panic(err)
	}

	f, err := os.CreateTemp("", "PlaydateOS.*.pdos")
	pdosFilename = f.Name()
	if err != nil {
		panic(err)
	}
	defer f.Close()
	resp, err = http.Get(res["url"])
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		panic(err)
	}

	f, err = os.CreateTemp("", "PlaydateOS.*.pdkey")
	pdkeyFilename = f.Name()
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.Write(decryptKey)
	if err != nil {
		panic(err)
	}
}

type LauncherInfo struct {
	filename       string
	targetFilename string
	url            string
	extractPath    string
}

var launchers map[string]LauncherInfo = map[string]LauncherInfo{}

func (a *App) DownloadLauncher(selectedLauncher string, url string, filename string, targetFilename string) {
	f, err := os.CreateTemp("", filename)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		panic(err)
	}

	launchers[selectedLauncher] = LauncherInfo{f.Name(), targetFilename, url, ""}
}

func (a *App) DownloadUPSM(url string) string {
	f, err := os.CreateTemp("", "*.upsm.zip")
	if err != nil {
		panic(err)
	}

	defer f.Close()
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		panic(err)
	}

	return f.Name()
}

var pdosExtractPath string = ""

func (a *App) ExtractPlaydateOS(funnyloader bool) {
	extractPath, err := os.MkdirTemp("", "PlaydateOS.*")
	if err != nil {
		panic(err)
	}

	zipReader, err := zip.OpenReader(pdosFilename)
	if err != nil {
		panic(err)
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		filePath := filepath.Join(extractPath, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				panic(err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}
		srcFile, err := f.Open()
		if err != nil {
			panic(err)
		}
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			panic(err)
		}

		dstFile.Close()
		srcFile.Close()
	}

	if funnyloader {
		err = os.Mkdir(filepath.Join(extractPath, "System", "Launchers"), os.ModePerm)
		if err != nil {
			panic(err)
		}
		err = os.Rename(filepath.Join(extractPath, "System", "Launcher.pdx"), filepath.Join(extractPath, "System", "Launchers", "StockLauncher.pdx"))
		if err != nil {
			panic(err)
		}
	} else {
		err = os.Rename(filepath.Join(extractPath, "System", "Launcher.pdx"), filepath.Join(extractPath, "System", "StockLauncher.pdx"))
		if err != nil {
			panic(err)
		}
	}

	pdosExtractPath = extractPath
}

func (a *App) ExtractLauncher(selectedLauncher string) {
	extractPath, err := os.MkdirTemp("", selectedLauncher)
	if err != nil {
		panic(err)
	}

	zipReader, err := zip.OpenReader(launchers[selectedLauncher].filename)
	if err != nil {
		panic(err)
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		filePath := filepath.Join(extractPath, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				panic(err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}
		srcFile, err := f.Open()
		if err != nil {
			panic(err)
		}
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			panic(err)
		}

		dstFile.Close()
		srcFile.Close()
	}

	if launcher, ok := launchers[selectedLauncher]; ok {
		launcher.extractPath = extractPath
		launchers[selectedLauncher] = launcher
	}
}

func (a *App) ExtractUPSM(path string) string {
	extractPath, err := os.MkdirTemp("", "*.upsm")
	if err != nil {
		panic(err)
	}

	zipReader, err := zip.OpenReader(path)
	if err != nil {
		panic(err)
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		filePath := filepath.Join(extractPath, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				panic(err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}
		srcFile, err := f.Open()
		if err != nil {
			panic(err)
		}
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			panic(err)
		}

		dstFile.Close()
		srcFile.Close()
	}

	info, err := os.ReadDir(extractPath)
	if err != nil {
		panic(err)
	}

	if len(info) == 1 && info[0].IsDir() && filepath.Ext(info[0].Name()) == ".upsm" {
		return filepath.Join(extractPath, info[0].Name())
	}
	return extractPath
}

func (a *App) GenerateUPSM(selectedLauncher string, pdxPath string) string {
	upsmPath, err := os.MkdirTemp("", "upsm")
	if err != nil {
		panic(err)
	}

	err = os.Rename(filepath.Join(launchers[selectedLauncher].extractPath, pdxPath), filepath.Join(upsmPath, "Launcher.pdx"))
	if err != nil {
		panic(err)
	}

	f, err := os.Create(filepath.Join(upsmPath, "upsminfo"))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if launchers[selectedLauncher].targetFilename == "Launcher.pdx" {
		if _, err = f.WriteString("loader=true"); err != nil {
			panic(err)
		}
	} else if _, err = f.WriteString("\nlauncherpath=" + launchers[selectedLauncher].targetFilename); err != nil {
		panic(err)
	}

	pdxinfo, err := os.ReadFile(filepath.Join(upsmPath, "Launcher.pdx", "pdxinfo"))
	if err != nil {
		panic(err)
	}

	launcherMetadata := ParseConfig(string(pdxinfo))

	if name, ok := launcherMetadata["name"]; ok {
		_, err = f.WriteString("\nname=" + name)
		if err != nil {
			panic(err)
		}
	}
	if version, ok := launcherMetadata["version"]; ok {
		_, err = f.WriteString("\nversion=" + version)
		if err != nil {
			panic(err)
		}
	}
	if author, ok := launcherMetadata["author"]; ok {
		_, err = f.WriteString("\nauthor=" + author)
		if err != nil {
			panic(err)
		}
	}
	if description, ok := launcherMetadata["description"]; ok {
		_, err = f.WriteString("\ndescription=" + description)
		if err != nil {
			panic(err)
		}
	}

	return upsmPath
}

func (a *App) InstallUPSM(path string, funnyLoader bool) {
	var mods []map[string]string

	content, err := os.ReadFile(filepath.Join(pdosExtractPath, "System", "upsm_mods.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			mods = make([]map[string]string, 0)
		} else {
			panic(err)
		}
	} else {
		if err := json.Unmarshal(content, &mods); err != nil {
			panic(err)
		}
	}

	upsminfoContent, err := os.ReadFile(filepath.Join(path, "upsminfo"))
	if err != nil {
		panic(err)
	}
	upsminfo := ParseConfig(string(upsminfoContent))

	modInfo := map[string]string{}
	if name, ok := upsminfo["name"]; ok {
		modInfo["name"] = name
	}
	if description, ok := upsminfo["description"]; ok {
		modInfo["description"] = description
	}
	if version, ok := upsminfo["version"]; ok {
		modInfo["version"] = version
	}
	if author, ok := upsminfo["author"]; ok {
		modInfo["author"] = author
	}

	mods = append(mods, modInfo)
	newContent, err := json.Marshal(mods)
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(filepath.Join(pdosExtractPath, "System", "upsm_mods.json"), newContent, 0644); err != nil {
		panic(err)
	}

	// Install System Apps

	apps, err := filepath.Glob(filepath.Join(path, "*.pdx"))
	if err != nil {
		panic(err)
	}
	for _, app := range apps {
		if filepath.Base(app) == "Launcher.pdx" && funnyLoader {
			if err = os.Rename(app, filepath.Join(pdosExtractPath, "System", "Launchers", upsminfo["launcherpath"])); err != nil {
				panic(err)
			}
			continue
		}
		if err = os.Rename(app, filepath.Join(pdosExtractPath, "System", filepath.Base(app))); err != nil {
			panic(err)
		}
	}

	// Handle `.gen` Directories

	gens, err := filepath.Glob(filepath.Join(path, "*.gen"))
	if err != nil {
		panic(err)
	}
	for _, gen := range gens {
		geninfoContent, err := os.ReadFile(filepath.Join(gen, "geninfo"))
		if err != nil {
			panic(err)
		}
		geninfo := ParseConfig(string(geninfoContent))

		if err = os.CopyFS(filepath.Join(gen, geninfo["path"]), os.DirFS(filepath.Join(pdosExtractPath, "System", strings.TrimSuffix(filepath.Base(gen), "gen")+"pdx"))); err != nil {
			panic(err)
		}

		// Extract PDZ Files

		pdzFiles, err := filepath.Glob(filepath.Join(gen, geninfo["path"], "*.pdz"))
		if err != nil {
			panic(err)
		}
		for _, pdzPath := range pdzFiles {
			pdzFile, err := os.Open(pdzPath)
			if err != nil {
				println("pdz1")
				panic(err)
			}
			defer pdzFile.Close()
			pdz := &PDZ{
				buffer:  pdzFile,
				Entries: make(map[string]*Entry),
			}
			if err = pdz.readHeader(); err != nil {
				println("pdz2")
				panic(err)
			}
			if err = pdz.readEntries(); err != nil {
				println("pdz3")
				panic(err)
			}
			if err = pdz.saveEntries(filepath.Join(gen, geninfo["path"]), false); err != nil {
				println("pdz4")
				panic(err)
			}
			if err = os.Remove(filepath.Join(pdzPath)); err != nil {
				println("pdz5")
				panic(err)
			}
		}
		// Unluac

		f, err := os.CreateTemp("", "unluac.*.jar")
		if err != nil {
			println("unluac1")
			panic(err)
		}

		defer f.Close()
		resp, err := http.Get("https://github.com/scratchminer/unluac/releases/latest/download/unluac.jar")
		if err != nil {
			println("unluac2")
			panic(err)
		}
		defer resp.Body.Close()
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			println("unluac3")
			panic(err)
		}

		err = filepath.WalkDir(filepath.Join(gen, geninfo["path"]), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && filepath.Ext(d.Name()) == ".luac" {
				cmd := exec.Command("java", "-jar", f.Name(), "-o", path[:len(path)-1], path)
				if err = cmd.Run(); err != nil {
					return err
				}
				if err = os.Remove(path); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			println("unluac4")
			panic(err)
		}

		// Remove CoreLibs

		if _, err = os.Stat(filepath.Join(gen, geninfo["path"], "CoreLibs")); err == nil {
			if err = os.RemoveAll(filepath.Join(gen, geninfo["path"], "CoreLibs")); err != nil {
				panic(err)
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			panic(err)
		}

		// Patch

		patches, err := filepath.Glob(filepath.Join(gen, "*.patch"))
		if err != nil {
			println("patch1")
			panic(err)
		}
		if len(patches) < 1 {
			panic("No Patches Found in `.gen` Directory.")
		}

		if err = os.Chdir(gen); err != nil {
			println("patch3")
			panic(err)
		}

		f, err = os.Open(patches[0])
		if err != nil {
			println("patch4")
			panic(err)
		}
		defer f.Close()

		cmd := exec.Command("patch", strings.Split(geninfo["options"], " ")...)
		cmd.Stdin = f
		if err = cmd.Run(); err != nil {
			println("patch5")
			panic(err)
		}

		// Build

		if err = os.RemoveAll(filepath.Join(pdosExtractPath, "System", strings.TrimSuffix(filepath.Base(gen), "gen")+"pdx")); err != nil {
			println("build1")
			panic(err)
		}

		cmd = exec.Command("pdc", "-s", geninfo["path"], filepath.Join(pdosExtractPath, "System", strings.TrimSuffix(filepath.Base(gen), "gen")+"pdx"))
		if err = cmd.Run(); err != nil {
			println("build2")
			panic(err)
		}
	}

	println(pdosExtractPath)
}

var pdosPatchedPath string = ""

func (a *App) CompressPlaydateOS() {
	zipFile, err := os.CreateTemp("", "PlaydateOS-Patched.*.pdos")
	if err != nil {
		panic(err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	absSourceDir, err := filepath.Abs(pdosExtractPath)
	if err != nil {
		panic(err)
	}

	err = filepath.Walk(absSourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		relPath, err := filepath.Rel(absSourceDir, filePath)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		zipName := strings.ReplaceAll(relPath, "\\", "/")
		if info.IsDir() {
			zipName += "/"
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = zipName
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
	if err != nil {
		panic(err)
	}

	pdosPatchedPath = zipFile.Name()
}

func (a *App) GetSerialPorts() []string {
	ports, err := serial.GetPortsList()
	if err != nil {
		panic(err)
	}
	return ports
}

func (a *App) UploadPatchedPlaydateOS(selectedPort string) {
	port, err := serial.Open(selectedPort, &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		Parity:   serial.NoParity,
	})
	if err != nil {
		panic(err)
	}
	defer port.Close()

	_, err = port.Write([]byte("\ndatadisk\n"))
	if err != nil {
		panic(err)
	}

	time.Sleep(8 * time.Second)

	info, err := FindMount("PLAYDATE")
	if err != nil {
		panic(err)
	}

	i, err := os.Open(pdosPatchedPath)
	if err != nil {
		panic(err)
	}
	defer i.Close()

	o, err := os.Create(filepath.Join(info.MountPoint, "PlaydateOS-Patched.pdos"))
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(o, i)
	if err != nil {
		panic(err)
	}

	if err = o.Sync(); err != nil {
		panic(err)
	}
	if err = o.Close(); err != nil {
		panic(err)
	}

	i, err = os.Open(pdkeyFilename)
	if err != nil {
		panic(err)
	}
	defer i.Close()

	o, err = os.Create(filepath.Join(info.MountPoint, "PlaydateOS.pdkey"))
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(o, i)
	if err != nil {
		panic(err)
	}

	if err = o.Sync(); err != nil {
		panic(err)
	}
	if err = o.Close(); err != nil {
		panic(err)
	}

	err = UnmountAndEject(info)
	if err != nil {
		panic(err)
	}
}

func (a *App) InstallPatchedPlaydateOS(selectedPort string) {
	port, err := serial.Open(selectedPort, &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	})
	if err != nil {
		panic(err)
	}
	defer port.Close()

	_, err = port.Write([]byte("\nfwup /PlaydateOS-Patched.pdos /PlaydateOS.pdkey\n"))
	if err != nil {
		panic(err)
	}
}

func (a *App) CleanUp(selectedPort string) {
	port, err := serial.Open(selectedPort, &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		Parity:   serial.NoParity,
	})
	if err != nil {
		panic(err)
	}
	defer port.Close()

	_, err = port.Write([]byte("\ndatadisk\n"))
	if err != nil {
		panic(err)
	}

	time.Sleep(8 * time.Second)

	info, err := FindMount("PLAYDATE")
	if err != nil {
		panic(err)
	}

	os.Remove(filepath.Join(info.MountPoint, "PlaydateOS-Patched.pdos"))
	os.Remove(filepath.Join(info.MountPoint, "PlaydateOS.pdkey"))

	err = UnmountAndEject(info)
	if err != nil {
		panic(err)
	}
}
