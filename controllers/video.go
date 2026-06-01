package controllers

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/server/web"

	"shadowing/database"
	"shadowing/models"
)

type VideoController struct {
	beego.Controller
}

func (c *VideoController) Get() {
	isAdmin := c.GetSession("admin")
	if isAdmin == nil {
		c.Redirect("/password", 302)
		return
	}
	c.TplName = "admin/videos.html"
}
func (c *VideoController) AdminVideos() {
	isAdmin := c.GetSession("admin")
	if isAdmin == nil {
		c.Redirect("/password", 302)
		return
	}
	var videos []models.Video
	database.DB.Find(&videos)

	c.Data["Videos"] = videos
	c.TplName = "admin/videos.html"
}
func (c *VideoController) AddVideo() {
	isAdmin := c.GetSession("admin")
	if isAdmin == nil {
		c.Redirect("/password", 302)
		return
	}
	title := c.GetString("title")
	description := c.GetString("description")
	levelStr := c.GetString("level") // 🔥 level_id emas, oddiy level

	// 🔥 LEVEL VALIDATION (1 - 20)
	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 1 || level > 20 {
		c.Ctx.WriteString("Xatolik: Daraja 1 dan 20 gacha bo'lishi kerak!")
		return
	}

	// Video fayl
	video, videoHeader, err := c.GetFile("video")
	if err != nil {
		c.Ctx.WriteString("Xatolik: Video yuklanmadi!")
		return
	}
	defer video.Close()

	// papka
	videoDir := "uploads/videos/"
	if err := os.MkdirAll(videoDir, os.ModePerm); err != nil {
		log.Println("Papka xatosi:", err)
	}

	videoPath := videoDir + videoHeader.Filename

	// saqlash
	if err := c.SaveToFile("video", videoPath); err != nil {
		log.Println("Save error:", err)
		c.Ctx.WriteString("Video saqlanmadi!")
		return
	}

	// DB ga yozish
	newVideo := models.Video{
		Title:       title,
		Description: description,
		Thumbnail:   "/uploads/videos/default_video_icon.png",
		VideoPath:   "/" + videoPath,
		Level:       level, // 🔥 ENDI INTEGER
	}

	if err := database.DB.Create(&newVideo).Error; err != nil {
		log.Println("DB error:", err)
		c.Ctx.WriteString("DB ga saqlanmadi!")
		return
	}

	c.Redirect("/admin/videos", 302)
}
func (c *VideoController) Watch() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Ctx.WriteString("ID xato")
		return
	}

	var video models.Video
	// GORM zanjiri orqali Videoga tegishli hamma audiolarni va ularning so'zlarini yuklaymiz
	if err := database.DB.Preload("Audios", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC") // Audiolar tartib bilan chiqishi uchun
	}).Preload("Audios.Words").First(&video, id).Error; err != nil {
		c.Ctx.WriteString("Video topilmadi")
		return
	}

	var words []models.AudioWord
	var audioText string

	// Birinchi marta sahifa ochilganda 1-audio ma'lumotlari default bo'lib turadi
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
	isAdmin := c.GetSession("admin")
	if isAdmin == nil {
		c.Redirect("/password", 302)
		return
	}
	id := c.Ctx.Input.Param(":id")

	var video models.Video
	if err := database.DB.First(&video, id).Error; err != nil {
		c.Ctx.WriteString("Video topilmadi")
		return
	}

	// 1. Videoga tegishli BARCHA audiolarni bazadan olamiz
	var audios []models.VideoAudio
	if err := database.DB.Where("video_id = ?", video.ID).Find(&audios).Error; err != nil {
		fmt.Println("Audiolarni yuklashda xatolik:", err)
	}

	// 2. HTML-ga qulay formatda jo'natish uchun vaqtinchalik struktura tuzamiz
	type AudioView struct {
		ID        uint
		Path      string
		Text      string
		WordsText string // "Hello=Salom\nApple=Olma" ko'rinishida saqlaydi
	}

	var audioViews []AudioView

	// 3. Har bir audioning so'zlarini alohida yuklab, stringga aylantiramiz
	for _, audio := range audios {
		var words []models.AudioWord
		database.DB.Where("video_audio_id = ?", audio.ID).Find(&words)

		var lines []string
		for _, w := range words {
			lines = append(lines, fmt.Sprintf("%s=%s", w.English, w.Uzbek))
		}

		// Strukturani to'ldiramiz
		audioViews = append(audioViews, AudioView{
			ID:        audio.ID,
			Path:      audio.Path,
			Text:      audio.Text,
			WordsText: strings.Join(lines, "\n"),
		})
	}

	// 4. Ma'lumotlarni HTML shabloniga uzatamiz
	c.Data["Video"] = video
	c.Data["AudioViews"] = audioViews // HTML-da range qiladiganimiz shu
	c.TplName = "admin/edit.html"
}

func (c *VideoController) Update() {
	fmt.Println("========== VIDEO UPDATE START ==========")

	// 1. Admin sessiyasini tekshirish
	isAdmin := c.GetSession("admin")
	if isAdmin == nil {
		fmt.Println("ADMIN SESSION YO'Q")
		c.Redirect("/password", 302)
		return
	}

	id := c.Ctx.Input.Param(":id")
	fmt.Println("VIDEO ID:", id)

	// 2. Videoni bazadan topish
	var video models.Video
	if err := database.DB.First(&video, id).Error; err != nil {
		fmt.Println("VIDEO TOPILMADI:", err)
		c.Ctx.WriteString("Video topilmadi")
		return
	}

	// 3. Video matnli ma'lumotlarini yangilash
	video.Title = c.GetString("title")
	video.Description = c.GetString("description")

	level, err := strconv.Atoi(c.GetString("level"))
	if err == nil {
		video.Level = level
	}

	// 4. VIDEO FAYLNI YANGILASH (Agar tahrirlashda yangi .mp4 fayl yuklangan bo'lsa)
	videoFile, videoHeader, err := c.GetFile("video")
	if err == nil && videoHeader != nil {
		fmt.Println("YANGI VIDEO FAYL:", videoHeader.Filename)
		defer videoFile.Close()

		dir := "uploads/videos/"
		os.MkdirAll(dir, os.ModePerm)
		path := dir + videoHeader.Filename

		if err := c.SaveToFile("video", path); err != nil {
			fmt.Println("VIDEO SAVE ERROR:", err)
		} else {
			fmt.Println("VIDEO SAVED:", path)
			video.VideoPath = "/" + path
		}
	} else {
		fmt.Println("YANGI VIDEO FAYL YUKLANMADI (Eski video fayli qoladi)")
	}

	// 5. Video modelidagi o'zgarishlarni bazaga saqlash
	if err := database.DB.Save(&video).Error; err != nil {
		fmt.Println("VIDEO UPDATE ERROR:", err)
		c.Ctx.WriteString("Saqlashda xatolik yuz berdi")
		return
	}

	fmt.Println("VIDEO UPDATED SUCCESS")
	fmt.Println("========== VIDEO UPDATE END ==========")

	// Hammasi yaxshi tugasa, ro'yxatga qaytaramiz
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
	os.MkdirAll(dir, os.ModePerm)

	fileName := fmt.Sprintf(
		"%d_%d.webm",
		videoID,
		time.Now().Unix(),
	)

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

	// Faylni serverdan o'chirish
	filePath := "." + recording.FilePath
	_ = os.Remove(filePath)

	// DB dan o'chirish
	database.DB.Delete(&recording)

	// 🔥 REDIRECT O'RNIGA ODDIYGINA STATUS 200 VA "OK" QAYTARAMIZ
	c.Ctx.Output.SetStatus(200)
	c.Ctx.WriteString("ok")
}
func (c *VideoController) DeleteAudio() {
	// Admin tekshiruvi...
	audioID := c.Ctx.Input.Param(":id")

	var audio models.VideoAudio
	if err := database.DB.First(&audio, audioID).Error; err == nil {
		// 1. Avval shu audioga tegishli so'zlarni o'chiramiz
		database.DB.Where("video_audio_id = ?", audio.ID).Delete(&models.AudioWord{})
		// 2. Faylini serverdan o'chirish (ixtiyoriy)
		os.Remove(strings.TrimPrefix(audio.Path, "/"))
		// 3. Audioni o'zini o'chiramiz
		database.DB.Delete(&audio)
	}

	// Ortga (qaysi sahifadan kelgan bo'lsa o'sha yerga) qaytarish
	c.Redirect(c.Ctx.Request.Referer(), 302)
}
func (c *VideoController) AddAudio() {
	fmt.Println("========== YANGI AUDIO QO'SHISH START ==========")

	// 1. Admin tekshiruvi
	isAdmin := c.GetSession("admin")
	if isAdmin == nil {
		fmt.Println("ADMIN SESSION YO'Q")
		c.Redirect("/password", 302)
		return
	}

	videoIDStr := c.Ctx.Input.Param(":id")
	videoID, err := strconv.Atoi(videoIDStr)
	if err != nil {
		fmt.Println("VIDEO ID XATO:", err)
		c.Ctx.WriteString("ID xato")
		return
	}

	// 2. HTML formadan audio faylni olish ("audio" name orqali)
	audioFile, audioHeader, err := c.GetFile("audio")
	if err != nil || audioHeader == nil {
		fmt.Println("AUDIO FAYL YUKLASHDA XATOLIK:", err)
		c.Redirect(c.Ctx.Request.Referer(), 302)
		return
	}
	defer audioFile.Close()

	// 3. Faylni server papkasiga saqlash
	dir := "uploads/audio/"
	os.MkdirAll(dir, os.ModePerm)
	path := dir + audioHeader.Filename

	if err := c.SaveToFile("audio", path); err != nil {
		fmt.Println("AUDIO FAYLNI SAQLASHDA XATOLIK:", err)
		c.Redirect(c.Ctx.Request.Referer(), 302)
		return
	}
	fmt.Println("AUDIO FAYL SAQLANDI:", path)

	// 4. VideoAudio obyektini yaratish va DB ga yozish
	audioText := c.GetString("audio_text") // HTML textarea name="audio_text"

	newAudio := models.VideoAudio{
		VideoID: uint(videoID),
		Path:    "/" + path,
		Text:    audioText,
	}

	if err := database.DB.Create(&newAudio).Error; err != nil {
		fmt.Println("BASEGA AUDIONI YOZISHDA XATOLIK:", err)
		c.Redirect(c.Ctx.Request.Referer(), 302)
		return
	}
	fmt.Println("YANGI AUDIO BAZAGA YOZILDI. ID:", newAudio.ID)

	// 5. So'zlarni parslash va saqlash (HTML textarea name="words")
	wordsText := c.GetString("words")
	if wordsText != "" {
		lines := strings.Split(wordsText, "\n")
		fmt.Println("KIRITILGAN SO'ZLAR SATRI:", len(lines))

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				fmt.Println("FORMAT XATO (O'tkazib yuborildi):", line)
				continue
			}

			word := models.AudioWord{
				VideoAudioID: newAudio.ID,
				English:      strings.TrimSpace(parts[0]),
				Uzbek:        strings.TrimSpace(parts[1]),
			}

			if err := database.DB.Create(&word).Error; err != nil {
				fmt.Println("SO'ZNI BAZAGA YOZISHDA XATOLIK:", err)
			} else {
				fmt.Println("SO'Z SAQLANDI:", word.English, "=>", word.Uzbek)
			}
		}
	}

	fmt.Println("========== YANGI AUDIO QO'SHISH END ==========")

	// Hammasi muvaffaqiyatli tugasa, xuddi shu edit sahifasini o'ziga qaytaradi
	// va yangi qo'shilgan audio darhol ro'yxatda paydo bo'ladi
	c.Redirect(c.Ctx.Request.Referer(), 302)

}
func (c *VideoController) UpdateSingleAudio() {
	audioID := c.Ctx.Input.Param(":id")

	var audio models.VideoAudio
	if err := database.DB.First(&audio, audioID).Error; err != nil {
		c.Ctx.WriteString("Audio topilmadi")
		return
	}

	// 🔥 FORMADAN MATNNI AYMAN SHU AUDIONING ID'SI BILAN CHAQIRIB OLAMIZ
	// Masalan: audio_text_3, audio_text_5 va h.k.
	dynamicFormKey := fmt.Sprintf("audio_text_%s", audioID)
	audio.Text = c.GetString(dynamicFormKey)

	// Bazaga faqat shu audioning o'zini saqlaymiz
	database.DB.Save(&audio)

	// So'zlarni yangilash qismi (O'zgarishsiz qoladi)
	database.DB.Where("video_audio_id = ?", audio.ID).Delete(&models.AudioWord{})
	wordsText := c.GetString("words") // Agar har bir audioning lug'atini ham dynamic qilmoqchi bo'lsangiz, buni ham words_{{.ID}} qilsa bo'ladi.

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
	isAdmin := c.GetSession("admin")
	if isAdmin == nil {
		c.Redirect("/password", 302)
		return
	}

	id := c.Ctx.Input.Param(":id")

	var video models.Video
	if err := database.DB.First(&video, id).Error; err != nil {
		c.Ctx.WriteString("Video topilmadi")
		return
	}

	// Avval audiolarni o'chir
	var audios []models.VideoAudio
	database.DB.Where("video_id = ?", video.ID).Find(&audios)
	for _, audio := range audios {
		database.DB.Where("video_audio_id = ?", audio.ID).Delete(&models.AudioWord{})
		os.Remove(strings.TrimPrefix(audio.Path, "/"))
		database.DB.Delete(&audio)
	}

	// Video faylini o'chir
	os.Remove(strings.TrimPrefix(video.VideoPath, "/"))

	// DB dan o'chir
	database.DB.Delete(&video)

	c.Redirect("/admin", 302)
}
