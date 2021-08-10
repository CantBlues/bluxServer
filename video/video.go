package video

import (
	"blux/global"
	"blux/db"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"crypto/md5"
	"encoding/hex"
	"io"
	"path"
)

// PATH is video file directory
const PATH string = global.PATH

func VideosChanged(path string) bool {
	var info db.Info
	db.DB.First(&info)
	_, files := getVideoFiles(path)
	filesMd5 := md5.New()
	for _, file := range files {
		filesMd5.Write([]byte(file))
	}
	newMd5 := filesMd5.Sum(nil)
	if hex.EncodeToString(newMd5) == info.Video {
		return false
	}
	info.Video = hex.EncodeToString(newMd5)
	db.DB.Save(&info)
	return true
}

func AudiosChanged(path string) bool {
	var info db.Info
	db.DB.First(&info)
	_, files := getAudioFiles(path)
	filesMd5 := md5.New()
	for _, file := range files {
		filesMd5.Write([]byte(file))
	}
	newMd5 := filesMd5.Sum(nil)
	if hex.EncodeToString(newMd5) == info.Audio {
		return false
	}
	info.Audio = hex.EncodeToString(newMd5)
	db.DB.Save(&info)
	return true
}

func CompareVideosMd5() (del, add map[string]string) {
	var videos []db.Video
	_, files := getVideoFiles(PATH)
	filesMap, rowsMap := map[string]string{}, map[string]string{}
	for _, file := range files {
		md5, _ := md5File(file)
		filesMap[md5] = file
	}
	database := db.DB
	database.Select("Md5").Find(&videos)
	for _, video := range videos {
		rowsMap[video.Md5] = ""
	}
	for i := range rowsMap {
		_, exist := filesMap[i]
		if exist {
			delete(filesMap, i)
			delete(rowsMap, i)
		}
	}
	return rowsMap, filesMap
}

func CompareAudiosMd5() (del, add map[string]string) {
	var audios []db.Audio
	_, files := getAudioFiles(PATH)
	filesMap, rowsMap := map[string]string{}, map[string]string{}
	for _, file := range files {
		md5, _ := md5File(file)
		filesMap[md5] = file
	}
	database := db.DB
	database.Select("Md5").Find(&audios)
	for _, Audio := range audios {
		rowsMap[Audio.Md5] = ""
	}
	for i := range rowsMap {
		_, exist := filesMap[i]
		if exist {
			delete(filesMap, i)
			delete(rowsMap, i)
		}
	}
	return rowsMap, filesMap
}

func md5File(fileName string) (string, error) {
	var result []byte
	file, err := os.Open(fileName)
	if err != nil {
		return "error", err
	}
	defer file.Close()
	buffer := make([]byte, 256)
	for i := 0; i < 30; i++ {
		if i%3 == 0 {
			ret, _ := file.Seek(0, 1)
			file.Seek(ret*5, 0)
		}
		bytesread, err := file.Read(buffer)
		result = append(result, buffer[:bytesread]...)
		if err == io.EOF || bytesread == 0 {
			//fmt.Println("break!!!")
			break
		}
	}
	hash := md5.New()
	hash.Write(result)
	hash.Write([]byte(fileName))
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func getVideoFiles(vpath string) (int, []string) {
	rd, _ := ioutil.ReadDir(vpath)
	videoFiles := []string{}
	suffixList := []interface{}{".mp4", ".mkv", ".avi", ".MPG", ".MPEG",  ".rm", ".rmvb",".mov",".wmv",  }
	count := 0
	for _, file := range rd {
		if file.IsDir() {
			fileCount, filesList := getVideoFiles(vpath + file.Name() + "/")
			videoFiles = append(videoFiles, filesList...)
			count += fileCount
		} else if file.Size() > 5*1024*1024 {
			suffix := path.Ext(file.Name())
			if ok := inArray(suffix, suffixList); ok {
				videoFiles = append(videoFiles, vpath+file.Name())
				count++
			}
		}
	}
	return count, videoFiles
}

func getAudioFiles(vpath string) (int, []string) {
	rd, _ := ioutil.ReadDir(vpath)
	audioFiles := []string{}
	suffixList := []interface{}{".mp3", ".wav", ".amr",  }
	count := 0
	for _, file := range rd {
		if file.IsDir() {
			fileCount, filesList := getAudioFiles(vpath + file.Name() + "/")
			audioFiles = append(audioFiles, filesList...)
			count += fileCount
		} else {
			suffix := path.Ext(file.Name())
			if ok := inArray(suffix, suffixList); ok {
				audioFiles = append(audioFiles, vpath+file.Name())
				count++
			}
		}
	}
	return count, audioFiles
}


func inArray(need interface{}, needArr []interface{}) bool {
	for _, v := range needArr {
		if need == v {
			return true
		}
	}
	return false
}

func dealVideo(filename, md5 string) {
	videoInfo, err := getVideoInfo(filename)
	if err != nil {
		return
	}
	duration := int(videoInfo["duration"].(float64))
	width := videoInfo["width"].(int) / 2
	height := videoInfo["height"].(int) / 2
	step := duration / 100
	var imgs []*bytes.Buffer
	if _, err := os.Stat(PATH + `imgs`); err != nil && os.IsNotExist(err) {
		os.Mkdir(PATH+`imgs`, 666)
	}
	for i := 0; i < 100; i++ {
		seek := step * i
		imgBytes := framer(filename, convertTime(seek), width, height)
		if i == 1 {
			ioutil.WriteFile(PATH+`imgs\`+md5+"thumb.jpg", imgBytes.Bytes(), 666)
		}
		imgs = append(imgs, imgBytes)
	}
	target := mergeImgs(imgs, width, height)
	if _, err := os.Stat(PATH + "imgs"); err != nil && os.IsNotExist(err) {
		os.Mkdir("imgs", 666)
	}
	file, _ := os.Create(PATH + "imgs/" + md5 + "process.jpg")
	defer file.Close()
	jpeg.Encode(file, target, &jpeg.Options{Quality: 70})
}

func framer(filename, seek string, width, height int) *bytes.Buffer {
	cmd := exec.Command("ffmpeg", "-ss", seek, "-i", filename, "-vframes", "1", "-s", fmt.Sprintf("%dx%d", width, height), "-f", "singlejpeg", "-")
	buffer := new(bytes.Buffer)
	cmd.Stdout = buffer
	erro := new(bytes.Buffer)
	cmd.Stderr = erro
	if err := cmd.Run() ;err != nil {
		fmt.Println("could not generate frame ",err," ",erro)
	}
	return buffer
}

// hide command window when 'Exec' executes
// Firstly import "syscall"

// cmd_path := "C:\\Windows\\system32\\cmd.exe"
// cmd_instance := exec.Command(cmd_path, "/c", "notepad")
// cmd_instance.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
// cmd_output, err := cmd_instance.Output()

func getVideoInfo(filename string) (map[string]interface{}, error) {
	type streams struct {
		Width     int
		Height    int
		Duration  string
		CodecType string `json:"codec_type"`
	}
	type info struct {
		Streams []streams `json:"streams"`
	}
	cmd := exec.Command("ffprobe", "-of", "json", "-show_streams", filename)
	output := new(bytes.Buffer)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		fmt.Println("get video info failed")
		return map[string]interface{}{}, err
	}
	infos := new(info)
	json.Unmarshal(output.Bytes(), &infos)
	var videoInfo streams
	for _, v := range infos.Streams {
		if v.CodecType == "video" {
			videoInfo = v
		}
	}
	duration, _ := strconv.ParseFloat(videoInfo.Duration, 32)
	ret := map[string]interface{}{"width": videoInfo.Width, "height": videoInfo.Height, "duration": duration}
	return ret, nil
}

func convertTime(duration int) string {
	hours := duration / 3600
	minutes := (duration % 3600) / 60
	s := (duration % 3600) % 60
	return fmt.Sprintf("%02s:%02s:%02s", strconv.Itoa(hours), strconv.Itoa(minutes), strconv.Itoa(s))
}

func mergeImgs(imgs []*bytes.Buffer, w, h int) *image.RGBA {
	target := image.NewRGBA(image.Rect(0, 0, w*10, h*10))
	for i, img := range imgs {
		decodedImg,_ := jpeg.Decode(img)
		position := decodedImg.Bounds().Add(image.Pt((i%10)*w, (i/10)*h))
		draw.Draw(target, position, decodedImg, decodedImg.Bounds().Min, draw.Src)
	}
	return target
}
