package ad_material

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	db2 "github.com/mao888/golang-guide/project/data-sync/db"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// AdMaterialSyncSuccess mapped from table <ad_material_sync_success>
type AdMaterialSyncSuccess struct {
	ID           int32  `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	MaterilaType int32  `gorm:"column:materila_type;not null;default:1" json:"materila_type"` // '附件文件类型， 1: file,  2: image,3: video',
	MaterialID   int32  `gorm:"column:material_id;not null" json:"material_id"`               // 素材id
	Name         string `gorm:"column:name;not null" json:"name"`                             // 素材名称 拼接而成
	URL          string `gorm:"column:url;not null" json:"url"`                               // 素材源地址
	MaterialMd5  string `gorm:"column:material_md5;not null" json:"material_md5"`             // 素材md5
	AccountID    string `gorm:"column:account_id;not null" json:"account_id"`                 // 所属账户
	Creator      int32  `gorm:"column:creator;not null" json:"creator"`                       // 创建者
	Type         int32  `gorm:"column:type;not null;default:1" json:"type"`                   // 上传日志类型 1：Facebook 2：YouTube 3：优量汇 4：今日头条
	SuccessID    string `gorm:"column:success_id;not null" json:"success_id"`                 // fb 返回 结果id
	BatchID      string `gorm:"column:batch_id;not null" json:"batch_id"`                     // 批处理id
}

// GetVideoMaterialResp 获取视频素材应答字段
type GetVideoMaterialResp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
	Data      struct {
		List []struct {
			BitRate          int           `json:"bit_rate"`
			CreateTime       string        `json:"create_time"`
			Duration         float64       `json:"duration"`
			Filename         string        `json:"filename"`
			Format           string        `json:"format"`
			Height           int           `json:"height"`
			Id               string        `json:"id"`
			Labels           []string      `json:"labels"`
			MaterialId       int64         `json:"material_id"`
			OrganizationTags []interface{} `json:"organization_tags"`
			PosterUrl        string        `json:"poster_url"`
			Signature        string        `json:"signature"`
			Size             int           `json:"size"`
			Source           string        `json:"source"`
			Url              string        `json:"url"`
			Width            int           `json:"width"`
		} `json:"list"`
		PageInfo struct {
			Page        int `json:"page"`
			PageSize    int `json:"page_size"`
			TotalNumber int `json:"total_number"`
			TotalPage   int `json:"total_page"`
		} `json:"page_info"`
	} `json:"data"`
}

// GetImageMaterialResp 获取图片素材应答字段
type GetImageMaterialResp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
	Data      struct {
		List []struct {
			Aigc       bool   `json:"aigc"`
			CreateTime string `json:"create_time"`
			Filename   string `json:"filename"`
			Format     string `json:"format"`
			Height     int    `json:"height"`
			Id         string `json:"id"`
			MaterialId int64  `json:"material_id"`
			Signature  string `json:"signature"`
			Size       int    `json:"size"`
			Url        string `json:"url"`
			Width      int    `json:"width"`
		} `json:"list"`
		PageInfo struct {
			Page        int `json:"page"`
			PageSize    int `json:"page_size"`
			TotalNumber int `json:"total_number"`
			TotalPage   int `json:"total_page"`
		} `json:"page_info"`
	} `json:"data"`
}

var videoImageIdMaterialIdMap = map[string]int64{}

func RunAdMaterialSyncSuccess() {

	// 1、获取 accountID
	var accountID []string
	err := db2.MySQLClientCruiser.Table("ad_material_sync_success").Distinct("account_id").
		Where("type = ?", 4).Where("id <= ?", 729610).Group("account_id").Find(&accountID).Error
	if err != nil {
		glog.Errorf("查询错误：%s", err)
		return
	}
	// 2、获取 每个accountID下的 视频素材,图片素材
	//	并保存视频id、图片id与素材id的映射关系
	for _, v := range accountID {
		glog.Infof("accountID: %s", v)
		// 获取视频素材 保存视频id 与 素材id 映射关系
		err := GetVideoMaterial(v)
		if err != nil {
			glog.Errorf("获取视频素材错误：%s", err)
			return
		}
		// 获取图片素材 保存图片id 与 素材id 映射关系
		err = GetImageMaterial(v)
		if err != nil {
			glog.Errorf("获取图片素材错误：%s", err)
			return
		}
	}
	// 3、根据 accountID 获取 ad_material_sync_success
	var adMaterialSyncSuccess = make([]*AdMaterialSyncSuccess, 0)
	//var accountIDStr string
	//for _, a := range accountID {
	//	accountIDStr += fmt.Sprintf("'%s',", a)
	//}
	err = db2.MySQLClientCruiser.Table("ad_material_sync_success").
		Where("account_id IN ?", accountID).Where("id <= ?", 729610).
		Find(&adMaterialSyncSuccess).Error
	if err != nil {
		glog.Errorf("查询错误：%s", err)
		return
	}

	// 3、更新 ad_material_sync_success
	var sqlList []string
	for _, adMaterialSyncSuccess := range adMaterialSyncSuccess {

		// 更新 ad_material_sync_success 中的 success_id
		//fmt.Println("id: ", adMaterialSyncSuccess.ID)
		//// 更新 ad_material_sync_success
		//if adMaterialSyncSuccess.MaterilaType == 1 {
		//	// 更新 ad_material_sync_success
		//	adMaterialSyncSuccess.MaterialID = videoImageIdMaterialIdMap[adMaterialSyncSuccess.SuccessID]
		//	err := db2.MySQLClientCruiserTest.Table("ad_material_sync_success").
		//		Where("id = ?", adMaterialSyncSuccess.ID).
		//		Updates(adMaterialSyncSuccess).Error
		//	if err != nil {
		//		fmt.Println("更新 ad_material_sync_success 错误：", err)
		//		return
		//	}
		//} else if adMaterialSyncSuccess.MaterilaType == 2 {
		//	// 更新 ad_material_sync_success
		//	adMaterialSyncSuccess.MaterialID = videoImageIdMaterialIdMap[adMaterialSyncSuccess.SuccessID]
		//	err := db2.MySQLClientCruiserTest.Table("ad_material_sync_success").
		//		Where("id = ?", adMaterialSyncSuccess.ID).
		//		Updates(adMaterialSyncSuccess).Error
		//	if err != nil {
		//		fmt.Println("更新 ad_material_sync_success 错误：", err)
		//		return
		//	}
		//}

		// 写出所有条目的更新sql语句, 并导出到本地文件中
		glog.Infof("success_id: %s", adMaterialSyncSuccess.SuccessID)
		sql := fmt.Sprintf("UPDATE ad_material_sync_success SET success_id = %d WHERE id = %d;",
			videoImageIdMaterialIdMap[adMaterialSyncSuccess.SuccessID], adMaterialSyncSuccess.ID)
		sqlList = append(sqlList, sql)
		glog.Infof("sql: %s", sql)
	}

	// 写出到本地文件
	err = WriteToFile(sqlList)
	if err != nil {
		glog.Errorf("写出到本地文件错误：%s", err)
		return
	}
}

// GetVideoMaterial 获取视频素材
func GetVideoMaterial(advertiserId string) error {
	//url := "https://ad.oceanengine.com/open_api/2/file/video/get/"
	//resp, err := resty.New().SetRetryCount(3).R().
	//	SetHeader("Access-Token", "7975e1f425b3adb547484362d97d9551fea69e07").
	//	SetBody(map[string]interface{}{
	//		"advertiser_id": advertiserId,
	//		"page":          1,
	//		"page_size":     100,
	//	}).Get(url)
	//if err != nil {
	//	glog.Errorf("请求错误：%s", err)
	//	return err
	//}
	//fmt.Println("resp: ", string(resp.Body()))
	url := "https://ad.oceanengine.com/open_api/2/file/video/get/"
	method := "GET"
	payload := strings.NewReader(fmt.Sprintf(`{
    			"advertiser_id": %s,
   				 "page":1,
   				 "page_size":100}`, advertiserId))
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		glog.Errorf("请求错误：%s", err)
		return err
	}
	req.Header.Add("Access-Token", "7975e1f425b3adb547484362d97d9551fea69e07")
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("请求错误 视频素材：%s", err)
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("请求错误 视频素材：%s", err)
		return err
	}
	glog.Infof("resp: %s", string(body))
	// 解析返回值
	var getVideoMaterialResp GetVideoMaterialResp
	err = json.Unmarshal(body, &getVideoMaterialResp)
	if err != nil {
		glog.Errorf("解析错误：%s", err)
		return err
	}
	if getVideoMaterialResp.Code != 0 {
		glog.Errorf("请求错误：%s", getVideoMaterialResp.Message)
		return err
	}
	// 保存 视频id 与 素材id 映射关系
	for _, s := range getVideoMaterialResp.Data.List {
		videoImageIdMaterialIdMap[s.Id] = s.MaterialId
	}

	// 当total_page > 1 时，需要循环请求
	if getVideoMaterialResp.Data.PageInfo.TotalPage > 1 {
		// 从第二页开始请求
		for i := 2; i <= getVideoMaterialResp.Data.PageInfo.TotalPage; i++ {
			payload := strings.NewReader(fmt.Sprintf(`{
    			"advertiser_id": %s,
   				 "page":%d,
   				 "page_size":100}`, advertiserId, i))
			client := &http.Client{}
			req, err := http.NewRequest(method, url, payload)
			if err != nil {
				glog.Errorf("请求错误：%s", err)
				return err
			}
			req.Header.Add("Access-Token", "7975e1f425b3adb547484362d97d9551fea69e07")
			req.Header.Add("Content-Type", "application/json")
			res, err := client.Do(req)
			if err != nil {
				glog.Errorf("请求错误 视频素材：%s", err)
				return err
			}
			defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				glog.Errorf("请求错误 视频素材：%s", err)
				return err
			}
			glog.Infof("resp: %s", string(body))

			var getVideoMaterialResp GetVideoMaterialResp
			err = json.Unmarshal(body, &getVideoMaterialResp)
			if err != nil {
				glog.Errorf("解析错误：%s", err)
				return err
			}
			if getVideoMaterialResp.Code != 0 {
				glog.Errorf("请求错误：%s", getVideoMaterialResp.Message)
				return err
			}
			for _, s := range getVideoMaterialResp.Data.List {
				videoImageIdMaterialIdMap[s.Id] = s.MaterialId
			}
		}
	}
	return nil
}

// GetImageMaterial 获取图片素材
func GetImageMaterial(advertiserId string) error {
	url := "https://api.oceanengine.com/open_api/2/file/image/get/"
	method := "GET"
	payload := strings.NewReader(fmt.Sprintf(`{
    			"advertiser_id": %s,
   				 "page":1,
   				 "page_size":100}`, advertiserId))
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		glog.Errorf("请求错误：%s", err)
		return err
	}
	req.Header.Add("Access-Token", "7975e1f425b3adb547484362d97d9551fea69e07")
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("请求错误 图片素材：%s", err)
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("请求错误 图片素材：%s", err)
		return err
	}
	fmt.Println(string(body))
	// 解析返回值
	var getImageMaterialResp GetImageMaterialResp
	err = json.Unmarshal(body, &getImageMaterialResp)
	if err != nil {
		fmt.Println("解析错误：", err)
		glog.Errorf("解析错误：%s", err)
	}
	if getImageMaterialResp.Code != 0 {
		glog.Errorf("请求错误：%s", getImageMaterialResp.Message)
		return err
	}
	// 保存 视频id 与 素材id 映射关系
	for _, s := range getImageMaterialResp.Data.List {
		videoImageIdMaterialIdMap[s.Id] = s.MaterialId
	}

	// 当total_page > 1 时，需要循环请求
	if getImageMaterialResp.Data.PageInfo.TotalPage > 1 {
		// 从第二页开始请求
		for i := 2; i <= getImageMaterialResp.Data.PageInfo.TotalPage; i++ {
			url := "https://api.oceanengine.com/open_api/2/file/image/get/"
			method := "GET"
			payload := strings.NewReader(fmt.Sprintf(`{
    			"advertiser_id": %s,
   				 "page":%d,
   				 "page_size":100}`, advertiserId, i))
			client := &http.Client{}
			req, err := http.NewRequest(method, url, payload)
			if err != nil {
				glog.Errorf("请求错误：%s", err)
				return err
			}
			req.Header.Add("Access-Token", "7975e1f425b3adb547484362d97d9551fea69e07")
			req.Header.Add("Content-Type", "application/json")
			res, err := client.Do(req)
			if err != nil {
				glog.Errorf("请求错误 图片素材：%s", err)
				return err
			}
			//defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				glog.Errorf("请求错误 图片素材：%s", err)
				return err
			}
			fmt.Println(string(body))
			// 解析返回值
			var getImageMaterialResp GetImageMaterialResp
			err = json.Unmarshal(body, &getImageMaterialResp)
			if err != nil {
				glog.Errorf("解析错误：%s", err)
				return err
			}
			if getImageMaterialResp.Code != 0 {
				glog.Errorf("请求错误：%s", getImageMaterialResp.Message)
				return err
			}
			for _, s := range getImageMaterialResp.Data.List {
				videoImageIdMaterialIdMap[s.Id] = s.MaterialId
			}
		}
	}
	return nil
}

// WriteToFile 写出到本地文件 封装一个函数
func WriteToFile(sqlList []string) error {
	// 获取当前可执行文件的路径
	exePath, err := os.Executable()
	if err != nil {
		glog.Errorf("无法获取当前可执行文件的路径: %s", err)
		return err
	}

	// 解析出项目根路径
	projectRoot := filepath.Dir(filepath.Dir(exePath))

	// 拼接文件夹路径
	outputFolderPath := filepath.Join(projectRoot, "project/data-sync/mysql-to-mysql/ad_material")

	// 确保文件夹路径存在
	if err := os.MkdirAll(outputFolderPath, 0755); err != nil {
		glog.Errorf("无法创建文件夹: %s", err)
		return err
	}

	// 拼接文件路径
	filePath := filepath.Join(outputFolderPath, "update_success_id.sql")

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		glog.Errorf("无法创建文件: %s", err)
		return err
	}
	defer file.Close()

	// 逐行写入 SQL 语句
	for _, sql := range sqlList {
		_, err := file.WriteString(sql + "\n")
		if err != nil {
			glog.Errorf("写出到本地文件错误: %s", err)
			return err
		}
	}
	glog.Infof("写出到本地文件成功: %s", filePath)
	return nil
}
