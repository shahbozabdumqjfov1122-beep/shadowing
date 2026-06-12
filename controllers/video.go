package controllers

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	"gorm.io/gorm"

	"shadowing/database"
	"shadowing/models"
)

type VideoController struct {
	beego.Controller
}

// 💡 ADMIN TEKSHIRUVINI BIR JOYGA YIĞAMIZ (Prepare har bir so'rovdan oldin ishlaydi)
// Faqat admin huquqi kerak bo'lgan metodlar uchun ishlaydi
func (c *VideoController) checkAdmin() bool {
	if c.GetSession("admin") == nil {
		c.Redirect("/password", 302)
		return false
	}
	return true
}

func (c *VideoController) Get() {
	if !c.checkAdmin() {
		return
	}
	c.TplName = "admin/videos.html"
}

func (c *VideoController) AdminVideos() {
	if !c.checkAdmin() {
		return
	}

	var videos []models.Video
	database.DB.Find(&videos)

	c.Data["Videos"] = videos
	c.TplName = "admin/videos.html"
}

func (c *VideoController) AddVideo() {
	if !c.checkAdmin() {
		return
	}

	err := c.Ctx.Request.ParseMultipartForm(32 << 20)
	if err != nil {
		c.Ctx.WriteString("Form xato: " + err.Error())
		return
	}

	title := c.Ctx.Request.FormValue("title")
	description := c.Ctx.Request.FormValue("description")
	levelStr := c.Ctx.Request.FormValue("level")

	level, err := strconv.Atoi(strings.TrimSpace(levelStr))
	if err != nil || level < 1 || level > 20 {
		c.Ctx.WriteString("Xatolik: Daraja 1 dan 20 gacha bo'lishi kerak!")
		return
	}

	videoDir := "uploads/videos/"
	if err := os.MkdirAll(videoDir, os.ModePerm); err != nil {
		c.Ctx.WriteString("Papka yaratishda xato!")
		return
	}

	file, header, err := c.Ctx.Request.FormFile("video")
	if err != nil {
		c.Ctx.WriteString("Video topilmadi!")
		return
	}
	defer file.Close()

	videoFileName := sanitizeFilename(header.Filename)
	videoPath := videoDir + videoFileName

	dst, err := os.Create(videoPath)
	if err != nil {
		c.Ctx.WriteString("Fayl yaratishda xato!")
		return
	}

	_, err = io.Copy(dst, file)
	dst.Close()

	if err != nil {
		os.Remove(videoPath)
		c.Ctx.WriteString("Video yozishda xato!")
		return
	}

	newVideo := models.Video{
		Title:       title,
		Description: description,
		Thumbnail:   "/uploads/videos/default_video_icon.png",
		VideoPath:   "/" + videoPath,
		Level:       level,
	}

	if err := database.DB.Create(&newVideo).Error; err != nil {
		c.Ctx.WriteString("DB ga saqlanmadi!")
		return
	}

	c.Redirect("/admin/videos", 302)
}
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	re := regexp.MustCompile(`[^\w\-.]`)
	name = re.ReplaceAllString(name, "_")
	return name
}

func (c *VideoController) Watch() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Ctx.WriteString("ID xato")
		return
	}

	var video models.Video
	if err := database.DB.Preload("Audios", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).Preload("Audios.Words").First(&video, id).Error; err != nil {
		c.Ctx.WriteString("Video topilmadi")
		return
	}

	var words []models.AudioWord
	var audioText string

	if len(video.Audios) > 0 {
		words = video.Audios[0].Words
		audioText = video.Audios[0].Text
	}

	c.Data["Video"] = video
	c.Data["AudioText"] = audioText
	c.Data["Words"] = words

	c.TplName = "watch.html"
}

func (c *VideoController) Edit() {
	if !c.checkAdmin() {
		return
	}
	id := c.Ctx.Input.Param(":id")

	var video models.Video
	if err := database.DB.First(&video, id).Error; err != nil {
		c.Ctx.WriteString("Video topilmadi")
		return
	}

	var audios []models.VideoAudio
	database.DB.Where("video_id = ?", video.ID).Find(&audios)

	type AudioView struct {
		ID        uint
		Path      string
		Text      string
		WordsText string
	}

	var audioViews []AudioView

	for _, audio := range audios {
		var words []models.AudioWord
		database.DB.Where("video_audio_id = ?", audio.ID).Find(&words)

		var lines []string
		for _, w := range words {
			lines = append(lines, fmt.Sprintf("%s=%s", w.English, w.Uzbek))
		}

		audioViews = append(audioViews, AudioView{
			ID:        audio.ID,
			Path:      audio.Path,
			Text:      audio.Text,
			WordsText: strings.Join(lines, "\n"),
		})
	}

	c.Data["Video"] = video
	c.Data["AudioViews"] = audioViews
	c.TplName = "admin/edit.html"
}

func (c *VideoController) Update() {
	fmt.Println("========== VIDEO UPDATE START ==========")

	if !c.checkAdmin() {
		return
	}

	id := c.Ctx.Input.Param(":id")
	fmt.Println("VIDEO ID:", id)

	var video models.Video
	if err := database.DB.First(&video, id).Error; err != nil {
		fmt.Println("VIDEO TOPILMADI:", err)
		c.Ctx.WriteString("Video topilmadi")
		return
	}

	video.Title = c.GetString("title")
	video.Description = c.GetString("description")

	if level, err := strconv.Atoi(c.GetString("level")); err == nil {
		video.Level = level
	}

	// 🔥 1. RASM FAYLINI (THUMBNAIL) YANGILASH
	imgFile, imgHeader, err := c.GetFile("thumbnail") // HTML da name="thumbnail" bo'ladi
	if err == nil && imgHeader != nil {
		fmt.Println("YANGI RASM FAYL:", imgHeader.Filename)
		defer imgFile.Close()

		imgDir := "uploads/thumbnails/"
		_ = os.MkdirAll(imgDir, os.ModePerm)
		imgPath := imgDir + imgHeader.Filename

		if err := c.SaveToFile("thumbnail", imgPath); err == nil {
			// Eski rasmni o'chirish (agar u default rasm bo'lmasa)
			if video.Thumbnail != "" && video.Thumbnail != "/uploads/videos/default_video_icon.png" {
				_ = os.Remove("." + video.Thumbnail)
			}
			video.Thumbnail = "/" + imgPath
			fmt.Println("RASM SAQLANDI:", imgPath)
		} else {
			fmt.Println("RASM SAVE ERROR:", err)
		}
	}

	// 2. VIDEO FAYLNI YANGILASH (O'zgarishsiz qoladi)
	videoFile, videoHeader, err := c.GetFile("video")
	if err == nil && videoHeader != nil {
		fmt.Println("YANGI VIDEO FAYL:", videoHeader.Filename)
		defer videoFile.Close()

		dir := "uploads/videos/"
		_ = os.MkdirAll(dir, os.ModePerm)
		path := dir + videoHeader.Filename

		if err := c.SaveToFile("video", path); err == nil {
			if video.VideoPath != "" {
				_ = os.Remove("." + video.VideoPath)
			}
			video.VideoPath = "/" + path
			fmt.Println("VIDEO SAVED:", path)
		}
	}

	if err := database.DB.Save(&video).Error; err != nil {
		fmt.Println("VIDEO UPDATE ERROR:", err)
		c.Ctx.WriteString("Saqlashda xatolik yuz berdi")
		return
	}

	fmt.Println("VIDEO UPDATED SUCCESS")
	fmt.Println("========== VIDEO UPDATE END ==========")

	c.Redirect("/admin", 302)
}

func (c *VideoController) UploadRecording() {
	idStr := c.Ctx.Input.Param(":id")
	videoID, _ := strconv.Atoi(idStr)

	file, _, err := c.GetFile("audio")
	if err != nil {
		c.Ctx.WriteString("audio topilmadi")
		return
	}
	defer file.Close()

	dir := "uploads/user-recordings/"
	_ = os.MkdirAll(dir, os.ModePerm)

	fileName := fmt.Sprintf("%d_%d.webm", videoID, time.Now().Unix())
	path := dir + fileName

	if err := c.SaveToFile("audio", path); err != nil {
		c.Ctx.WriteString("save error")
		return
	}

	recording := models.UserRecording{
		VideoID:  uint(videoID),
		FilePath: "/" + path,
	}
	database.DB.Create(&recording)

	c.Ctx.WriteString("ok")
}

func (c *VideoController) DeleteRecording() {
	idStr := c.Ctx.Input.Param(":id")
	var recording models.UserRecording

	if err := database.DB.First(&recording, idStr).Error; err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Ctx.WriteString("Recording topilmadi")
		return
	}

	// 💡 Fayl yo'li aniq standartga keltirildi
	_ = os.Remove("." + recording.FilePath)

	database.DB.Delete(&recording)
	c.Ctx.Output.SetStatus(200)
	c.Ctx.WriteString("ok")
}

func (c *VideoController) DeleteAudio() {
	if !c.checkAdmin() {
		return
	}
	audioID := c.Ctx.Input.Param(":id")

	var audio models.VideoAudio
	if err := database.DB.First(&audio, audioID).Error; err == nil {
		database.DB.Where("video_audio_id = ?", audio.ID).Delete(&models.AudioWord{})
		// 💡 Standart bo'yicha nuqta qo'shildi
		_ = os.Remove("." + audio.Path)
		database.DB.Delete(&audio)
	}

	c.Redirect(c.Ctx.Request.Referer(), 302)
}

func (c *VideoController) AddAudio() {
	if !c.checkAdmin() {
		return
	}

	// 🔥 100MB limit qo'shish
	if err := c.Ctx.Request.ParseMultipartForm(100 << 20); err != nil {
		c.Ctx.WriteString("Fayl juda katta yoki forma xato: " + err.Error())
		return
	}

	videoIDStr := c.Ctx.Input.Param(":id")
	videoID, err := strconv.Atoi(videoIDStr)
	if err != nil {
		c.Ctx.WriteString("ID xato")
		return
	}

	audioFile, audioHeader, err := c.GetFile("audio")
	if err != nil || audioHeader == nil {
		c.Redirect(c.Ctx.Request.Referer(), 302)
		return
	}
	defer audioFile.Close()

	dir := "uploads/audio/"
	_ = os.MkdirAll(dir, os.ModePerm)
	path := dir + audioHeader.Filename

	if err := c.SaveToFile("audio", path); err != nil {
		c.Redirect(c.Ctx.Request.Referer(), 302)
		return
	}

	audioText := c.GetString("audio_text")
	newAudio := models.VideoAudio{
		VideoID: uint(videoID),
		Path:    "/" + path,
		Text:    audioText,
	}

	if err := database.DB.Create(&newAudio).Error; err != nil {
		c.Redirect(c.Ctx.Request.Referer(), 302)
		return
	}

	wordsText := c.GetString("words")
	if wordsText != "" {
		lines := strings.Split(wordsText, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			word := models.AudioWord{
				VideoAudioID: newAudio.ID,
				English:      strings.TrimSpace(parts[0]),
				Uzbek:        strings.TrimSpace(parts[1]),
			}
			database.DB.Create(&word)
		}
	}

	c.Redirect(c.Ctx.Request.Referer(), 302)
}
func (c *VideoController) UpdateSingleAudio() {
	if !c.checkAdmin() {
		return
	}
	audioID := c.Ctx.Input.Param(":id")

	var audio models.VideoAudio
	if err := database.DB.First(&audio, audioID).Error; err != nil {
		c.Ctx.WriteString("Audio topilmadi")
		return
	}

	dynamicFormKey := fmt.Sprintf("audio_text_%s", audioID)
	audio.Text = c.GetString(dynamicFormKey)
	database.DB.Save(&audio)

	database.DB.Where("video_audio_id = ?", audio.ID).Delete(&models.AudioWord{})

	// 🔥 TUZATILDI: Endi har bir audio o'ziga tegishli dinamik lug'at satrini o'qiydi
	dynamicWordsKey := fmt.Sprintf("words_%s", audioID)
	wordsText := c.GetString(dynamicWordsKey)

	if wordsText != "" {
		lines := strings.Split(wordsText, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				word := models.AudioWord{
					VideoAudioID: audio.ID,
					English:      strings.TrimSpace(parts[0]),
					Uzbek:        strings.TrimSpace(parts[1]),
				}
				database.DB.Create(&word)
			}
		}
	}

	c.Redirect(c.Ctx.Request.Referer(), 302)
}

func (c *VideoController) DeleteVideo() {
	if !c.checkAdmin() {
		return
	}
	id := c.Ctx.Input.Param(":id")

	var video models.Video
	if err := database.DB.First(&video, id).Error; err != nil {
		c.Ctx.WriteString("Video topilmadi")
		return
	}

	var audios []models.VideoAudio
	database.DB.Where("video_id = ?", video.ID).Find(&audios)
	for _, audio := range audios {
		database.DB.Where("video_audio_id = ?", audio.ID).Delete(&models.AudioWord{})
		_ = os.Remove("." + audio.Path)
		database.DB.Delete(&audio)
	}

	_ = os.Remove("." + video.VideoPath)
	database.DB.Delete(&video)

	c.Redirect("/admin", 302)
}
