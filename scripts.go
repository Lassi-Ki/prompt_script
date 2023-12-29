package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jszwec/csvutil"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"regexp"
	"strconv"
)

var raw = `{
   "txt2img_payload": {
       "enable_hr": false,
       "denoising_strength": 0,
       "hr_scale": 2,
       "hr_upscaler": "",
       "hr_second_pass_steps": 0,
       "hr_resize_x": 0,
       "hr_resize_y": 0,
       "prompt": "breathtaking selfie photograph of astronaut floating in space, earth in the background. award-winning, professional, highly detailed <lora:2c16563e-8086-43ca-abcd-7f8098e81de3:0.5> <lora:8813f115-db38-4861-9450-c89653d66c8f:1>",
       "negative_prompt": "anime, cartoon, graphic, text, painting, crayon, graphite, abstract glitch, blurry",
       "styles": [
           ""
       ],
       "seed": -1,
       "subseed": -1,
       "subseed_strength": 0,
       "seed_resize_from_h": -1,
       "seed_resize_from_w": -1,
       "sampler_name": "",
       "batch_size": 4,
       "n_iter": 1,
       "steps": 30,
       "cfg_scale": 7,
       "width": 1024,
       "height": 1024,
       "restore_faces": false,
       "tiling": false,
       "do_not_save_samples": false,
       "do_not_save_grid": false,
       "eta": 0,
       "s_churn": 0,
       "s_tmax": 0,
       "s_tmin": 0,
       "s_noise": 1,
       "override_settings": {},
       "override_settings_restore_afterwards": true,
       "script_args": [],
       "sampler_index": "Euler a",
       "script_name": "",
       "send_images": true,
       "save_images": false,
       "alwayson_scripts": {}
   },
   "model": "sd_xl_base_1.0.safetensors",
   "task": "text-to-image"
}`

type PredictionStyle struct {
	Uuid           string `json:"uuid"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Category       string `json:"category"`
	Label          string `json:"label"`
	Keywords       string `json:"keywords"`
	Preview        string `json:"preview"`
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt"`
	Payload        string `json:"payload"`
	Parameters     string `json:"parameters"`
	Default        int    `json:"default"`
	Visibility     int    `json:"visibility"`
	InputType      string `json:"input_type" gorm:"column:input_type"`
	Sequence       int    `gorm:"sequence"`
	ImagePath      string `gorm:"image_path"`
}

type Model struct {
	Name  string
	Type  string
	Model string
}

type Lora struct {
	Name   string `csv:"Name"`
	Prompt string `csv:"Prompt"`
}

func (*Model) TableName() string {
	return "demos"
}

func (*PredictionStyle) TableName() string {
	return "test_style"
}

var db *gorm.DB

var sequence = 1

var tmp map[string]interface{}

func init() {
	_ = json.Unmarshal([]byte(raw), &tmp)
	dsn := "423ac025-7450-416e-8214-a9882b702c19:f327xgPXLgiDiHlh@tcp(k8s-jumpserv-jumpserv-43ba1922f6-105fd2b24be59fac.elb.ap-southeast-1.amazonaws.com:33060)" +
		"/dev02_pay_aigc_service?charset=utf8mb4&parseTime=True&loc=Local"
	db, _ = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Error)})
}

/**
(masterpiece:1.5), (half-body shot:1.8), (extremely intricate:1.2), featured on artstation, 8k, highly detailed, nikon z9,(freezing ice mountain and landscape full of ice in the background:1.8), (dim light:1.8), (cinematic lighting:1.8),(elegant:1.5), <lora:Stef_1024Capt:0.8>, an award winning illustration digital painting of ohwx woman wearing wearing blue parka with fur hood
Negative prompt: (helmet:1.8),(planet:1.8)(anime:1.8),crown,,leaves, tree, plants, canvas frame, cartoon, 3d, ((disfigured)), ((bad art)), ((deformed)),((extra limbs)),((close up)),((b&w)), weird colors, blurry, (((duplicate))), ((morbid)), ((mutilated)), [out of frame], extra fingers, mutated hands, ((poorly drawn hands)), ((poorly drawn face)), (((mutation))), (((deformed))), ((ugly)), blurry, ((bad anatomy)), (((bad proportions))), ((extra limbs)), cloned face, (((disfigured))), out of frame, ugly, extra limbs, (bad anatomy), gross proportions, (malformed limbs), ((missing arms)), ((missing legs)), (((extra arms))), (((extra legs))), mutated hands, (fused fingers), (too many fingers), (((long neck))), Photoshop, video game, ugly, tiling, poorly drawn hands, poorly drawn feet, poorly drawn face, out of frame, mutation, mutated, extra limbs, extra legs, extra arms, disfigured, deformed, cross-eye, body out of frame, blurry, bad art, bad anatomy, 3d render

Steps: 30, Sampler: DPM2 a, CFG scale: 9, Seed: 2542333859, Size: 1026x1368, Model hash: 31e35c80fc, Model: sd_xl_base_1.0, Lora hashes: "Stef_1024Capt: 5e83f7c30aa0", Version: v1.5.1
*/

//regex := regexp.MustCompile(`((?P<prompt>.*)(\n)?)((Negative prompt: (?P<negative>.*)\n(?P<payload>.*), Size.*)?)`)

// Underwater Cat  |
//
// girl Japanese Kimono Style |Scientist Wannabe | Comic Space Astronaut
// boy Autumn Boy | Dragon Warrior| Cosmic Space Boy
// man Luffy
// woman

func main() {
	helper("./csv/boy.csv", "boy")
	helper("./csv/girl.csv", "girl")
	helper("./csv/man.csv", "man")
	helper("./csv/woman.csv", "woman")
}

func helper(path string, category string) {

	loras := LoraFromCsv(path)

	for _, lora := range loras {
		//for i := 0; i < 1; i++ {
		//regex := regexp.MustCompile(`(\s*)(?P<prompt>.*)\n*((Negative prompt: (?P<negative>.*)\s*Steps: (?P<steps>[^,]+).*Sampler: (?P<sampler>[^,]+).*CFG scale: (?P<scale>[^,]+).*Seed: (?P<seed>[^,]+).*Size.*)?)`)
		regex := regexp.MustCompile(`(\s*)(?P<prompt>.*)\n*((\s?)(Negative prompt: (?P<negative>.*)\s*Steps: (?P<steps>[^,]+).*Sampler: (?P<sampler>[^,]+).*CFG scale: (?P<scale>[^,]+).*Seed: (?P<seed>[^,]+).*Model: (?P<model>[^,]+).*)?)`)
		matches := regex.FindStringSubmatch(lora.Prompt)
		matchMap := make(map[string]string)
		if len(matches) >= 0 {
			for i, name := range regex.SubexpNames() {
				if i != 0 && name != "" {
					matchMap[name] = matches[i]
				}
			}

		}
		if matchMap["model"] == "" {
			fmt.Println("========", lora.Name, "==========")
		}
		if matchMap["steps"] != "" {
			steps, err := strconv.Atoi(matchMap["steps"])
			if err != nil {
				fmt.Println(lora.Name)
				return
			}

			scale, err := strconv.ParseFloat(matchMap["scale"], 64)
			if err != nil {
				fmt.Println(lora.Name)
				return
			}

			seed, err := strconv.Atoi(matchMap["seed"])
			if err != nil {
				fmt.Println(lora.Name)
				return
			}
			tmp["txt2img_payload"].(map[string]interface{})["steps"] = steps
			tmp["txt2img_payload"].(map[string]interface{})["sampler_index"] = matchMap["sampler"]
			tmp["txt2img_payload"].(map[string]interface{})["cfg_scale"] = scale
			tmp["txt2img_payload"].(map[string]interface{})["seed"] = seed
			tmp["vae"] = "sd_xl_base_1.0_vae.safetensors"

		}
		bf := bytes.NewBuffer([]byte{})
		jsonEncoder := json.NewEncoder(bf)
		jsonEncoder.SetEscapeHTML(false)
		jsonEncoder.Encode(tmp)

		style := PredictionStyle{
			Uuid:           uuid.New().String(),
			Name:           lora.Name,
			Description:    lora.Name,
			Category:       category,
			Label:          category,
			Keywords:       category,
			Preview:        category,
			Prompt:         matchMap["prompt"],
			NegativePrompt: matchMap["negative"],
			Payload:        bf.String(),
			Parameters:     "",
			Default:        0,
			Visibility:     1,
			InputType:      "text",
			Sequence:       sequence,
			ImagePath:      "",
		}

		sequence += 1

		//m := &Model{
		//	Name:  lora.Name,
		//	Type:  name,
		//	Model: matchMap["model"],
		//}
		// 自动迁移（创建表）
		err_ := db.AutoMigrate(&PredictionStyle{})
		if err_ != nil {
			fmt.Println("Error creating table:", err_)
		} else {
			fmt.Println("Table created successfully")
		}
		err := db.Create(&style).Error
		if err != nil {
			fmt.Println(lora.Name)
		}

	}
}

func LoraFromCsv(path string) []Lora {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var loras []Lora

	if err := csvutil.Unmarshal(file, &loras); err != nil {
		log.Fatal("error:", err)
	}
	return loras
}
