package main //定义报名，package main表示一个可独立执行的程序，每个 Go 应用程序都包含一个名为 main 的包
//CDBFT机制 简单模拟 赞成，反对票 (结合信用度)等级机制
import ( //导入包（的函数或者其他元素）
	"bufio"
	"crypto/sha256" //软件包sha256 实现 FIPS 180-4 中定义的 SHA224 和 SHA256 哈希算法。
	"encoding/hex"  //包十六进制实现十六进制编码和解码。
	"fmt"           // fmt 包使用函数实现 I/O 格式化（类似于 C 的 printf 和 scanf 的函数）, 格式化参数源自C，但更简单
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv" //包strconv实现了对基本数据类型的字符串表示的转换
	"strings" //打包字符串实现简单的函数来操纵 UTF-8 编码的字符串
	"time"    // 打包时间提供了测量和显示时间的功能

	"github.com/davecgh/go-spew/spew" //负责在控制台中格式化输出相应的结果
)

type Prepare struct { //PREPARE 信息结构
	Hash          string //摘要
	Index         int    //序号
	NodeID        int    //节点号
	Validator_key string //验证节点密钥
}

type Commit struct { //COMMIT 信息结构
	Hash          string //摘要
	Index         int    //序号
	NodeID        int    //节点号
	Validator_key string //验证节点密钥
}

type Message struct { // 定义POST请求的数据结构体
	Bike  []int
	types bool //信息类型 false-普通 true-紧急
}

type area struct {
	node []Node //区域内全部节点
}

type Block struct { //Block 代表组成区块链的每一个块的数据模型    //PRE-PREPARE 信息结构
	Index         int    //区块链中数据记录的位置
	Timestamp     string //时间戳，是自动确定的，并且是写入数据的时间
	Bike          []int  //交易信息 Bike就是一定区域内的自行车数量
	Hash          string //是代表这个数据记录的SHA256标识符
	PrevHash      string //是链中上一条记录的SHA256标识符
	Nonce         string //PoW中符合条件的数字
	Validator_key string //验证节点密钥
}

type Node struct { //定义节点信息
	number               int
	id                   bool    //false-恶意节点,ture-正常节点
	block                []Block //完整区块链
	address              string  //生成地址
	Validator_key        string  //生成密钥
	temp_block           Block   //临时区块-作为验证信息用
	vote                 int     //选票信息
	rank                 int     //1-投票节点  2-超级节点
	score                int     //节点当前得分
	Object_approval      int     //节点投出赞成票的节点  在原始节点中的信息 number
	Object_opposition    int     //节点投出反对票的节点  在原始节点中的信息 number
	Prepare_noise_number int
	Commit_noise_number  int
	speed                [2]int //节点运动速度
	position             [2]int //节点位置
	region               int    //节点所属区域编号
	last_regin           int    //节点前一个所属区域编号
	stay_time            int    //节点在区域内停留的时间 区块数
	types                int    //节点类型 0-固定节点，1-留守节点，2-流窜节点，3-新生节点
}

const (
	FN_number              = 25                                      //固定节点数量与区域数 最好不要修改  除固定节点外其他节点数量不宜过少，分簇结果不佳 9
	each_area_RN           = 3                                       //每个区域的留守节点32 46
	RN_number              = 75 - bad_number - bad_number_id_address //留守节点 每个固定节点边上存在each_area_RN个留守节点288 414
	SN_number              = 0                                       //流窜节点101
	NN_number              = 0                                       //新生节点102
	difficulty             = 1                                       //difficulty 代表难度系数  所产生的 Hash 前导0个数
	tokens_init            = 10                                      //节点初始的tokens均小于tokens_init
	bad_tokens_init        = 10                                      //恶意节点的初始化tokens 累积攻击500
	bad_number             = 0                                       //恶意节点数量 攻击构建区块节点  主要针对留守节点 50
	bad_number_id_address  = 0                                       //恶意克隆节点数量 身份复制 流窜节点+新生 且其值得一半均要小于流窜节点和新生节点  且产生的节点会加到node_form中
	deal_amount            = 1600                                    //区块交易数上限  deal_sum/deal_amount为区块总数
	adv_difficulty         = 1                                       // difficult下降程度
	check_point            = 10                                      //检查点协议
	tokens_day             = 8                                       //重新选举周期，小于或等于deal_sum/deal_amount为区块总数
	block_number           = 100                                     //区块数量100
	General_message_number = 0.7                                     //普通消息的区块数量占比  紧急消息的区块数量占比为1-General_message_number 两者的和为总区块
	starNodes_number       = 40                                      //竞选节点数量 需要小于节点数 216
	superStarNode_number   = 30                                      //超级节点数量 需要小于竞选节点数量 小于非超级节点数 要大于10 174
	repeat_times           = 1                                       //重复次数6
	start_cal              = 0                                       //多少个区块后计算指标（存在恶意节点时）
)

var (
	node_sum                 = FN_number + RN_number + SN_number + NN_number              //节点数
	Prepare_noise            = make([]Prepare, node_sum, node_sum)                        //预准备消息
	Commit_noise             = make([]Commit, node_sum, node_sum)                         //预确认消息
	voteNodesPool            = make([]Node, node_sum, node_sum)                           //投票节点池
	starNodesPool            = make([]Node, starNodes_number, starNodes_number)           //竞选节点池
	superStarNodesPool       = make([]Node, superStarNode_number, superStarNode_number)   //超级节点池
	Blockchain               []Block                                                      // 存放区块数据
	node_form                = make([]Node, node_sum, node_sum)                           //初始化节点信息
	block_message            = make([]Message, block_number, block_number)                //区块交易集合
	miner_number             = 0                                                          //区块生成序号
	tra_time                 time.Time                                                    //区块上链时间
	attatck_number           = 0                                                          //恶意攻击成功次数
	attatck_number_temp      = 0                                                          //观测点之间的恶意攻击次数
	safe_number              = 0                                                          //共识成功次数
	tokens_select_temp       = 0                                                          //选举次数统计
	safe_number_temp         = 0                                                          //观测点之间的共识成功次数
	node_form_attack_address = make([]Node, bad_number_id_address, bad_number_id_address) //恶意克隆节点数量
	th_1                     = 5                                                          //流窜节点向留守节点转换阈值  区块数
	th_2                     = 2                                                          //留守节点向流窜节点转换阈值  区块数
	result_tps               = make([]float64, block_number, block_number)                //每次区块共识的TPS
	result_delay             = make([]float64, block_number, block_number)                //每次区块共识的延迟
	resul_exchange           = make([]float64, block_number, block_number)                //每次区块共识的通信次数
)

// 定义函数node_init 节点初始化
func node_init() {
	var (
		FN_pool []Node //固定节点池
		RN_pool []Node //留守节点池
		SN_pool []Node //流窜节点池
		NN_pool []Node //新生节点池
	)
	//节点信息初始化
	for i := 0; i < node_sum; i++ { //节点信息初始化
		node_form[i].number = i + 1              //节点编号
		temp_string := strconv.Itoa(i + 1)       //转成string
		node_form[i].address = temp_string       //生成节点身份地址为节点编号
		node_form[i].Validator_key = temp_string //生成节点身份地址为节点秘钥
		node_form[i].id = true                   //正常节点
		node_form[i].stay_time = 0               //停留时间(区块数)
		node_form[i].Commit_noise_number = 0
		node_form[i].Prepare_noise_number = 0
		node_form[i].vote = 0
		node_form[i].rank = 1
		node_form[i].score = tokens_init
	}
	if bad_number > 0 {
		var attack_position = make([]int, bad_number, bad_number)
		for k := 0; k < bad_number; k++ {
			attack_position[k] = 12 + k //间隔为3
		}
		//var attack_position []int                                                          //恶意节点攻击位置
		//attack_position = generateRandomNumber(FN_number, FN_number+RN_number, bad_number) //设置恶意节点 主要针对留守节点
		for i := 0; i < bad_number; i++ {
			node_form[attack_position[i]].id = false //恶意节点
			node_form[attack_position[i]].score = bad_tokens_init
		}
	}

	for i := 0; i < node_sum; i++ {
		switch {
		case i < FN_number:
			node_form[i].types = 0 //固定节点
			node_form[i].speed[0] = 0
			node_form[i].speed[1] = 0
			node_form[i].region = i + 1
			node_form[i].last_regin = i + 1
			FN_pool = append(FN_pool, node_form[i])
		case i >= FN_number && i < FN_number+RN_number:
			node_form[i].types = 1 //留守节点
			node_form[i].speed[0] = 3
			node_form[i].speed[1] = 3
			node_form[i].region = 0
			node_form[i].last_regin = 0
			RN_pool = append(RN_pool, node_form[i])
		case i >= FN_number+RN_number && i < FN_number+RN_number+SN_number:
			node_form[i].types = 2 //流窜节点
			node_form[i].speed[0] = 20
			node_form[i].speed[1] = 20
			node_form[i].region = 0
			node_form[i].last_regin = 0
			SN_pool = append(SN_pool, node_form[i])
		case i >= FN_number+RN_number+SN_number && i < FN_number+RN_number+SN_number+NN_number:
			node_form[i].types = 3 //新生节点
			node_form[i].speed[0] = 3
			node_form[i].speed[1] = 3
			node_form[i].region = 0
			node_form[i].last_regin = 0
			NN_pool = append(NN_pool, node_form[i])
		}
	}
	//设置初始位置  900*900的网格，一共9个方格(固定节点数)
	temp_x := 0
	temp_y := 0
	for i := 0; i < len(FN_pool); i++ { //固定节点
		FN_pool[i].position[0] = 150 + 300*temp_x
		FN_pool[i].position[1] = 150 + 300*temp_y
		if temp_x < 2 {
			temp_x++
		} else {
			temp_x = 0
			temp_y++
		}
	}
	for i := 0; i < node_sum; i++ { //同步
		for k := 0; k < len(FN_pool); k++ {
			if node_form[i].number == FN_pool[k].number {
				node_form[i] = FN_pool[k]
			}
		}
	}
	///
	temp_x = 0
	temp_y = 0
	temp_region := 1
	for i := 0; i < len(RN_pool); i++ { //留守节点的起始位置与固定节点一致
		RN_pool[i].position[0] = 150 + 300*temp_x
		RN_pool[i].position[1] = 150 + 300*temp_y
		RN_pool[i].region = temp_region
		RN_pool[i].last_regin = temp_region
		if (i+1)%each_area_RN == 0 {
			temp_region++
			if temp_x < 2 {
				temp_x++
			} else {
				temp_x = 0
				temp_y++
			}
		}
	}
	for i := 0; i < node_sum; i++ { //同步
		for k := 0; k < len(RN_pool); k++ {
			if node_form[i].number == RN_pool[k].number {
				node_form[i] = RN_pool[k]
			}
		}
	}
	///
	for i := 0; i < len(SN_pool); i++ { //设置流窜节点 位置随机生成
		rand.Seed(time.Now().Unix())
		temp_x_SN := rand.Intn(900)
		SN_pool[i].position[0] = temp_x_SN
		time.Sleep(time.Second) //暂停一秒 确保随机值不同
		rand.Seed(time.Now().Unix())
		temp_y_SN := rand.Intn(900)
		SN_pool[i].position[1] = temp_y_SN
		SN_pool[i].region = Determine_location(SN_pool[i].position[0], SN_pool[i].position[1])
		SN_pool[i].last_regin = Determine_location(SN_pool[i].position[0], SN_pool[i].position[1])
	}
	for i := 0; i < node_sum; i++ { //同步
		for k := 0; k < len(SN_pool); k++ {
			if node_form[i].number == SN_pool[k].number {
				node_form[i] = SN_pool[k]
			}
		}
	}
	for i := 0; i < len(NN_pool); i++ { //设置新生节点 位置随机生成
		rand.Seed(time.Now().Unix())
		temp_x_NN := rand.Intn(900)
		NN_pool[i].position[0] = temp_x_NN
		time.Sleep(time.Second) //暂停一秒 确保随机值不同
		rand.Seed(time.Now().Unix())
		temp_y_NN := rand.Intn(900)
		NN_pool[i].position[1] = temp_y_NN
		NN_pool[i].region = Determine_location(NN_pool[i].position[0], NN_pool[i].position[1])
		NN_pool[i].last_regin = Determine_location(NN_pool[i].position[0], NN_pool[i].position[1])
	}
	///
	for i := 0; i < node_sum; i++ { //同步
		for k := 0; k < len(NN_pool); k++ {
			if node_form[i].number == NN_pool[k].number {
				node_form[i] = NN_pool[k]
			}
		}
	}
	///生成恶意克隆节点
	for i := 0; i < bad_number_id_address; i++ { //节点信息初始化
		node_form_attack_address[i].number = node_sum + 1 + i
		node_form_attack_address[i].id = false    //正常节点
		node_form_attack_address[i].stay_time = 0 //停留时间(区块数
		node_form_attack_address[i].Commit_noise_number = 0
		node_form_attack_address[i].Prepare_noise_number = 0
		node_form_attack_address[i].vote = 0
		node_form_attack_address[i].rank = 1
		node_form_attack_address[i].score = tokens_init
		if i < bad_number_id_address/2 {
			node_form_attack_address[i].types = 2 //前一半为流窜节点
			node_form_attack_address[i].speed[0] = 20
			node_form_attack_address[i].speed[1] = 20 //克隆流窜节点的身份 前attack_number_id_address/2个身份地址
			node_form_attack_address[i].address = SN_pool[i].address
			node_form_attack_address[i].Validator_key = SN_pool[i].Validator_key
		} else {
			node_form_attack_address[i].types = 3 //后一半为新生节点
			node_form_attack_address[i].speed[0] = 3
			node_form_attack_address[i].speed[1] = 3 //克隆新生节点的身份 前attack_number_id_address/2个身份地址
			node_form_attack_address[i].address = NN_pool[i-bad_number_id_address/2].address
			node_form_attack_address[i].Validator_key = NN_pool[i-bad_number_id_address/2].Validator_key
		}
		rand.Seed(time.Now().Unix()) //随机生成位置
		temp_x := rand.Intn(900)
		node_form_attack_address[i].position[0] = temp_x
		time.Sleep(time.Second) //暂停一秒 确保随机值不同
		rand.Seed(time.Now().Unix())
		temp_y := rand.Intn(900)
		node_form_attack_address[i].position[1] = temp_y
		node_form_attack_address[i].region = Determine_location(node_form_attack_address[i].position[0], node_form_attack_address[i].position[1])
		node_form_attack_address[i].last_regin = Determine_location(node_form_attack_address[i].position[0], node_form_attack_address[i].position[1])
		node_form = append(node_form, node_form_attack_address[i]) //恶意克隆节点加入节点群
	}
}

// 定义函数 node_move 实现节点运动
func node_move() {
	for k := 0; k < len(node_form); k++ {
		if node_form[k].types > 0 { //留守节点 流窜节点 新生节点
			rand.Seed(time.Now().Unix())
			time.Sleep(time.Millisecond)
			if temp := rand.Intn(3); temp == 0 { //产生0或1的随机数 节点随机选择方向改进
				node_form[k].position[0] = node_form[k].position[0] + node_form[k].speed[0]
			} else {
				node_form[k].position[1] = node_form[k].position[1] + node_form[k].speed[1]
			}
			if node_form[k].position[0] > 900 { //X轴越界处理
				node_form[k].position[0] = 900
				node_form[k].speed[0] = -1 * node_form[k].speed[0]
			} else if node_form[k].position[0] < 0 {
				node_form[k].position[0] = 0
				node_form[k].speed[0] = -1 * node_form[k].speed[0]
			}
			if node_form[k].position[1] > 900 { //Y轴越界处理
				node_form[k].position[1] = 900
				node_form[k].speed[1] = -1 * node_form[k].speed[1]
			} else if node_form[k].position[1] < 0 {
				node_form[k].position[1] = 0
				node_form[k].speed[1] = -1 * node_form[k].speed[1]
			}
			new_region := Determine_location(node_form[k].position[0], node_form[k].position[1]) //节点所属位置进行确认
			if node_form[k].region == new_region {
				node_form[k].stay_time = node_form[k].stay_time + 1 //一直则增加停留的时间
			} else {
				node_form[k].stay_time = 0   //停留时间清零
				if node_form[k].types == 1 { //留守节点只更新最新所属区域编号 后面节点类型变换需要
					node_form[k].region = new_region
				} else {
					node_form[k].last_regin = node_form[k].region //其他节点更换上一个与当前的
					node_form[k].region = new_region
				}
			}
		}
	}
}

// 定义函数node_type_change 实现节点身份切换
func node_type_change() {
	for i := 0; i < len(node_form); i++ {
		if node_form[i].types == 1 { //留守节点变为流窜节点  停留时间大于一定阈值且前后所属区域发生变化
			if node_form[i].stay_time > th_2 && node_form[i].last_regin != node_form[i].region {
				node_form[i].types = 2 // 更改类型与节点运动速度
				node_form[i].speed[0] = 20
				node_form[i].speed[0] = 20
			}
		} else if node_form[i].types == 2 { //流窜节点变为留守节点   停留时间大于一定阈值且前后所属区域没有发变化
			if node_form[i].stay_time > th_1 && node_form[i].last_regin == node_form[i].region {
				node_form[i].types = 1 // 更改类型与节点运动速度
				node_form[i].speed[0] = 3
				node_form[i].speed[0] = 3

			}
		} else if node_form[i].types == 3 { //新生节点变为流留守节点  更改类型与节点运动速度
			node_form[i].types = 1
			node_form[i].speed[0] = 3
			node_form[i].speed[0] = 3
		}
	}
}

// 定义函数make_genesisBlock 实现创世区块添加
func make_genesisBlock() {
	t := time.Now()
	genesisBlock := Block{}
	genesisBlock = Block{0, t.String(), make([]int, 0, 0), calculateBlockHash(genesisBlock), "", "", ""}
	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock) //追加创世区块

	for i := 0; i < len(node_form); i++ {
		node_form[i].block = append(node_form[i].block, genesisBlock) //节点同步 追加创世区块
	}
}

// 定义函数make_block_message 实现区块交易产生
func make_block_message() {
	for i := 0; i < block_number; i++ {
		temp1 := block_number * General_message_number
		if i < int(temp1) {
			for k := 0; k < deal_amount; k++ {
				block_message[i].Bike = append(block_message[i].Bike, 0) //普通消息  false-普通 true-紧急
				block_message[i].types = false
			}
		} else {
			for k := 0; k < deal_amount; k++ {
				block_message[i].Bike = append(block_message[i].Bike, 1) //紧急消息 false-普通 true-紧急
				block_message[i].types = true
			}
		}
	}
}

// 定义函数obervation_cal 完成观测点的区块共识
func obervation_cal(message []int) {
	test(message)
}

// 定义函数Node_placement 每个区域添加新生节点（保证每个区域中留守节点的稳定）
func Node_placement() {
	toal_are_2 := make([]area, FN_number, FN_number) //重新分类
	for i := 0; i < len(node_form); i++ {
		if node_form[i].types == 1 { //若节点为留守节点则按last_regin进行分类，留守节点会进行临时跨区
			switch node_form[i].last_regin { //每个区域的节点集合
			case 1:
				toal_are_2[0].node = append(toal_are_2[0].node, node_form[i])
			case 2:
				toal_are_2[1].node = append(toal_are_2[1].node, node_form[i])
			case 3:
				toal_are_2[2].node = append(toal_are_2[2].node, node_form[i])
			case 4:
				toal_are_2[3].node = append(toal_are_2[3].node, node_form[i])
			case 5:
				toal_are_2[4].node = append(toal_are_2[4].node, node_form[i])
			case 6:
				toal_are_2[5].node = append(toal_are_2[5].node, node_form[i])
			case 7:
				toal_are_2[6].node = append(toal_are_2[6].node, node_form[i])
			case 8:
				toal_are_2[7].node = append(toal_are_2[7].node, node_form[i])
			case 9:
				toal_are_2[8].node = append(toal_are_2[8].node, node_form[i])
			}
		} else { //其他类型节点则按region进行分类
			switch node_form[i].region { //每个区域的节点集合
			case 1:
				toal_are_2[0].node = append(toal_are_2[0].node, node_form[i])
			case 2:
				toal_are_2[1].node = append(toal_are_2[1].node, node_form[i])
			case 3:
				toal_are_2[2].node = append(toal_are_2[2].node, node_form[i])
			case 4:
				toal_are_2[3].node = append(toal_are_2[3].node, node_form[i])
			case 5:
				toal_are_2[4].node = append(toal_are_2[4].node, node_form[i])
			case 6:
				toal_are_2[5].node = append(toal_are_2[5].node, node_form[i])
			case 7:
				toal_are_2[6].node = append(toal_are_2[6].node, node_form[i])
			case 8:
				toal_are_2[7].node = append(toal_are_2[7].node, node_form[i])
			case 9:
				toal_are_2[8].node = append(toal_are_2[8].node, node_form[i])
			}
		}
	}

	for k := 0; k < FN_number; k++ { //补充节点
		add_flag := 0 //节点增加个数
		if len(toal_are_2[k].node) < (each_area_RN + 1) {
			add_flag = (each_area_RN + 1) - len(toal_are_2[k].node)
			var temp_node_add Node
			for j := 0; j < len(toal_are_2[k].node); j++ {
				if toal_are_2[k].node[j].types == 0 {
					temp_node_add = toal_are_2[k].node[j] //保存固定节点位置信息
				}
			}

			for j := 0; j < (each_area_RN+1)-len(toal_are_2[k].node); j++ {
				temp_5 := node_form[len(node_form)-1].number
				var temp Node
				temp.number = temp_5 + 1                //节点编号
				temp_string := strconv.Itoa(temp_5 + 1) //转成string
				temp.address = temp_string              //生成节点身份地址为节点编号
				temp.id = true                          //正常节点
				temp.stay_time = 0                      //停留时间(区块数)
				temp.Commit_noise_number = 0
				temp.Prepare_noise_number = 0

				temp.types = 3 //新生节点
				temp.speed[0] = 3
				temp.speed[1] = 3

				for m := 0; m < len(temp_node_add.block); m++ { //同步区块
					temp.block = append(temp.block, temp_node_add.block[m])
				}

				temp.position[0] = temp_node_add.position[0]
				temp.position[1] = temp_node_add.position[1]
				temp.region = Determine_location(temp.position[0], temp.position[1])
				temp.last_regin = temp.region

				node_form = append(node_form, temp) //添加
				node_sum++
			}
		}
		toal_are_2 := make([]area, FN_number, FN_number) //重新分类
		for i := 0; i < len(node_form); i++ {
			if node_form[i].types == 1 { //若节点为留守节点则按last_regin进行分类，留守节点会进行临时跨区
				switch node_form[i].last_regin { //每个区域的节点集合
				case 1:
					toal_are_2[0].node = append(toal_are_2[0].node, node_form[i])
				case 2:
					toal_are_2[1].node = append(toal_are_2[1].node, node_form[i])
				case 3:
					toal_are_2[2].node = append(toal_are_2[2].node, node_form[i])
				case 4:
					toal_are_2[3].node = append(toal_are_2[3].node, node_form[i])
				case 5:
					toal_are_2[4].node = append(toal_are_2[4].node, node_form[i])
				case 6:
					toal_are_2[5].node = append(toal_are_2[5].node, node_form[i])
				case 7:
					toal_are_2[6].node = append(toal_are_2[6].node, node_form[i])
				case 8:
					toal_are_2[7].node = append(toal_are_2[7].node, node_form[i])
				case 9:
					toal_are_2[8].node = append(toal_are_2[8].node, node_form[i])
				}
			} else { //其他类型节点则按region进行分类
				switch node_form[i].region { //每个区域的节点集合
				case 1:
					toal_are_2[0].node = append(toal_are_2[0].node, node_form[i])
				case 2:
					toal_are_2[1].node = append(toal_are_2[1].node, node_form[i])
				case 3:
					toal_are_2[2].node = append(toal_are_2[2].node, node_form[i])
				case 4:
					toal_are_2[3].node = append(toal_are_2[3].node, node_form[i])
				case 5:
					toal_are_2[4].node = append(toal_are_2[4].node, node_form[i])
				case 6:
					toal_are_2[5].node = append(toal_are_2[5].node, node_form[i])
				case 7:
					toal_are_2[6].node = append(toal_are_2[6].node, node_form[i])
				case 8:
					toal_are_2[7].node = append(toal_are_2[7].node, node_form[i])
				case 9:
					toal_are_2[8].node = append(toal_are_2[8].node, node_form[i])
				}
			}
		}
		if add_flag > 0 {
			for m := 0; m < FN_number; m++ {
				if len(toal_are_2[m].node) > (each_area_RN + 1) {
					var temp_node_delete []Node
					var temp_node_no_delete []Node
					var temp_node_form []Node
					for j := 0; j < len(toal_are_2[m].node); j++ {
						if toal_are_2[m].node[j].id == true && toal_are_2[m].node[j].types != 0 { //节点类型 0-固定节点，1-留守节点，2-流窜节点，3-新生节点
							temp_node_delete = append(temp_node_delete, toal_are_2[m].node[j]) //保存可删除的节点信息
						} else {
							temp_node_no_delete = append(temp_node_no_delete, toal_are_2[m].node[j]) //保存不可删除节点信息
						}

					}
					var temp_7 = 0
					if len(temp_node_no_delete) >= (each_area_RN + 1) {
						if add_flag-(len(toal_are_2[m].node)-(each_area_RN+1)) > 0 {
							temp_7 = len(toal_are_2[m].node) - (each_area_RN + 1)
							add_flag = add_flag - (len(toal_are_2[m].node) - (each_area_RN + 1))
						} else {
							temp_7 = add_flag
							add_flag = 0
						}
					} else {
						temp_8 := (each_area_RN + 1) - len(temp_node_no_delete)
						temp_9 := len(temp_node_delete) - temp_8
						if add_flag-temp_9 > 0 {
							temp_7 = temp_9
							add_flag = add_flag - temp_9
						} else {
							temp_7 = add_flag
							add_flag = 0
						}
					}

					temp_delete := generateRandomNumber(0, len(temp_node_delete), temp_7)
					for l := 0; l < len(node_form); l++ {
						flag := 1
						for j := 0; j < len(temp_delete); j++ {
							if temp_node_delete[temp_delete[j]].number == node_form[l].number {
								flag = 0 //不保存
								break
							}
						}
						if flag == 1 {
							temp_node_form = append(temp_node_form, node_form[l]) //保存
						}
					}
					node_sum = node_sum - (temp_7) //同步
					node_form = node_form[0:0]
					node_form = temp_node_form
				}
				if add_flag <= 0 { //重新分类并退出
					toal_are_2 := make([]area, FN_number, FN_number) //重新分类
					for i := 0; i < len(node_form); i++ {
						if node_form[i].types == 1 { //若节点为留守节点则按last_regin进行分类，留守节点会进行临时跨区
							switch node_form[i].last_regin { //每个区域的节点集合
							case 1:
								toal_are_2[0].node = append(toal_are_2[0].node, node_form[i])
							case 2:
								toal_are_2[1].node = append(toal_are_2[1].node, node_form[i])
							case 3:
								toal_are_2[2].node = append(toal_are_2[2].node, node_form[i])
							case 4:
								toal_are_2[3].node = append(toal_are_2[3].node, node_form[i])
							case 5:
								toal_are_2[4].node = append(toal_are_2[4].node, node_form[i])
							case 6:
								toal_are_2[5].node = append(toal_are_2[5].node, node_form[i])
							case 7:
								toal_are_2[6].node = append(toal_are_2[6].node, node_form[i])
							case 8:
								toal_are_2[7].node = append(toal_are_2[7].node, node_form[i])
							case 9:
								toal_are_2[8].node = append(toal_are_2[8].node, node_form[i])
							}
						} else { //其他类型节点则按region进行分类
							switch node_form[i].region { //每个区域的节点集合
							case 1:
								toal_are_2[0].node = append(toal_are_2[0].node, node_form[i])
							case 2:
								toal_are_2[1].node = append(toal_are_2[1].node, node_form[i])
							case 3:
								toal_are_2[2].node = append(toal_are_2[2].node, node_form[i])
							case 4:
								toal_are_2[3].node = append(toal_are_2[3].node, node_form[i])
							case 5:
								toal_are_2[4].node = append(toal_are_2[4].node, node_form[i])
							case 6:
								toal_are_2[5].node = append(toal_are_2[5].node, node_form[i])
							case 7:
								toal_are_2[6].node = append(toal_are_2[6].node, node_form[i])
							case 8:
								toal_are_2[7].node = append(toal_are_2[7].node, node_form[i])
							case 9:
								toal_are_2[8].node = append(toal_are_2[8].node, node_form[i])
							}
						}
					}
					break //退出当前
				}
			}
		}
	}
}

// 定义函数Random_representative_election 实现第一次选举与周期选举
func Random_representative_election() {
	copy(voteNodesPool, node_form)                  //投票节点池为全部节点
	sort.Slice(voteNodesPool, func(i, j int) bool { //将投票节点按tokens从大到小排序
		return voteNodesPool[i].score > voteNodesPool[j].score
	})
	starNodesPool = voteNodesPool[:starNodes_number] //选择前starNodes_number节点作为竞选节点池

	voting() //发起投票获得超级节点
}

func Single_calculation(calculation_time int) {
	//节点初始化
	node_init()
	//节点运动
	node_move()
	//节点身份切换
	node_type_change()
	//创世区块添加
	make_genesisBlock()
	//区块交易产生
	make_block_message()
	//实现第一次代表选举
	Random_representative_election()
	for i := 0; i < block_number; i++ {
		begain := time.Now() //记录开始时间 每次完成一个区块共识重新计时
		obervation_cal(block_message[i].Bike)
		//指标计算 时效性，考虑到后边代码的编程效果影响这个因此指标
		result_tps[i], result_delay[i], resul_exchange[i] = Index_calculation(begain, i)
		//节点运动
		node_move()
		node_type_change() //实现节点身份切换
		Node_placement()   //节点投放 添加新生节点
		//每个区块上链打印
		fmt.Println("最新区块序号")
		spew.Dump(Blockchain[len(Blockchain)-1].Index)
		number := 0
		for k := 0; k < len(superStarNodesPool); k++ {
			if superStarNodesPool[k].id == false {
				number++
			}
		}
	}
	//打印结果
	printf_result(calculation_time)
	//下一次全局变量初始化
	Global_variable_initialization()
}

func main() {
	for j := 0; j < repeat_times; j++ {
		Single_calculation(j) //执行单次计算
	}
}

// 定义函数 Index_calculation 完成指标计算
func Index_calculation(begain time.Time, block_persent int) (float64, float64, float64) {
	tra_time = time.Now() //保存区块上链时间
	var temp_1 int
	if (superStarNode_number-1)/10 == 0 {
		temp_1 = 1
	} else {
		temp_1 = (superStarNode_number - 1) / 10
	}
	var temp_2 int
	if (superStarNode_number-1)/20 == 0 {
		temp_2 = 1
	} else {
		temp_2 = (superStarNode_number - 1) / 20
	}
	//指标计算
	//TPS吞吐量  定义：上链的总交易数/其区块生成时间  生成时间包括交易+共识机制执行+区块
	//result_time_2 := float64(tra_time.Sub(begain).Nanoseconds()) / 1000000000 //计算将交易量用于生成区块的时间 单位秒
	var result_time_2 float64 = 0
	if block_persent == block_number-1 {
		tokens_select_temp = tokens_select_temp + node_sum/30 //代表节点越区补充
	}
	//交易与区块进行模拟
	//其中,交易广播时间与交易数量有关, 视图切换广播与节点数量，交易数量有关 区块组装的过程时间与交易数量有关,
	// PRE-PREPARE消息广播与节点数量，交易数量有关, PREPARE消息广播与节点数量，交易数量有关,  COMMIT消息广播与节点数量，交易数量有关, 检查点协议与节点数量有关
	result_time_2 = result_time_2 + (float64(deal_amount)*0.01)*float64(safe_number_temp) + //交易广播
		(float64(temp_2)*float64(deal_amount)*0.01+float64(temp_1)*0.01)*float64(attatck_number_temp) + //视图切换广播
		(float64(deal_amount)*0.05)*float64(attatck_number_temp+safe_number_temp) + //其他组装区块
		float64(temp_2+temp_1+temp_1)*float64(deal_amount)*0.02*float64(safe_number_temp) + //通信三阶段
		((float64(node_sum-1)/10)+(float64(node_sum-1)/10)+(float64(node_sum-1)/10))*0.02*float64(attatck_number_temp+tokens_select_temp) // 投票广播+投票广播+代表广播
	if (len(Blockchain)-1)%check_point == 0 { //当上链区块数为check_point的倍数时执行检查点协议
		result_time_2 = result_time_2 + (float64(temp_1) * float64(deal_amount) * 0.03)
	}
	var result_tps float64
	result_tps = float64((deal_amount)) / result_time_2

	//delay 时间延迟 定义：交易发出到确认的时间 平均值
	// result_delay := float64(tra_time.Sub(begain).Nanoseconds()) / 1000000000 //计算将交易量用于生成区块的时间 单位秒
	var result_delay float64 = 0
	// 交易与区块进行模拟
	//其中,交易广播时间与交易数量有关, 视图切换广播与节点数量, 交易数量有关, 其他组装区块的过程时间与交易数量有关,检查点协议广播
	// PRE-PREPARE消息广播与节点数量，交易数量有关, PREPARE消息广播与节点数量有关,  COMMIT消息广播与节点数量有关
	result_delay = result_delay + (float64(deal_amount)*0.01)*float64(safe_number_temp) + //交易广播
		(float64(temp_2)*float64(deal_amount)*0.01+float64(temp_1)*0.01)*float64(attatck_number_temp) + //视图切换广播
		(float64(deal_amount)*0.05)*float64(attatck_number_temp+safe_number_temp) + //其他组装区块
		float64(temp_2+temp_1+temp_1)*float64(deal_amount)*0.02*float64(safe_number_temp) + //通信三阶段
		((float64(node_sum-1)/10)+(float64(node_sum-1)/10)+(float64(node_sum-1)/10))*0.02*float64(attatck_number_temp+tokens_select_temp) // 投票广播+投票广播+代表广播
	if (len(Blockchain)-1)%check_point == 0 { //当上链区块数为check_point的倍数时执行检查点协议
		result_delay = result_delay + (float64(temp_1) * float64(deal_amount) * 0.03)
	}

	//容错性  定义：保证共识安全的前提下，增加恶意节点数量，查看TPS与delay
	//见前面

	//通信代价 定义为节点间的通信次数  交易广播通信, 视图切换广播, PRE-PREPARE消息广播,   PREPARE消息广播,  COMMIT消息广播, 检查点协议广播
	resul_exchange := float64((deal_amount))*float64(safe_number_temp) + //交易广播
		float64(float64(superStarNode_number-1)*float64(deal_amount)+float64(superStarNode_number-1)*float64(superStarNode_number))*float64(attatck_number_temp) + //视图切换广播
		float64((superStarNode_number-1))*float64(safe_number_temp) + // PRE-PREPARE消息广播
		float64((superStarNode_number-1)*(superStarNode_number-1))*float64(safe_number_temp) + //PREPARE消息广播 主节点不参与
		float64((superStarNode_number-1)*(superStarNode_number))*float64(safe_number_temp) + //COMMIT消息广播
		float64((node_sum-1)*(node_sum)+(node_sum-1)*(node_sum)+(node_sum-1)*(node_sum))*float64(attatck_number_temp+tokens_select_temp) //tokens通信，投票广播，代表广播
	if (len(Blockchain)-1)%check_point == 0 { //当上链区块数为check_point的倍数时执行检查点协议广播 全部节点广播
		resul_exchange = resul_exchange + float64((node_sum-1)*(node_sum))
	}
	resul_exchange = resul_exchange / float64(node_sum) //平均通信开销

	attatck_number_temp = 0 //中间计数器-清零
	safe_number_temp = 0
	tokens_select_temp = 0
	return result_tps, result_delay, resul_exchange

}

// 定义函数node_voteing 节点投票结合score 针对全部竞选矿池
func node_voteing(v Node) int {
	var result int
	if v.id == true {
		lotteryPool := []string{} //验证者池持有所有验证者的地址，这些验证者都有机会成为一个胜利者

		for i := 0; i < len(starNodesPool); i++ {
			for j := 0; j < starNodesPool[i].score; j++ {
				lotteryPool = append(lotteryPool, starNodesPool[i].address) //权重体现
			}
		}
		var lotteryWinner string
		// 从验证者池中随机选择胜利者
		rand.Seed(time.Now().UnixNano())
		r := rand.Intn(len(lotteryPool)) //在[0.len(lotteryPool))的范围内产生随机数
		lotteryWinner = lotteryPool[r]   //随机选择胜利者 轮盘赌
		for i := 0; i < len(starNodesPool); i++ {
			if lotteryWinner == starNodesPool[i].address {
				result = i
				break
			}
		}
	} else {
		flag := 0
		for k := 0; k < len(starNodesPool); k++ { //投票给自己
			if starNodesPool[k].number == v.number {
				flag = 1
				result = k
			}
		}
		if flag == 0 {
			lotteryPool := []string{} //验证者池持有所有验证者的地址，这些验证者都有机会成为一个胜利者
			for i := 0; i < len(starNodesPool); i++ {
				for j := 0; j < starNodesPool[i].score; j++ {
					lotteryPool = append(lotteryPool, starNodesPool[i].address) //权重体现
				}
			}
			var lotteryWinner string
			// 从验证者池中随机选择胜利者
			rand.Seed(time.Now().UnixNano())
			r := rand.Intn(len(lotteryPool)) //在[0.len(lotteryPool))的范围内产生随机数
			lotteryWinner = lotteryPool[r]   //随机选择胜利者 轮盘赌
			for i := 0; i < len(starNodesPool); i++ {
				if lotteryWinner == starNodesPool[i].address {
					result = i
					break
				}
			}
		}
	}

	return result //返回
}

// 定义函数voting 第一次投票选择超级节点  所有节点都有投票权
func voting() {
	for _, v := range voteNodesPool { //获取投票节点池中的节点
		r := node_voteing(v)             //在[0.superStarNode_number)
		starNodesPool[r].vote += v.score //投票
	}
	sort.Slice(starNodesPool, func(i, j int) bool {
		return starNodesPool[i].vote > starNodesPool[j].vote //按选票进行从大到小排序
	})
	superStarNodesPool = starNodesPool[:superStarNode_number] //选择前面几个作为超级节点池

	for i := 0; i < node_sum; i++ { //在原始节点中按超级节点改变等级  便于后期分类
		for j := 0; j < len(superStarNodesPool); j++ {
			if node_form[i].number == superStarNodesPool[j].number {
				node_form[i].rank = 2
				break
			}
		}
	}
}

// 定义函数generateBlock 生成新区块
func generateBlock(oldBlock Block, Bike []int, address string, miner int) Block {
	var newBlock Block //新区块

	t := time.Now()                     //获取当前时间
	newBlock.Index = oldBlock.Index + 1 //区块的增加，index也加一
	newBlock.Timestamp = t.String()     //时间戳
	newBlock.Bike = Bike                //一定区域内的自行车数量
	newBlock.PrevHash = oldBlock.Hash   //新区块的PrevHash存储上一个区块的Hash
	newBlock.Validator_key = address    //验证者地址

	for i := 0; ; i++ { //通过循环改变 Nonce
		hex := fmt.Sprintf("%x", i) //获得字符串形式 Nonce
		newBlock.Nonce = hex        //Nonce赋值给newBlock.Nonce
		// 判断Hash的前导0个数，是否与难度系数一致
		if !isHashValid(calculateBlockHash(newBlock), difficulty-adv_difficulty) { //difficulty-adv_difficulty 表示超级节点会在计算哈希时降低难度值
			fmt.Println(calculateBlockHash(newBlock), " do more work!") //继续挖矿中
			time.Sleep(time.Second)                                     //暂停一秒
			continue
		} else {
			//fmt.Println(calculateBlockHash(newBlock), " work done!") //挖矿成功
			newBlock.Hash = calculateBlockHash(newBlock) //重新计算hash并赋值
			break
		}
	}
	return newBlock //返回新区快
}

// 定义函数isHashValid 验证Hash的前导0个数是否与困难度系数一致
func isHashValid(hash string, difficulty int) bool {
	// 返回difficulty个0串联的字符串prefix
	prefix := strings.Repeat("0", difficulty)
	//若hash是以prefix开头，则返回true，否则返回false
	return strings.HasPrefix(hash, prefix)
}

// 定义函数calculateHash 生成Hash值
func calculateHash(s string) string {
	h := sha256.New()                 //创建一个基于SHA256算法的hash.Hash接口的对象
	h.Write([]byte(s))                //输入数据
	hashed := h.Sum(nil)              //计算哈希值
	return hex.EncodeToString(hashed) //将字符串编码为16进制格式,返回字符串
}

// 定义函数calculateBlockHash 生成区块Hash值
func calculateBlockHash(block Block) string {
	var temp string
	for i := 0; i < len(block.Bike); i++ {
		temp = temp + strconv.Itoa(block.Bike[i])
	}
	record := string(block.Index) + block.Timestamp + temp + block.PrevHash //区块内其他信息作为字符串   strconv.Itoa 整形变字符串
	return calculateHash(record)                                            //调用函数calculateHash 生成Hash值
}

// 定义函数isPRE_PREPARE_Valid 处理PRE_PREPARE消息  (主节点不处理PRE_PREPARE消息)
func isPRE_PREPARE_Valid(newBlock Block, i int, miner_number_temp int) {
	//确认Index的增长正确  确认密钥与主节点相同  确认摘要计算正确
	if node_form[i].temp_block.Index+1 == newBlock.Index &&
		newBlock.Validator_key == node_form[miner_number_temp].Validator_key &&
		calculateBlockHash(newBlock) == newBlock.Hash {
		node_form[i].temp_block = newBlock //保存到临时区块并产生Prepare消息
		Prepare_noise[i] = Prepare{node_form[i].temp_block.Hash, node_form[i].temp_block.Index, node_form[i].number, node_form[i].Validator_key}
	}
}

/*Prepare_noise_number int
Commit_noise_number int*/
//定义函数isPREPARE_Valid 处理PREPARE消息  (主节点没有产生PREPARE消息)
func isPREPARE_Valid(i int, Prepare_noise []Prepare, miner_number_temp int) {
	for k := 0; k < node_sum; k++ {
		var temp_Validator_key string
		for j := 0; j < len(node_form); j++ {
			if node_form[j].number == Prepare_noise[k].NodeID {
				temp_Validator_key = node_form[j].Validator_key //获取身份地址
				break
			}
		}
		//确认Index的增长正确  确认密钥与来源节点相同  确认摘要计算正确
		if node_form[i].temp_block.Index == Prepare_noise[k].Index &&
			temp_Validator_key == Prepare_noise[k].Validator_key &&
			node_form[i].temp_block.Hash == Prepare_noise[k].Hash {
			node_form[i].Prepare_noise_number++                //认可加1
			if node_form[i].Prepare_noise_number >= node_sum { //判断是否大于2f
				//产生COMMIT消息 计算摘要  序号 节点号 重新使用节点密钥
				Commit_noise[i] = Commit{node_form[i].temp_block.Hash, node_form[i].temp_block.Index, node_form[i].number, node_form[i].Validator_key}
				node_form[i].Prepare_noise_number = 0
			}
		}
	}
}

// 定义函数isCOMMIT_Valid 处理COMMIT消息
func isCOMMIT_Valid(i int, Commit_noise []Commit) {
	//确认Index的增长正确  确认密钥与来源节点相同  确认摘要计算正确
	for k := 0; k < node_sum; k++ {
		var temp_Validator_key string
		for j := 0; j < len(node_form); j++ {
			if node_form[j].number == Commit_noise[k].NodeID {
				temp_Validator_key = node_form[j].Validator_key //获取身份地址
				break
			}
		}
		if node_form[i].temp_block.Index == Commit_noise[k].Index &&
			temp_Validator_key == Commit_noise[k].Validator_key &&
			node_form[i].temp_block.Hash == Commit_noise[k].Hash {
			node_form[i].Commit_noise_number++        //认可加1
			if node_form[i].Commit_noise_number > 0 { //判断节点是否大于2f
				node_form[i].Commit_noise_number = 0
			}
		}
	}
}

// 定义函数test 循环发送交易
func test(message []int) { //操作函数里同时包含写和读 需要分开写
LOOP1:
	var miner int
	miner_Validator_key := superStarNodesPool[miner_number].Validator_key
	miner_number_temp := miner_number //用于标注主节点，便于后面验证
	miner_number = miner_number + 1   //视图切换 与恶意节点分布有关
	if miner_number >= len(superStarNodesPool) {
		miner_number = 0
	}
	for i := 0; i < len(superStarNodesPool); i++ {
		if miner_Validator_key == superStarNodesPool[i].Validator_key {
			miner = i //根据验证者的地址确定其序号(保存格式为切片，从0开始)
			break
		}
	}

	if superStarNodesPool[miner].id == false {
		attatck_number = attatck_number + 1 //恶意攻击成功次数加1
		attatck_number_temp = attatck_number_temp + 1
		newselecting(miner) //若存在恶意节点，则重新选举 添加新的节点作为超级节点
		goto LOOP1
	}

	newBlock := generateBlock(Blockchain[len(Blockchain)-1], message, miner_Validator_key, miner) //产生区块  PRE-PREPARE消息
	node_form[miner_number_temp].temp_block = newBlock                                            //主节点自行保存PRE-PREPARE消息

	for i := 0; i < node_sum; i++ { //主节点不需要处理PRE-PREPARE消息
		isPRE_PREPARE_Valid(newBlock, i, miner_number_temp) //处理PRE-PREPARE消息保存到临时区块  并生成PREPARE消息
	}

	for i := 0; i < node_sum; i++ { //主节点没有产生PREPARE消息
		isPREPARE_Valid(i, Prepare_noise, miner_number_temp) //验证PREPARE消息 并生成COMMIT消息
	}

	for i := 0; i < node_sum; i++ {
		isCOMMIT_Valid(i, Commit_noise) //验证COMMIT消息
	}

	for j := 0; j < node_sum; j++ {
		if node_form[j].number == superStarNodesPool[miner_number].number { //奖励
			node_form[j].score = node_form[j].score + 30
		}
	}

	Blockchain = append(Blockchain, newBlock) //追加区块

	for j := 0; j < len(node_form); j++ {
		node_form[j].block = append(node_form[j].block, newBlock) //其他节点同步
	}
	safe_number = safe_number + 1 //共识成功次数加1
	safe_number_temp = safe_number_temp + 1

	if (len(Blockchain)-1)%tokens_day == 0 && (len(Blockchain)-1) != 0 { //周期选举
		for j := 0; j < len(node_form); j++ {
			node_form[j].rank = 1 //节点等级回到投票节点
		}
		Random_representative_election()            //重新选举
		tokens_select_temp = tokens_select_temp + 1 //加1
	}
}

// 定义函数generateRandomNumber 产生count个[start,end)结束的不重复的随机数
func generateRandomNumber(start int, end int, count int) []int {
	//范围检查
	if end < start || (end-start) < count {
		return nil
	}
	//存放结果的slice
	nums := make([]int, 0)
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		//生成随机数
		num := r.Intn((end - start)) + start
		//查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}
		if !exist {
			nums = append(nums, num)
		}
	}
	return nums
}

// 定义函数newselecting 剔除恶意节点并重新投票选择超级节点  投票节点才有投票权
func newselecting(k int) {
	var voteNodesPool_2 = make([]Node, (node_sum - superStarNode_number + 1), (node_sum - superStarNode_number + 1)) //产生中间投票节点池
	var starNodesPool_2 = make([]Node, starNodes_number, starNodes_number)                                           //产生中间竞选节点池
	for i := 0; i < node_sum; i++ {
		if node_form[i].number == superStarNodesPool[k].number {
			node_form[k].rank = 1 //降级并扣除
			node_form[k].score = node_form[k].score - 30
		}
	}

	voteNodesPool_temp := 0
	for i := 0; i < node_sum; i++ { //将投票节点与超级节点分类
		if node_form[i].rank == 1 {
			voteNodesPool_2[voteNodesPool_temp] = node_form[i]
			voteNodesPool_temp++
		}
	}

	sort.Slice(voteNodesPool_2, func(i, j int) bool { //投票节点按tokens从大到小排序，并选择前starNodes_number个作为竞选节点池
		return voteNodesPool_2[i].score > voteNodesPool_2[j].score
	})
	starNodesPool_2 = voteNodesPool_2[:starNodes_number]

	for _, v := range voteNodesPool_2 { //获取投票节点池中的节点  轮盘赌
		r := node_voteing_2(v, starNodesPool_2) //选择需要投票数
		starNodesPool_2[r].vote += v.score      //投票
	}

	sort.Slice(starNodesPool_2, func(i, j int) bool {
		return starNodesPool_2[i].vote > starNodesPool_2[j].vote //竞选节点池按选票进行从大到小排序
	})
	superStarNodesPool[k] = starNodesPool_2[0] //选择前面1个作为超级节点替换恶意节点

	for i := 0; i < node_sum; i++ { //在原始节点按超级节点(新节点)改变等级
		if node_form[i].number == superStarNodesPool[k].number {
			node_form[i].rank = 2
		}
	}
}

// 定义函数node_voteing_2 节点投票结合节点tokens 针对剔除恶意节点时的投票
func node_voteing_2(v Node, starNodesPool_2 []Node) int {
	var result int
	if v.id == true {
		lotteryPool := []string{} //验证者池持有所有验证者的地址，这些验证者都有机会成为一个胜利者

		for i := 0; i < len(starNodesPool_2); i++ {
			for j := 0; j < starNodesPool_2[i].score; j++ {
				lotteryPool = append(lotteryPool, starNodesPool_2[i].address) //权重体现
			}
		}

		var lotteryWinner string
		// 从验证者池中随机选择胜利者
		rand.Seed(time.Now().UnixNano())
		r := rand.Intn(len(lotteryPool)) //在[0.len(lotteryPool))的范围内产生随机数
		lotteryWinner = lotteryPool[r]   //随机选择胜利者 轮盘赌
		for i := 0; i < len(starNodesPool_2); i++ {
			if lotteryWinner == starNodesPool_2[i].address {
				result = i
				break
			}
		}
	} else {
		flag := 0
		for k := 0; k < len(starNodesPool_2); k++ { //投票给自己
			if starNodesPool_2[k].number == v.number {
				flag = 1
				result = k

			}
		}
		if flag == 0 {
			lotteryPool := []string{} //验证者池持有所有验证者的地址，这些验证者都有机会成为一个胜利者
			for i := 0; i < len(starNodesPool_2); i++ {
				for j := 0; j < starNodesPool_2[i].score; j++ {
					lotteryPool = append(lotteryPool, starNodesPool_2[i].address) //权重体现
				}
			}

			var lotteryWinner string
			// 从验证者池中随机选择胜利者
			rand.Seed(time.Now().UnixNano())
			r := rand.Intn(len(lotteryPool)) //在[0.len(lotteryPool))的范围内产生随机数
			lotteryWinner = lotteryPool[r]   //随机选择胜利者 轮盘赌
			for i := 0; i < len(starNodesPool_2); i++ {
				if lotteryWinner == starNodesPool_2[i].address {
					result = i
					break
				}
			}
		}
	}

	return result //返回
}

// 定义函数 Determine_location 生成节点所属区域编号
func Determine_location(temp_x int, temp_y int) int {
	if temp_x >= 0 && temp_x < 300 {
		if temp_y >= 0 && temp_y < 300 {
			return 1
		} else if temp_y >= 300 && temp_y < 600 {
			return 4
		} else {
			return 7
		}
	} else if temp_x >= 300 && temp_x < 600 {
		if temp_y >= 0 && temp_y < 300 {
			return 2
		} else if temp_y >= 300 && temp_y < 600 {
			return 5
		} else {
			return 8
		}
	} else {
		if temp_y >= 0 && temp_y < 300 {
			return 3
		} else if temp_y >= 300 && temp_y < 600 {
			return 6
		} else {
			return 9
		}
	}
}

// 定义函数printf_result 实现结果打印与保存
func printf_result(calculation_time int) {
	//计算平均结果
	var result_tps_avg float64 = 0
	var result_delay_avg float64 = 0
	var resul_exchange_avg float64 = 0
	for i := start_cal; i < block_number; i++ {
		result_tps_avg = result_tps_avg + result_tps[i]
	}
	if start_cal > 0 {
		result_tps_avg = result_tps_avg / (block_number - float64(start_cal) - 1)
	} else {
		result_tps_avg = result_tps_avg / block_number
	}
	result_tps_avg = result_tps_avg * 100
	for i := start_cal; i < block_number; i++ {
		result_delay_avg = result_delay_avg + result_delay[i]
	}
	if start_cal > 0 {
		result_delay_avg = result_delay_avg / (block_number - float64(start_cal) - 1)
	} else {
		result_delay_avg = result_delay_avg / block_number / 1.1
	}
	result_delay_avg = result_delay_avg * (block_number * 0.025) / (math.Sqrt(float64(block_number) / 400))
	for i := start_cal; i < block_number; i++ {
		resul_exchange_avg = resul_exchange_avg + resul_exchange[i]
	}
	if start_cal > 0 {
		resul_exchange_avg = resul_exchange_avg / (block_number - float64(start_cal) - 1)
	} else {
		resul_exchange_avg = resul_exchange_avg / block_number
	}
	resul_exchange_avg = resul_exchange_avg * (block_number * 0.005)
	//打印
	fmt.Println("最终平均结果：")
	fmt.Println("TPS：")
	fmt.Println(result_tps_avg)

	fmt.Println("Delay：")
	fmt.Println(result_delay_avg)

	fmt.Println("Exchange：")
	fmt.Println(resul_exchange_avg)

	fmt.Println("恶意攻击成功次数")
	fmt.Println(attatck_number)
	fmt.Println("共识成功次数")
	fmt.Println(safe_number)

	//结果保存
	filename := "result.txt"
	var f *os.File
	var err1 error
	_, err1 = os.Stat(filename)
	if err1 != nil {
		// 错误不为空，表示文件不存在
		f, err1 = os.Create(filename) //创建文件
	} else {
		// 错误为空，表示文件存在
		f, err1 = os.OpenFile(filename, os.O_APPEND, 0644) //打开文件
	}
	if err1 != nil {
		panic("文件创建或打开失败")
	}
	writer := bufio.NewWriter(f)
	writer.WriteString("重复运行次数：") //指标标志
	writer.WriteString(strconv.Itoa(calculation_time + 1))
	writer.WriteString("\n")

	writer.WriteString("TSP") //写入TSP
	writer.WriteString("\n")
	writer.WriteString(strconv.FormatFloat(result_tps_avg, 'E', -1, 64))
	writer.WriteString("\n")
	writer.Flush()

	writer.WriteString("Delay:") //写入Delay
	writer.WriteString("\n")
	writer.WriteString(strconv.FormatFloat(result_delay_avg, 'E', -1, 64))
	writer.WriteString("\n")
	writer.Flush()

	writer.WriteString("Exchange:") //写入Exchange
	writer.WriteString("\n")
	writer.WriteString(strconv.FormatFloat(resul_exchange_avg, 'E', -1, 64))
	writer.WriteString("\n")
	writer.Flush()

	/*		writer.WriteString("完整TSP") //写入完整TSP
			writer.WriteString("\n")
			for k:=0;k<block_number;k++{
				writer.WriteString(strconv.FormatFloat(result_tps[k], 'E', -1, 64))
				writer.WriteString("\n")
				writer.Flush()
			}


			writer.WriteString("完整Delay:") //写入完整Delay
			writer.WriteString("\n")
			for k:=0;k<block_number;k++{
				writer.WriteString(strconv.FormatFloat(result_delay[k], 'E', -1, 64))
				writer.WriteString("\n")
				writer.Flush()
			}


			writer.WriteString("完整Exchange:") //写入完整Exchange
			writer.WriteString("\n")
			for k:=0;k<block_number;k++ {
				writer.WriteString(strconv.FormatFloat(resul_exchange[k], 'E', -1, 64))
				writer.WriteString("\n")
				writer.Flush()
			}*/
}

// 定义函数Global_variable_initialization 完成下一次全局变量初始化
func Global_variable_initialization() {
	node_sum = FN_number + RN_number + SN_number + NN_number                              //节点
	Prepare_noise = make([]Prepare, node_sum, node_sum)                                   //预准备消息
	Commit_noise = make([]Commit, node_sum, node_sum)                                     //预确认消息
	Blockchain = Blockchain[0:0]                                                          // 存放区块数据
	node_form = make([]Node, node_sum, node_sum)                                          //初始化节点信息
	block_message = make([]Message, block_number, block_number)                           //区块交易集合
	miner_number = 0                                                                      //区块生成序号
	attatck_number = 0                                                                    //恶意攻击成功次数
	attatck_number_temp = 0                                                               //观测点之间的恶意攻击次数
	safe_number = 0                                                                       //共识成功次数
	safe_number_temp = 0                                                                  //观测点之间的共识成功次数
	tokens_select_temp = 0                                                                //选举次数统计
	node_form_attack_address = make([]Node, bad_number_id_address, bad_number_id_address) //恶意克隆节点数量
	th_1 = 2                                                                              //流窜节点向留守节点转换阈值  区块数
	th_2 = 2                                                                              //留守节点向流窜节点转换阈值  区块数
	result_tps = make([]float64, block_number, block_number)                              //每次区块共识的TPS
	result_delay = make([]float64, block_number, block_number)                            //每次区块共识的延迟
	resul_exchange = make([]float64, block_number, block_number)                          //每次区块共识的通信次数
	voteNodesPool = make([]Node, node_sum, node_sum)                                      //投票节点池
	starNodesPool = make([]Node, starNodes_number, starNodes_number)                      //竞选节点池
	superStarNodesPool = make([]Node, superStarNode_number, superStarNode_number)         //超级节点池
}
