package video

import (
	"blux/global"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

var (
	dstImgPath string = global.PATH
)

// HandleVideosChange is to
func HandleVideosChange(del, add map[string]string) {
	var wg sync.WaitGroup
	for i, v := range add {
		wg.Add(1)
		go func(v, i string) {
			dealVideo(v,i)
			wg.Done()
		}(v,i)
	}
	wg.Wait()
	for i := range del {
		os.Remove(dstImgPath + `imgs\` + i + "thumb.jpg")
		os.Remove(dstImgPath + `imgs\` + i + "process.jpg")
	}
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
	var imgs []string
	if _, err := os.Stat("tempImgs"); err != nil && os.IsNotExist(err) {
		os.Mkdir("tempImgs", 666)
	}
	if _, err := os.Stat(dstImgPath + `imgs`); err != nil && os.IsNotExist(err) {
		os.Mkdir(dstImgPath+`imgs`, 666)
	}
	for i := 0; i < 100; i++ {
		seek := step * i
		imgBytes := framer(filename, convertTime(seek), width, height)
		imgName := `tempImgs\` + getRandomString(10) + strconv.Itoa(i) + ".jpg"
		err := ioutil.WriteFile(imgName, imgBytes.Bytes(), 666)
		if i == 1 {
			ioutil.WriteFile(dstImgPath+`imgs\`+md5+"thumb.jpg", imgBytes.Bytes(), 666)
		}
		if err == nil {
			imgs = append(imgs, imgName)
		}
	}
	target := mergeImgs(imgs, width, height)
	if _, err := os.Stat(dstImgPath + "imgs"); err != nil && os.IsNotExist(err) {
		os.Mkdir("imgs", 666)
	}
	file, _ := os.Create(dstImgPath + "imgs/" + md5 + "process.jpg")
	defer file.Close()
	jpeg.Encode(file, target, &jpeg.Options{Quality: 70})
	for _, imgPath := range imgs {
		os.Remove(imgPath)
	}
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

func getRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func mergeImgs(imgs []string, w, h int) *image.RGBA {
	target := image.NewRGBA(image.Rect(0, 0, w*10, h*10))
	for i, imgPath := range imgs {
		img, err := os.Open(imgPath)
		defer img.Close()
		if err != nil {
			fmt.Println("open file failed!")
		}
		decodedImg, err := jpeg.Decode(img)
		position := decodedImg.Bounds().Add(image.Pt((i%10)*w, (i/10)*h))
		draw.Draw(target, position, decodedImg, decodedImg.Bounds().Min, draw.Src)
	}
	return target
}
