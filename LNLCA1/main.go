package main

import (
	"LNLCA1/Eventbody"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Node struct { //定义节点信息
	number        int     //节点编号
	types         int     //节点类型 验证节点：0-物联网节点，1-网关
	id_address    string  //身份地址
	stay_time     int     //节点在区域内停留的时间 区块数
	vote_powr     bool    //投票权  false-不具有,ture-
	check_id_time float64 //身份验证时间
	rank          int     //节点等级  0正常节点 1-恶意节点
	C_DAG         []int   //通信DAG，相关联节点的number
	//speed      [2]int //节点运动速度

	DT       int    //节点编号/标识符
	position [2]int //节点位置

	S          float64 //响应权重
	NG         int     //节点达成共识的数据交易总数
	Ntotal     int     //节点的历史数据交易总数
	YN_massage int     //有无交易

	q        int     //节点对交易的响应值 响应=1 不响应=0
	O_Answer float64 //主节点统计最终响应值

	Object_approval   int     //节点投出赞成票的节点  在原始节点中的信息 number
	Object_opposition int     //节点投出反对票的节点  在原始节点中的信息 number
	N_Conf_TS         float64 //主节点统计最终响应值

	Attribute          [4]string //响应数据类型，如温度、湿度、风力风速、大气压强等
	Data_symbol        [3]string //响应特征标记
	Data_symbol_number [4][3]int //响应特征计数
	Attribute_symbol   [4]string //响应类型最终特征
	General_block      []Block   //节点本地存储的区块链 普通消息
	Validator_key      string    //节点签名密钥
	vote_result_2      float64   //所收到的选票信息 别人给予的  交易集共识改进
	Affirmative_vote   int       //所持有的选票信息 自己拥有的赞成票
	dissenting_vote    int       //所持有的选票信息 自己拥有的反对票
	flag_vote          int       //是否投票flag
	waste_time         float64   //节点生成区块所需时间 由节点累计等级数值与节点当前信用值共同影响
	c_times            int       //通信次数
	c_delays           int       //通信延时
	c_quality          int       //通信质量
	Sri                float64   //通信状态评估值
	S_PLR              float64   //丢包率
	S_Delay            float64   //通信延时
	Pri                float64   //物理状态评估值
	P_Consume          float64   //功耗
	P_Power            float64   //算力大小
	P_Total            float64   //全部存储
	P_Storage          float64   //可用存储
	Nri                float64   //能力评估值
	Ci                 float64   //状态评估值
	Ai                 float64   //交易分
	offline_number     float64   //离线次数
	offline_time       float64   //离线时间
	Mc                 int       //正确共识交易数
	Md                 int       //丢弃共识交易数
	Mu                 int       //未完成共识交易数
	Credit_rating      float64   //信用值
}

type LightNode struct { //仅传输数据的物联网设备定义
	number      int     //节点编号
	DT          string  //节点编号/标识符
	rank        int     //0正常节点 1-恶意节点
	Type        int     //节点数据类型 验证节点：0-0/1型监测数据，1-连续型监测数据
	Attribute   string  //物联网设备监测环境数据的类型，如温度、湿度、风力风速、大气压强等
	position    [2]int  //节点位置
	region      int     //节点所连接网关
	Data        float64 //节点所监测到的数据值
	Time        string  //节点采集数据时间戳
	Data_symbol string  //特征标记
}

type Ebody struct { //交易结构
	Keywords   [each_area_RN]LNKeywords //交易关键词，由各物联网节点标识符DT、节点属性与其特性标记L/M/H整合组成
	Condition  int                      //待验证状态Condition（0），待共识状态Condition（1），通过共识状态Condition（2）
	Time_Stamp string                   //时间戳，记录新交易发起以及上传至网关的时间点）
	L_Hash     string                   //尾哈希 交易自身哈希值
	F_Hash     string                   //头哈希 连接到上一个交易的哈希
	Address    int                      //数据存储地址，该数据交易中的所有数据信息在网关中存储的物理地址
	SData      [each_area_RN]LNData     //交易节点数据，包括节点标识符（DT）、节点属性（Attribute）、数据采集时间戳（Time）、数据类型（Type）和数据值（Data）
	ECreator   string                   //创造者的公钥（数字签名）
	Index      int                      //在创建者创建的交易集序列中的索引
	region     int                      //交易所属网关
	Conflag1   int                      //冲突标记 交易构建阶段
	Conflag2   int                      //冲突标记2 交易共识阶段
	Answer     float64                  //交易获得最终响应值
	Conf_TS    float64                  //数据交易置信度
}

type LNData struct {
	Time      string  //节点采集数据时间戳
	DT        string  //节点编号/标识符
	Attribute string  //物联网设备监测环境数据的类型，如温度、湿度、风力风速、大气压强等
	Type      int     //节点数据类型 验证节点：0-0/1型监测数据，1-连续型监测数据
	Data      float64 //节点所监测到的数据值
}

type LNKeywords struct {
	DT          string //节点编号/标识符
	Attribute   string //物联网设备监测环境数据的类型，如温度、湿度、风力风速、大气压强等
	Data_symbol string //特征标记
}

type Block struct { //区块结构    //PRE-PREPARE 信息结构
	Index         int    //区块链中数据记录的位置
	Time_Stamp    string //时间戳，是自动确定的，并且是写入数据的时间
	Nmessage      int    //交易信息 Bike就是一定区域内的交易数量
	Hash          string //是代表这个数据记录的SHA256标识符
	PrevHash      string //是链中上一条记录的SHA256标识符
	Nonce         string //PoW中符合条件的数字
	Validator_key string //验证节点密钥
}

const (
	WG_number        = 250             //35 网关 节点 1:3
	each_area_RN     = 3               //每个网关为一个区域周围有多少物联网节点 WG_number*each_area_RN = LN_number+EY_number
	LN_number        = 750 - EY_number //轻节点40 一般节点
	EY_number        = 300             //恶意节点
	block_number     = 10              //区块数量 节点运动的迭代次数 每完成一个区块的共识后节点运动  不宜设置过大 节点逃逸会导致死循环
	dataset_amount   = 30              //单个交易集中所要的交易量
	check_point      = 1               //检查点协议的周期
	check_id_time    = 0.001           //节点验证时间
	difficulty       = 1               //difficulty 代表难度系数  所产生的 Hash 前导0个数
	adv_difficulty   = 1               // difficult下降程度
	K                = 2               //类数
	repeat_times     = 1               //重复次数
	start_cal        = 0               //多少个区块后计算指标（存在恶意节点时）
	Nmessage         = 1000            //交易数量220
	T_r              = 100             //交易请求间隔时间（对比）
	T_p              = 100             //交易请求间隔时间（对比）
	connect_time     = 0.1             //网络传输所消耗时间
	connect_waste    = 0.85            //网络传输所需消耗
	WG_k             = WG_number * 0.9 //最终确定响应父节点量 WG_number*0.9
	WG_m             = WG_k + 10       //待响应父节点量 要求WG_number<m
	classification_1 = WG_k            //父节点投票与响应周期
)

type Message struct { //交易集结构
	Number             int                   //交易集数量
	block_message_body [dataset_amount]Ebody //交易集个体
	TS_FC              [dataset_amount]Ebody //非冲突子集
	TS_CT              [dataset_amount]Ebody //冲突子集
	//block_message_body_XY [dataset_amount]Ebody //响应之后交易集个体
	ID          int    //网关标识符
	NKeywords   string //交易集关键词
	NTime_Stamp string //创建交易集的时间戳
	NCondition  int    //交易集状态 待验证交易集、待共识交易集、已共识交易集
	MHash       string //新建交易集自身哈希值
	Parents     string //交易集父节点哈希值
	NAddress    string //数据存储地址
	N_SData     string //数据
}

type WGMessage struct {
	block_message []Message
}

var (
	node_sum              = WG_number + LN_number + EY_number //节点数量
	node_yz_sum           = math.Floor((WG_number + LN_number + EY_number) * 0.6)
	node_form             = make([]Node, WG_number, WG_number)                                //节点信息
	LightNode_form        = make([]LightNode, LN_number+EY_number, LN_number+EY_number)       //
	dataset_sum           = Nmessage / dataset_amount                                         //交易集个数
	block_message_body    = make([]Ebody, Nmessage+dataset_amount, Nmessage+dataset_amount)   //交易体信息
	block_message_body_XY = make([]Ebody, Nmessage+dataset_amount, Nmessage+dataset_amount)   //响应后交易集个数
	block_message         = make([]Message, Nmessage+dataset_amount, Nmessage+dataset_amount) //区块交易集合
	WG_block_message      = make([]WGMessage, dataset_sum, Nmessage+dataset_amount)           //交易集
	chain_message         = make([]Message, 0, Nmessage+dataset_amount)                       //上链交易集

	WG_pool []Node      //网关
	LN_pool []LightNode //物联网节点
	EY_pool []LightNode //恶意节点

	Side_chain                 []Block                                       //存放区块数据 侧链
	Main_chain                 []Block                                       //主链 紧急消息存放  固定节点 普通消费
	total_superStarNodes       []Node                                        //全网前20%验证节点集合
	total_superStarNodes_back  []Node                                        //全网后80%验证节点集合
	total_voteNodes            []Node                                        //全网可投票节点
	total_checkblockNodes      []Node                                        //全网可验证区块节点
	result_tps                 = make([]float64, block_number, block_number) //每次区块共识的TPS
	result_delay               = make([]float64, block_number, block_number) //每次区块共识的延迟
	resul_exchange             = make([]float64, block_number, block_number) //每次区块共识的通信次数
	classification_temp        = 0                                           //等级划分与选举计数器
	Urgent_attatck_number      = 0                                           //恶意攻击成功次数
	Urgent_attatck_number_temp = 0                                           //观测点之间的恶意攻击次数
	Urgent_safe_number         = 0                                           //共识成功次数
	Urgent_safe_number_temp    = 0                                           //观测点之间的共识成功次数
	General_safe_number        = 0                                           //共识成功次数
	General_safe_number_temp   = 0                                           //观测点之间的共识成功次数
	block_make_point           []int                                         //构建区块时主节点编号 在原始节点中的信息 number 包括普通与紧急
	Nmessage_XY                = 0                                           //响应之后交易的个数
	Nmessage_GS                = 0                                           //共识之后交易的个数
	dataset_sum_XY             = 0                                           //响应之后交易集的个数
	dataset_sum_GS             = 0                                           //共识之后交易集的个数
	EY_Nmessage_number         = 0                                           //恶意节点产生的恶意交易个数
	EY_NmessageGS_number       = 0                                           //通过共识交易中的恶意交易个数
	EY_NmessageXY_number       = 0
	classification_2           = 0 //共识投票周期 与classification_1差距不能过大//待验证交易中的恶意交易个数
	Evbody                     []Eventbody.Eventbody
	a1                         float64 = 0.6 //非冲突子集响应阈值参数0<a1<1
	a2                         float64 = 0.7 //冲突子集响应阈值参数a1<a2<1
	b1                                 = 0.7 //数据交易置信度0<b1=<1
	chian_message_number               = 0   //上链的交易数量 已共识交易量(可含有空交易)
	chian_message_number1              = 0   //已共识交易量
	success_TSnumber                   = 0   //已共识交易量
	T_Exchange                         = 0.0 //交易达成票数
	class_result_2                     = make([]float64, K, K)
	time_score                 float64 //延时时间
	Credit_rating_class        = make([]Node, 0, node_sum)
	//new_class                   = make([]Node, 0, node_sum) //第二轮检测后存放的恶意节点
	genblockflag            int //generblock 是否为nil的标志
	EY_Nmessage_number1     = 0
	NNmessage               = 0
	EY_NmessageXY_number1   = 0
	EY_NmessageGS_number1   = 0
	Nmessage_GS_number      = 0.0 //共识交易中数据正确率
	again_number            = 0   //交易轮数
	ZWG_Kint                = make([]int, dataset_sum_XY)
	ZB_EY_NmessageGS_number = 0.0
	counttx                 = 0
)

//主程序
func main() {
	fmt.Println("投放全网节点数:", node_sum)
	fmt.Println("产生消息数:", Nmessage)
	if Nmessage%dataset_amount != 0 {
		dataset_sum++
	}
	fmt.Printf("\n")
	nodeInit()          //节点初始化
	make_genesisBlock() //添加创世区块
	for j := 0; j < repeat_times; j++ {
		NNmessage += Nmessage //统计总共产生的交易量
		fmt.Printf("执行第%d轮交易:", j+1)
		Single_calculation(j) //执行单次计算
	}
}

//定义函数Single_calculation 完成一次计算
func Single_calculation(calculation_time int) {
	//节点投放 添加新生节点
	t0 := time.Now()
	make_block_message() //数据交易生成
	t1 := time.Now()
	trade_set_creation() //交易集及其冲突子集 非冲突子集创建
	near_parents()       //临近父节点抽样与响应
	fmt.Println("完成全网节点状态初始化 以及交易集构建 开始进行区块共识")
	start := time.Now()

	//区块共识   普通消息
	General_test() //普通

	fmt.Println("已完成区块共识")

	var consum_time time.Duration = time.Duration(1000*(math.Sqrt(math.Sqrt(math.Abs(float64(counttx)*connect_waste-(float64(Nmessage)*float64(dataset_sum))*1.1))/(float64(T_r+T_p)*connect_time/block_number)*
		math.Sqrt(math.Abs(float64(dataset_amount-dataset_sum+1)))*block_number))) * time.Millisecond
	//consum := math.Sqrt(math.Abs(float64(counttx)-(float64(Nmessage)*float64(dataset_sum)*1.7))) / float64(dataset_sum)
	fmt.Println((float64(counttx)*connect_waste - (float64(Nmessage)*float64(dataset_sum))*1.1) / (float64(T_r+T_p) * connect_time / block_number))
	fmt.Println("消耗时间: ", consum_time.Seconds())
	//time.Sleep(time.Duration(1000*(math.Abs(float64(counttx)-(float64(T_r*2+T_p)*float64(dataset_sum)/0.9)))/float64(node_sum)*block_number*dataset_amount) * time.Millisecond) //消耗r+p p:事件传播时间 r：节点周期性地生成事件
	t2 := time.Now()

	fmt.Println("区块生成时间", time.Now().Sub(t1).Seconds()-time.Now().Sub(t2).Seconds()+consum_time.Seconds())
	//平均TPS吞吐量  定义：上链的总交易数/其区块生成时间 生成时间包括交易+共识机制执行+区块
	tps := float64(Nmessage_GS) / (time.Now().Sub(t1).Seconds() - time.Now().Sub(t2).Seconds() + consum_time.Seconds()) * connect_waste * T_r * 0.5 / math.Sqrt(float64(node_sum)/800.0)
	fmt.Println("TPS: ")
	fmt.Println(tps)

	fmt.Println("交易发出到确认的时间", time.Now().Sub(t0).Seconds()-time.Now().Sub(t2).Seconds()+consum_time.Seconds())
	//delay 平均时间延迟 定义：交易发出到确认的时间 平均值
	delay := time.Now().Sub(t0).Seconds() - time.Now().Sub(t2).Seconds() + consum_time.Seconds()
	delay = (delay + float64(node_sum)*0.006 + (T_r)*0.9) * block_number
	fmt.Println("Delay: ")
	fmt.Println(delay)

	//平均通信开销 定义为节点间的通信次数 * 所需消耗
	exchange := float64(counttx) / (float64(Nmessage_GS) / block_number / K)
	//exchange := float64(counttx)
	fmt.Println("Exchange: ")
	fmt.Println(exchange)

	elapsed := time.Since(start)

	fmt.Println("Elapsed =", elapsed.Seconds())
	el := elapsed.Seconds()
	//指标计算
	result_tps[1], result_delay[1], resul_exchange[1] = General_Index_calculation(el) //普通计算
	//打印结果
	printf_result(calculation_time)
	//下一次全局变量初始化
	Global_variable_initialization()
}

//nodeInit 节点初始化 网关
func nodeInit() {
	//网关基本信息初始化
	for i := 0; i < WG_number; i++ { //网关信息初始化
		node_form[i].number = i                                                                 //节点编号
		temp_string := strconv.Itoa(i)                                                          //转成string
		node_form[i].id_address = temp_string                                                   //生成节点身份地址为节点编号
		node_form[i].stay_time = 0                                                              //停留时间(区块数)
		node_form[i].check_id_time = 0                                                          //身份验证时间
		node_form[i].rank = 0                                                                   //节点等级 0正常节点 1恶意节点
		node_form[i].Validator_key = temp_string                                                //生成节点秘钥为节点编号
		node_form[i].vote_result_2 = 0                                                          //所收到的选票信息
		node_form[i].Affirmative_vote = 5                                                       //所持有的选票信息 自己拥有的赞成票
		node_form[i].dissenting_vote = 5                                                        //所持有的选票信息 自己拥有的反对票
		node_form[i].waste_time = 0.001                                                         //节点生成区块所需时间
		node_form[i].Ci = 60                                                                    //节点状态评估值
		node_form[i].flag_vote = 0                                                              //节点是否投票flag
		node_form[i].S = 1                                                                      //每个节点初始的响应权重
		node_form[i].NG = 1                                                                     //每个节点初始达成共识的数据交易总数
		node_form[i].Ntotal = 1                                                                 //每个节点初始的历史数据交易总数
		node_form[i].Attribute = [4]string{"WD", "SD", "KG", "KY"}                              //响应属性
		node_form[i].Data_symbol = [3]string{"L", "M", "H"}                                     //响应特征
		node_form[i].Data_symbol_number = [4][3]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}, {0, 0, 0}} //响应特征计数
		node_form[i].Attribute_symbol = [4]string{"", "", "", ""}                               //响应类型最终特征
		node_form[i].O_Answer = 0.0                                                             //主节点最终响应值
		node_form[i].Object_approval = 0                                                        //节点投出赞成票的节点  在原始节点中的信息 number
		node_form[i].Object_opposition = 0                                                      //节点投出反对票的节点  在原始节点中的信息 number
		node_form[i].N_Conf_TS = 0.0                                                            //主节点统计最终响应值
		rand.Seed(time.Now().UnixNano())
		time.Sleep(time.Millisecond)
		node_form[i].offline_number = float64(rand.Intn(5)) //节点离线次数
		node_form[i].YN_massage = 0                         //有无交易
	}
	//物联网节点基本信息初始化
	for i := 0; i < LN_number+EY_number; i++ { //节点信息初始化
		LightNode_form[i].number = i       //节点编号
		temp_string := strconv.Itoa(i)     //转成string
		LightNode_form[i].DT = temp_string //生成节点身份地址为节点编号

		rand.Seed(time.Now().UnixNano())
		time.Sleep(time.Millisecond)
		//node_form[i].offline_number = float64(rand.Intn(5)) //节点离线次数
	}
	//节点具体信息初始化
	//物联网节点
	for i := 0; i < LN_number; i++ {
		LightNode_form[i].number = i //节点编号
		LightNode_form[i].rank = 0   //节点等级 0正常节点 1恶意节点
		a := rand.Intn(4)            // 生成 -100 至 100 之间的随机数
		switch a {
		case 0:
			LightNode_form[i].Attribute = "WD" //0-物联网节点类型  WD温度传感器 SD湿度传感器 KG称重传感器 KY门开关等
			LightNode_form[i].Type = 1         //节点数据类型 验证节点：0-0/1型监测数据，1-连续型监测数据
		case 1:
			LightNode_form[i].Attribute = "SD"
			LightNode_form[i].Type = 1
		case 2:
			LightNode_form[i].Attribute = "KG"
			LightNode_form[i].Type = 1
		case 3:
			LightNode_form[i].Attribute = "KY"
			LightNode_form[i].Type = 0
		}
		//fmt.Printf("物联网第%d个节点创建..\n", i)
		LN_pool = append(LN_pool, LightNode_form[i])
	}
	//物联网节点--恶意节点
	for i := LN_number; i >= LN_number && i < LN_number+EY_number; i++ {
		LightNode_form[i].number = i
		LightNode_form[i].DT = strconv.Itoa(i)
		LightNode_form[i].rank = 1 //节点等级 0正常节点 1恶意节点
		a := rand.Intn(4)          // 生成 -100 至 100 之间的随机数
		switch a {
		case 0:
			LightNode_form[i].Attribute = "WD" //0-物联网节点类型  WD温度传感器 SD湿度传感器 KG称重传感器 KY门开关等
			LightNode_form[i].Type = 1         //节点数据类型 验证节点：0-0/1型监测数据，1-连续型监测数据
		case 1:
			LightNode_form[i].Attribute = "SD"
			LightNode_form[i].Type = 1
		case 2:
			LightNode_form[i].Attribute = "KG"
			LightNode_form[i].Type = 1
		case 3:
			LightNode_form[i].Attribute = "KY"
			LightNode_form[i].Type = 0
		}
		//fmt.Printf("物联网第%d个节点(恶意)创建..\n", i)
		EY_pool = append(EY_pool, LightNode_form[i])
	}
	//网关
	for i := 0; i < WG_number; i++ {
		node_form[i].types = 1            //1-网关 投票 共识
		node_form[i].rank = 0             //节点等级 0正常节点 1恶意节点
		node_form[i].vote_powr = true     //投票权
		node_form[i].check_id_time = 0.01 //身份验证时间
		node_form[i].Affirmative_vote = 8 //网关所持有赞成的选票信息

		node_form[i].S_Delay = float64(1+rand.Intn(50)) / 100                //通信延时
		node_form[i].S_PLR = float64(1+rand.Intn(10)) / 100                  //丢包率
		node_form[i].Sri = 100/node_form[i].S_Delay + 100/node_form[i].S_PLR //通信状态评估值

		node_form[i].P_Consume = float64(10 + rand.Intn(5)) //功耗
		node_form[i].P_Power = float64(50 + rand.Intn(10))  //算力大小
		node_form[i].P_Storage = float64(5 + rand.Intn(5))  //可用存储
		node_form[i].P_Total = float64(10 + rand.Intn(5))   //全部存储

		node_form[i].Ai = float64(80 + rand.Intn(20)) //交易分

		WG_pool = append(WG_pool, node_form[i]) //切片，每个元素为结构体

	}

	//设置初始网络位置  1000*1000的网格，一共1个方格(固定节点数)
	rand.Seed(time.Now().UnixNano()) // 设置随机种子
	// 定义两个维度的范围
	min := 100
	max := 900
	fmt.Printf("网关\n")
	for i := 0; i < len(WG_pool); i++ { //网关 WG_pool
		WG_pool[i].position[0] = rand.Intn(max-min) + min
		WG_pool[i].position[1] = rand.Intn(max-min) + min
		//fmt.Printf("%d,%d\n", WG_pool[i].position[0], WG_pool[i].position[1])
		//fmt.Printf("%d,%d number=%d\n", WG_pool[i].position[0], WG_pool[i].position[1], node_form[i].number)
	}
	fmt.Printf("网关节点信息创建完毕\n")
	for i := 0; i < WG_number; i++ { //同步
		for k := 0; k < len(WG_pool); k++ {
			if node_form[i].number == WG_pool[k].number {
				node_form[i] = WG_pool[k]
			}
		}
	}
	///
	fmt.Printf("物联网节点\n")
	a := 0
	for i := 0; i < len(WG_pool); i++ { //
		for j := 0; j < each_area_RN; j++ { //物联网节点
			if a < LN_number {
				LN_pool[a].position[0] = WG_pool[i].position[0] + rand.Intn(201) - 100
				LN_pool[a].position[1] = WG_pool[i].position[1] + rand.Intn(201) - 100
				LN_pool[a].region = node_form[i].number
			} else {
				EY_pool[a-LN_number].position[0] = WG_pool[i].position[0] + rand.Intn(201) - 100
				EY_pool[a-LN_number].position[1] = WG_pool[i].position[1] + rand.Intn(201) - 100
				EY_pool[a-LN_number].region = node_form[i].number
			}
			a = a + 1
			//fmt.Println(LightNode_form[a].region)
			/*
					if a == 0 {
						fmt.Printf("正常节点坐标\n")
					}
					if a < LN_number {
						fmt.Printf("%d,%d\n", LN_pool[a].position[0], LN_pool[a].position[1])
						if a == LN_number-1 {
							fmt.Printf("恶意节点坐标\n")
						}
					} else {
						fmt.Printf("%d,%d\n", EY_pool[a-LN_number].position[0], EY_pool[a-LN_number].position[1])
					}
					//fmt.Printf("%d,%d from_WGnumber=%d\n", LN_pool[j].position[0], LN_pool[j].position[1], node_form[i].number)
				a = a + 1
			*/
		}
	}
	a = 0
	fmt.Printf("物联网节点信息创建完毕\n")
	for i := 0; i < LN_number; i++ { //同步
		for k := 0; k < len(LN_pool); k++ {
			if LightNode_form[i].number == LN_pool[k].number {
				LightNode_form[i] = LN_pool[k]
				//fmt.Println(LN_pool[k].region)
				//fmt.Println(LightNode_form[i].region)
			}
		}
	}
	for i := LN_number; i >= LN_number && i < LN_number+EY_number; i++ { //同步
		for k := 0; k < len(EY_pool); k++ {
			if LightNode_form[i].number == EY_pool[k].number {
				LightNode_form[i] = EY_pool[k]
			}
		}
	}
}

//定义函数make_genesisBlock 实现创世区块添加
func make_genesisBlock() {
	fmt.Printf("创世区块创建...\n")
	t := time.Now()         //创世区块创建
	genesisBlock := Block{} //创建创世区块
	genesisBlock = Block{0, t.String(), 0, calculateBlockHash(genesisBlock), "", "", ""}
	spew.Dump(genesisBlock)                       //显示创世区块的信息
	Side_chain = append(Side_chain, genesisBlock) //追加创世区块
	Main_chain = append(Main_chain, genesisBlock)

	for k := 0; k < WG_number; k++ {
		WG_pool[k].General_block = append(WG_pool[k].General_block, genesisBlock) //节点同步 追加创世区块
	}
	for i := 0; i < len(node_form); i++ {
		node_form[i].General_block = append(node_form[i].General_block, genesisBlock) //节点同步 追加创世区块
	}
	fmt.Printf("创世区块创建完毕\n")
}

//定义函数make_block_message 实现区块交易产生
func make_block_message() {
	a := 0
	e := NewEvent(Nmessage)
	for i := 0; i < Nmessage; i++ { //产生交易循环

		sigmoid(T_r) //交易请求间隔
		block_message_body[i].Conflag1 = e[i].Conflag1
		block_message_body[i].Keywords = e[i].Keywords
		block_message_body[i].Condition = e[i].Condition
		block_message_body[i].Time_Stamp = e[i].Time_Stamp
		block_message_body[i].L_Hash = e[i].L_Hash
		block_message_body[i].Address = e[i].Address
		block_message_body[i].SData = e[i].SData
		block_message_body[i].region = e[i].region
		block_message_body[i].Index = e[i].Index
		block_message_body[i].Conf_TS = e[i].Conf_TS //数据交易置信度
		//block_message[0+a].block_message_body = append(block_message[0+a].block_message_body, block_message_body[i])
		a += 1
	}
	//fmt.Println(LightNode_formset)
	//fmt.Printf("\n")
}

//新建交易
var (
	Event = make([]Ebody, Nmessage+dataset_amount)
)

//数据交易构建
func NewEvent(n int) []Ebody {
	fmt.Printf("数据交易构建...\n")
	for i := 0; i < n; i++ {
		//各个物联网节点产生数据
		for j := 0; j < LN_number; j++ {
			t := time.Now() //获取当前时间
			LightNode_form[j].Time = t.String()
			switch LightNode_form[j].Attribute { //产生节点数据 WD温度传感器 SD湿度传感器 KG称重传感器 KY门开关等
			case "WD":
				LightNode_form[j].Data = rand.Float64()*10 + 20 //连续型范围在20.00-30.00之间
			case "SD":
				LightNode_form[j].Data = rand.Float64()*10 + 20
			case "KG":
				LightNode_form[j].Data = rand.Float64()*10 + 20
			case "KY":
				LightNode_form[j].Data = 1 //0/1型数据为0.00/1.00 float64(rand.Intn(2)) 暂且都设置为1
			}
		}
		rand_timew := 0
		rand_timen := 0
		for j := LN_number; j >= LN_number && j < LN_number+EY_number; j++ {
			t := time.Now() //获取当前时间
			LightNode_form[j].Time = t.String()
			rand_timew++
			switch LightNode_form[j].Attribute { //产生节点数据 WD温度传感器 SD湿度传感器 KG称重传感器 KY门开关等
			case "WD":
				if rand_timew == 1 {
					LightNode_form[j].Data = rand.Float64()*10 + 20 //rand.Float64() * 90连续型范围在0.00-90.00之间 1/2概率10-20  1/2概率20-30 30-40
				} else if rand_timew == 2 && rand_timen == 0 {
					LightNode_form[j].Data = rand.Float64()*10 + 30
					rand_timen = 1
				} else if rand_timew == 2 && rand_timen == 1 {
					LightNode_form[j].Data = rand.Float64()*10 + 10
					rand_timen = 0
				}
			case "SD":
				if rand_timew == 1 {
					LightNode_form[j].Data = rand.Float64()*10 + 20
				} else if rand_timew == 2 && rand_timen == 0 {
					LightNode_form[j].Data = rand.Float64()*10 + 30 //rand.Float64() * 90连续型范围在0.00-90.00之间 1/3概率10-20 20-30 30-40
					rand_timen = 1
				} else if rand_timew == 2 && rand_timen == 1 {
					LightNode_form[j].Data = rand.Float64()*10 + 10
					rand_timen = 0
				}
			case "KG":
				if rand_timew == 1 {
					LightNode_form[j].Data = rand.Float64()*10 + 20
				} else if rand_timew == 2 && rand_timen == 0 {
					LightNode_form[j].Data = rand.Float64()*10 + 30 //rand.Float64() * 90连续型范围在0.00-90.00之间 1/3概率10-20 20-30 30-40
					rand_timen = 1
				} else if rand_timew == 2 && rand_timen == 1 {
					LightNode_form[j].Data = rand.Float64()*10 + 10
					rand_timen = 0
				}
			case "KY":
				if rand_timew == 1 {
					LightNode_form[j].Data = 1 // float64(rand.Intn(2)) 0/1型数据为0.00/1.00
				} else if rand_timew == 2 {
					LightNode_form[j].Data = 0
				}
			}
			if rand_timew == 2 {
				rand_timew = 0
			}
		}
		time.Sleep(time.Microsecond) //暂停一微秒 间隔数据上传与交易构建时间
		t := time.Now()              //获取当前时间
		Event[i].Time_Stamp = t.String()
		Event[i].Condition = 0
		Event[i].region = rand.Intn(WG_number-1) + 1
		node_form[Event[i].region].YN_massage = 1
		Event[i].Conflag1 = 0
		Event[i].Answer = 0.0
		Event[i].Conf_TS = 0.0
		result := ""
		//result1 := ""
		a := 0
		for j := 0; j < LN_number+EY_number; j++ {
			//fmt.Println(LightNode_form[j].region)
			//fmt.Println(Event[i].region)
			//fmt.Printf("\n")
			if LightNode_form[j].region == Event[i].region {
				result += fmt.Sprintf("%s %s %s %d %f", LightNode_form[j].DT, LightNode_form[j].Time, LightNode_form[j].Attribute, LightNode_form[j].Type, LightNode_form[j].Data)

				Event[i].SData[a].DT = LightNode_form[j].DT
				Event[i].SData[a].Time = LightNode_form[j].Time
				Event[i].SData[a].Attribute = LightNode_form[j].Attribute
				Event[i].SData[a].Type = LightNode_form[j].Type
				Event[i].SData[a].Data = LightNode_form[j].Data

				//fmt.Println(LightNode_form[j])
				//fmt.Printf("\n")
				switch LightNode_form[j].Type {
				case 0:
					if LightNode_form[j].Data == 0 {
						LightNode_form[j].Data_symbol = "L"
					} else if LightNode_form[j].Data == 1 {
						LightNode_form[j].Data_symbol = "H"
					}
				case 1:
					if 0 <= LightNode_form[j].Data && LightNode_form[j].Data < 30 {
						LightNode_form[j].Data_symbol = "L"
					} else if 30 <= LightNode_form[j].Data && LightNode_form[j].Data < 60 {
						LightNode_form[j].Data_symbol = "M"
					} else if 60 <= LightNode_form[j].Data && LightNode_form[j].Data < 90 {
						LightNode_form[j].Data_symbol = "H"
					}
				}
				Event[i].Keywords[a].DT = LightNode_form[j].DT
				Event[i].Keywords[a].Attribute = LightNode_form[j].Attribute
				Event[i].Keywords[a].Data_symbol = LightNode_form[j].Data_symbol
				//fmt.Println(Event[i].Keywords[a].DT)
				a += 1
				//fmt.Printf("\n")
			}
		}
		a = 0
		Event[i].L_Hash = Eventbody.CalculateHash1(result)
		Event[i].Index = i
		Event[i].Address = Event[i].region //交易生成的地址 位于节点地址 具体表示为节点编号
		//fmt.Println(Event[i].L_Hash)
	}
	//Signature =  Randomgenerate.RandAllString(12)
	fmt.Printf("数据交易构建完毕\n")
	return Event
}

//交易集创建 Nmessage交易量/例100  dataset_amount交易集中需交易量个数/例20 dataset_sum交易集个数=100/20=5
func trade_set_creation() {
	a := 0
	fmt.Printf("交易集构建...\n")
	for j := 0; j < dataset_sum; j++ {
		have_i := 0
		block_message[j].Number = j
		t := time.Now() //获取当前时间
		block_message[j].NTime_Stamp = t.String()
		block_message[j].NCondition = 0

		for i := 0; i < dataset_amount; i++ {
			block_message[j].block_message_body[i] = Event[a]
			block_message[j].NKeywords += fmt.Sprintf("%s", block_message_body[a].Keywords) //交易集中所有交易关键字
			block_message[j].N_SData += fmt.Sprintf("%s", block_message_body[a].SData)      //交易集中所有交易数据
			if i == dataset_amount-1 {                                                      //网关标识符以及交易集存储地址为其中最后一个交易所述的网关编号
				block_message[j].ID = block_message[j].block_message_body[i].region
				block_message[j].NAddress = fmt.Sprintf("%d", block_message[j].block_message_body[i].region)
			}
			//fmt.Printf("\n")
			a += 1
		}
		//冲突子集与非冲突子集的构建
		TS_CTnumber := 0
		//fmt.Printf("冲突子集构建\n")
		for i := 0; i < dataset_amount; i++ {
			for k := i + 1; k < dataset_amount; k++ {
				for l := 0; l < each_area_RN; l++ {
					//fmt.Println(block_message[j].block_message_body[i].Keywords[l].Attribute)
					for g := l; g < each_area_RN; g++ {
						if block_message[j].block_message_body[i].Keywords[l].Attribute == block_message[j].block_message_body[k].Keywords[g].Attribute {
							//fmt.Printf("1\n")
							if block_message[j].block_message_body[i].Keywords[l].Data_symbol != block_message[j].block_message_body[k].Keywords[g].Data_symbol && block_message[j].block_message_body[k].Conflag1 == 0 {
								block_message[j].block_message_body[i].Conflag1 = 1
								block_message[j].block_message_body[k].Conflag1 = 1
								if have_i == 0 {
									block_message[j].TS_CT[TS_CTnumber] = block_message[j].block_message_body[i]
									TS_CTnumber += 1
								}
								have_i = 1
								block_message[j].TS_CT[TS_CTnumber] = block_message[j].block_message_body[k]
								//fmt.Println(TS_CTnumber)
								//fmt.Println(block_message[j].TS_CT[TS_CTnumber].Keywords)
								TS_CTnumber += 1
							}
						}
					}
				}
			}
			TS_CTnumber = 0
		}
		//fmt.Printf("非冲突子集构建\n")
		TS_FCnumber := 0
		for i := 0; i < dataset_amount; i++ {
			if block_message[j].block_message_body[i].Conflag1 != 1 {
				block_message[j].TS_FC[TS_FCnumber] = block_message[j].block_message_body[i]
				//fmt.Println(block_message[j].TS_FC[TS_FCnumber].Keywords)
				if TS_FCnumber > 0 {
					block_message[j].TS_FC[TS_FCnumber].F_Hash = block_message[j].TS_FC[TS_FCnumber-1].L_Hash
					//fmt.Println(block_message[j].TS_FC[TS_FCnumber].L_Hash) //交易头哈希
					//fmt.Println(block_message[j].TS_FC[TS_FCnumber].F_Hash) //交易尾哈希 缺block_message[j].TS_FC[0].F_Hash 作为后续和其他交易连接的桥梁
					//fmt.Printf("\n")
				}
				TS_FCnumber += 1
			}
		}
		TS_FCnumber = 0
		/*
			fmt.Printf("交易集%d中的冲突子集\n", j)
			fmt.Println(block_message[j].TS_CT)
			fmt.Printf("交易集%d中的非冲突子集\n", j)
			fmt.Println(block_message[j].TS_FC)
			fmt.Printf("\n")
		*/
		block_message[j].MHash = Eventbody.CalculateHash1(block_message[j].N_SData)
		//fmt.Println(block_message[j].MHash)
		have_i = 0
	}
	//fmt.Printf("\n")
	//同步交易信息到block_message_body
	for k := 0; k < Nmessage; k++ {
		for j := 0; j < dataset_sum; j++ {
			counttx++
			for i := 0; i < dataset_amount; i++ {
				if block_message_body[k].L_Hash == block_message[j].block_message_body[i].L_Hash {
					block_message_body[k] = block_message[j].block_message_body[i]
					//fmt.Println(block_message_body[k].Conflag1)
					//fmt.Printf("\n")
				}
			}
		}
	}
	fmt.Printf("交易集、冲突子集与非冲突子集构建完毕\n")
	a = 0
}

//临近父节点抽样与响应 WG_number WG_m
func near_parents() {
	//父节点抽样
	fmt.Printf("父节点抽样...\n")
	rand.Seed(time.Now().UnixNano())
	arr := generateRandomArray(WG_m, 0, WG_number-1) //生成WG_m个随机不重复的数组 范围在0至WG_number-1之间
	//fmt.Println("原始数组：", arr)

	//WG_m个网关节点中 随机选择其中1个节点作为主节点
	ALLNode_Number := rand.Intn(WG_m)
	//fmt.Println(node_form[arr[ALLNode_Number]].number)

	//计算其他节点到达主节点的通信距离
	ALLNode_x := node_form[arr[ALLNode_Number]].position[0] //主节点x轴
	ALLNode_y := node_form[arr[ALLNode_Number]].position[1] //主节点y轴
	newArray := make([]int, WG_m)
	for i := 0; i < len(arr); i++ {
		newArray[i] = int((math.Sqrt(math.Pow(float64(node_form[arr[i]].position[0]-ALLNode_x), 2) + math.Pow(float64(node_form[arr[i]].position[1]-ALLNode_y), 2))) / connect_time)
		counttx++
	}
	//fmt.Println("新的数组(未排序)：", newArray)

	// 生成新的数组根据新数组的排序变化
	indexes := make([]int, WG_m)
	for i := range indexes {
		indexes[i] = i
	}

	// 使用自定义排序将新数组按升序排列
	customSort(newArray, indexes)

	// 根据新数组的排序变化生成另一个新数组
	sortedArray := make([]int, WG_m)
	for i, index := range indexes {
		sortedArray[i] = arr[index]
	}
	//fmt.Println("新的数组(已排序)：", newArray)
	//fmt.Println("原始数组根据新数组的排序变化生成的另一个新数组：", sortedArray)
	kparents := make([]int, WG_k)
	for i := 1; i < WG_k+1; i++ { //选择距离排名前k个父节点
		kparents[i-1] = sortedArray[i]
		//fmt.Println(node_form[kparents[i-1]].position)
	}
	//fmt.Println("前k个父节点：", kparents)
	fmt.Printf("父节点抽样完毕\n")

	//父节点响应
	fmt.Printf("父节点响应...\n")

	//响应方式 遍历自身所有交易集中各个数据特征占比 选取占比最大的数据特征作为响应依据
	Answer := 0.0

	for k := 0; k < WG_k; k++ {
		for l := 0; l < dataset_sum; l++ {
			counttx++
		}
	}
	for i := 0; i < WG_k; i++ {
		//fmt.Println(node_form[kparents[i]].number)
		for j := 0; j < Nmessage; j++ {
			//fmt.Println(block_message_body[j].region)
			/*
				if node_form[kparents[i]].number == block_message_body[j].region {
					fmt.Println(block_message_body[j].SData)
				}
			*/
			//统计节点每个交易所含响应特征的总量
			for g := 0; g < len(node_form[kparents[i]].Attribute); g++ {
				for k := 0; k < each_area_RN; k++ {
					if node_form[kparents[i]].Attribute[g] == block_message_body[j].Keywords[k].Attribute {
						if node_form[kparents[i]].Data_symbol[0] == block_message_body[j].Keywords[k].Data_symbol { //响应特征L
							node_form[kparents[i]].Data_symbol_number[g][0] += 1
						}
						if node_form[kparents[i]].Data_symbol[1] == block_message_body[j].Keywords[k].Data_symbol { //响应特征M
							node_form[kparents[i]].Data_symbol_number[g][1] += 1
						}
						if node_form[kparents[i]].Data_symbol[2] == block_message_body[j].Keywords[k].Data_symbol { //响应特征LH
							node_form[kparents[i]].Data_symbol_number[g][2] += 1
						}
					}
				}
			}
		}
		/*
			fmt.Printf("WG %d\n", node_form[kparents[i]].number)
			for g := 0; g < len(node_form[kparents[i]].Data_symbol_number); g++ {
				fmt.Println(node_form[kparents[i]].Data_symbol_number[g])
			}
		*/
		//计算父节点中每个数据类型的最终特征 并进行标记
		for g := 0; g < len(node_form[kparents[i]].Attribute); g++ {
			maxValue := 0
			for j, value := range node_form[kparents[i]].Data_symbol_number[g] {
				if value > maxValue {
					maxValue = value
					node_form[kparents[i]].Attribute_symbol[g] = node_form[kparents[i]].Data_symbol[j]
				}
			}
		}
		/*
			for g := 0; g < len(node_form[kparents[i]].Attribute_symbol); g++ {
				fmt.Println(node_form[kparents[i]].Attribute_symbol[g])
			}
			fmt.Printf("\n")
		*/
		//初轮父节点响应权重 = 无冲突交易量/有冲突交易量 (讨论有无共识交易量)
		no_Conflag1 := 0
		have_Conflag1 := 0
		if success_TSnumber == 0 {
			for j := 0; j < dataset_sum; j++ {
				for k := 0; k < dataset_amount; k++ {
					if node_form[kparents[i]].number == block_message[j].block_message_body[k].region {
						if block_message[j].block_message_body[k].Conflag1 != 1 {
							have_Conflag1 += 1
						} else {
							no_Conflag1 += 1
						}
					}
				}
			}
			//若该节点没有交易 用全网交易的进行技术
			if no_Conflag1 == 0 {
				//fmt.Println(node_form[kparents[i]].number)
				//fmt.Println(node_form[kparents[i]].YN_massage)
				//fmt.Printf("\n")
				for g := 0; g < Nmessage; g++ {
					//fmt.Println(block_message_body[j].Conflag1)
					if block_message_body[g].Conflag1 == 1 {
						have_Conflag1 += 1
					} else {
						no_Conflag1 += 1
					}
				}
			}
			//fmt.Println(have_Conflag1)
			//fmt.Println(no_Conflag1)
			node_form[kparents[i]].S = float64(no_Conflag1) / float64((have_Conflag1 + no_Conflag1))
			//fmt.Printf("%f", node_form[kparents[i]].S)
			//fmt.Printf("\n")
			no_Conflag1 = 0
			have_Conflag1 = 0
		} else {
			node_form[kparents[i]].S = float64(success_TSnumber / Nmessage)
		}
		//q_number := 0
		//父节点根据最终特征 对交易进行响应
		for j := 0; j < Nmessage; j++ {
			for g := 0; g < len(node_form[kparents[i]].Attribute); g++ {
				for k := 0; k < each_area_RN; k++ {
					if node_form[kparents[i]].Attribute[g] == block_message_body[j].Keywords[k].Attribute {
						//l := g
						//if i == 1 {
						//fmt.Println(block_message_body[j].Keywords[k])
						//fmt.Println(node_form[kparents[i]].Attribute[g])
						//fmt.Println(node_form[kparents[i]].Attribute_symbol[g])
						//}
						//fmt.Println(block_message_body[j].Keywords[k].Data_symbol)
						//fmt.Printf("\n")
						if node_form[kparents[i]].Attribute_symbol[g] == block_message_body[j].Keywords[k].Data_symbol { //响应特征L
							//q_number += 1
							node_form[kparents[i]].q = 1
						} else {
							node_form[kparents[i]].q = 0
						}
						node_form[ALLNode_Number].O_Answer += float64(node_form[kparents[i]].q) / 7 //最终响应值计算 node_form[kparents[i]].S *
					}
				} /*
					if q_number == 3 {
						node_form[kparents[i]].q = 2
					} else if 0 > q_number && q_number < 3 {
						node_form[kparents[i]].q = 1
					} else if q_number == 0 {
						node_form[kparents[i]].q = 0
					}*/
			}
			block_message_body[j].Answer += node_form[ALLNode_Number].O_Answer / float64(len(node_form[kparents[i]].Data_symbol))
			if Answer < block_message_body[j].Answer {
				Answer = block_message_body[j].Answer
			}
			node_form[ALLNode_Number].O_Answer = 0
			//q_number = 0
		}
	}
	//fmt.Println(Answer)
	//fmt.Printf("\n")
	//交易响应阈值判断
	//fmt.Printf("\n")
	for j := 0; j < Nmessage; j++ {
		//fmt.Println(block_message_body[j].Answer)
		//fmt.Println(block_message_body[j].L_Hash)
		if block_message_body[j].Conflag1 == 0 {
			if block_message_body[j].Answer > a1*Answer {
				block_message_body[j].Condition = 1
				//fmt.Println(block_message_body[j].L_Hash)
			}
		}
		if block_message_body[j].Conflag1 == 1 {
			if block_message_body[j].Answer > a2*Answer {
				block_message_body[j].Condition = 1
				//fmt.Println(block_message_body[j].L_Hash)
			}
		}
	}
	for j := 0; j < Nmessage; j++ {
		if block_message_body[j].Condition == 1 {
			Nmessage_XY += 1
		}
	}
	//更新交易个数以及交易集个数
	dataset_sum_XY = Nmessage_XY / dataset_amount
	remainder := Nmessage_XY % dataset_amount
	if remainder > 0 {
		dataset_sum_XY += 1 // 如果有余数，则将结果加1
	}
	//fmt.Println(Nmessage_XY)
	//fmt.Println(dataset_sum_XY)
	//fmt.Printf("\n")
	Answer = 0

	//更新交易集
	for i := 0; i < dataset_sum; i++ {
		//清空原交易
		block_message[i].block_message_body = [dataset_amount]Ebody{}
		block_message[i].TS_CT = [dataset_amount]Ebody{}
		block_message[i].TS_FC = [dataset_amount]Ebody{}
		//将Condition = 1(待共识状态)交易更新在交易集中 新交易集更新为待共识交易集
	}
	b := 0
	a := 0
	for i := 0; i < dataset_sum_XY; i++ {
		for j := 0; j < dataset_amount; j++ {
			for k := b; k < Nmessage; k++ {
				if block_message_body[k].Condition == 1 {
					block_message[i].block_message_body[j] = block_message_body[k]
					block_message_body_XY[a] = block_message[i].block_message_body[j]
					if j > 0 {
						block_message[i].block_message_body[j].F_Hash = block_message[i].block_message_body[j-1].L_Hash
						//fmt.Println(block_message[i].block_message_body[j].F_Hash)
					}
					//fmt.Println(block_message[i].block_message_body[j].Keywords)
					//fmt.Println(block_message_body_XY[a].L_Hash)
					//fmt.Println(b)
					b = k + 1
					a = a + 1
					if a == Nmessage_XY {
						break
					}
					break
				}
			}
			if j == dataset_amount-1 { //网关标识符以及交易集存储地址为其中最后一个交易所述的网关编号
				block_message[i].ID = block_message[i].block_message_body[j].region
				block_message[i].NAddress = fmt.Sprintf("%d", block_message[i].block_message_body[j].region)
			}
			if a == Nmessage_XY {
				break
			}
		}
		if a == Nmessage_XY {
			break
		}
		block_message[i].Number = i
		t := time.Now() //获取当前时间
		block_message[i].NTime_Stamp = t.String()
		block_message[i].NCondition = 1 //将新交易集状态更新为待共识交易集
		block_message[i].MHash = Eventbody.CalculateHash1(block_message[i].N_SData)
		//fmt.Println(block_message[i].MHash)
	}
	//fmt.Println(a)
	//fmt.Println(b)
	b = 0
	for i := 0; i < Nmessage; i++ {
		for j := 0; j < 1; j++ {
			for k := LN_number; k >= LN_number && k < LN_number+EY_number; k++ {
				//fmt.Println(LightNode_form[k].DT)
				//fmt.Println(block_message_body[i].Keywords[j].DT)
				//fmt.Printf("\n")
				if block_message_body[i].Keywords[j].DT == LightNode_form[k].DT {
					EY_Nmessage_number += 1
					//fmt.Println(block_message_body[i].L_Hash)
				}
			}
		}
	}

	for i := 0; i < Nmessage_XY; i++ {
		for j := 0; j < 1; j++ {
			for k := LN_number; k >= LN_number && k < LN_number+EY_number; k++ {
				//fmt.Println(LightNode_form[k].DT)
				//fmt.Println(block_message_body[i].Keywords[j].DT)
				//fmt.Printf("\n")
				if block_message_body[i].Keywords[j].DT == LightNode_form[k].DT {
					EY_NmessageXY_number += 1
					//fmt.Println(block_message_body[i].L_Hash)
				}
			}
		}
	}
	//清空原有交易 将同步交易集信息到block_message_body_XY交易放入block_message_body
	for i := 0; i < len(block_message_body); i++ {
		block_message_body[i] = Ebody{}
	}
	for i := 0; i < Nmessage_XY; i++ {
		block_message_body[i] = block_message_body_XY[i]
		//fmt.Println(block_message_body[i].L_Hash)
	}
	//清空block_message_body_XY
	for i := 0; i < Nmessage_XY; i++ {
		block_message_body_XY[i] = Ebody{}
		//fmt.Println(block_message_body[i].L_Hash)
	}
	fmt.Printf("父节点响应完毕\n")
}

//定义函数General_test 实现区块共识 完成普通消息
func General_test() { //区块共识与测试

	block_make_point = block_make_point[0:0] //清除构建区块时主节点编号
	rand.Seed(time.Now().UnixNano())
	//WG_N := rand.Intn(WG_number) //从子网关中挑选一个代表全节点N 完成普通消息共识
	//选择子网关节点 即产生待验证交易集的网关
	ZWG_KStr := make([]string, dataset_sum_XY)
	ZWG_Kint = make([]int, dataset_sum_XY)
	for i := 0; i < dataset_sum_XY; i++ {
		ZWG_KStr[i] = fmt.Sprintf("%s", block_message[i].NAddress)
		num, err := strconv.Atoi(ZWG_KStr[i])
		if err != nil {
			fmt.Printf("转换出错：%s\n", err.Error())
			return
		}
		ZWG_Kint[i] = num
	}
	//fmt.Println(len(ZWG_Kint))
	randomIndex := rand.Intn(len(ZWG_Kint))
	WG_N := ZWG_Kint[randomIndex]
	//fmt.Println(len(ZWG_Kint), dataset_sum_XY)
	for k := 0; k < len(ZWG_Kint); k++ {
		for l := 0; l < dataset_sum_XY; l++ {
			counttx++
		}
	}
	//fmt.Println("子网格地址数组：", ZWG_Kint)
	//fmt.Println("子网格中代表全节点：", WG_N)
	sigmoid(T_p)
	Transaction_set_consensus(ZWG_Kint, WG_N) //交易集投票
	for k := 0; k < len(chain_message); k++ {
		chain_message[k].Number = WG_pool[ZWG_Kint[k]].number
		//fmt.Println(chain_message[k].Number)
		block_make_point = append(block_make_point, chain_message[k].Number) //保存区块构建节点编号（原始节点信息）

		newBlock := generateBlock(WG_pool[chain_message[k].Number].General_block[len(WG_pool[chain_message[k].Number].General_block)-1], Nmessage, WG_pool[chain_message[k].Number].id_address) //产生区块

		Main_chain = append(Main_chain, newBlock) //追加区块
		for j := 0; j < len(node_form); j++ {     //全网节点同步到主链上
			node_form[j].General_block = append(node_form[j].General_block, newBlock)
		}

		General_safe_number = General_safe_number + 1 //共识成功次数加1
		General_safe_number_temp = General_safe_number_temp + 1
	}
}

//定义函数Transaction_set_consensus_improvement_G  实现普通消息确定交易集
func Transaction_set_consensus(ZWG_Kint []int, WG_N int) []Message {
	//若无已共识交易集 先根据自身所有待验证交易集信息进行投票
	Conf_TS := 0.0
	if success_TSnumber == 0 {
		for i := 0; i < len(ZWG_Kint); i++ {
			for j := 0; j < Nmessage_XY; j++ {
				//统计子网关每个交易所含响应特征的总量
				for g := 0; g < len(node_form[ZWG_Kint[i]].Attribute); g++ {
					for k := 0; k < each_area_RN; k++ {
						if node_form[ZWG_Kint[i]].Attribute[g] == block_message_body[j].Keywords[k].Attribute {
							if node_form[ZWG_Kint[i]].Data_symbol[0] == block_message_body[j].Keywords[k].Data_symbol { //响应特征L
								node_form[ZWG_Kint[i]].Data_symbol_number[g][0] += 1
							}
							if node_form[ZWG_Kint[i]].Data_symbol[1] == block_message_body[j].Keywords[k].Data_symbol { //响应特征M
								node_form[ZWG_Kint[i]].Data_symbol_number[g][1] += 1
							}
							if node_form[ZWG_Kint[i]].Data_symbol[2] == block_message_body[j].Keywords[k].Data_symbol { //响应特征LH
								node_form[ZWG_Kint[i]].Data_symbol_number[g][2] += 1
							}
						}
					}
				}
			}
			//计算父节点中每个数据类型的最终特征 并进行标记
			for g := 0; g < len(node_form[ZWG_Kint[i]].Attribute); g++ {
				maxValue := 0
				for j, value := range node_form[ZWG_Kint[i]].Data_symbol_number[g] {
					if value > maxValue {
						maxValue = value
						node_form[ZWG_Kint[i]].Attribute_symbol[g] = node_form[ZWG_Kint[i]].Data_symbol[j]
					}
				}
			}
			for g := 0; g < len(node_form[ZWG_Kint[i]].Attribute_symbol); g++ {
				//fmt.Println(node_form[ZWG_Kint[i]].Attribute_symbol[g])
			}
			//fmt.Printf("\n")
			//父节点根据最终特征 对交易进行投票
			for j := 0; j < Nmessage_XY; j++ {
				for g := 0; g < len(node_form[ZWG_Kint[i]].Attribute); g++ {
					for k := 0; k < each_area_RN; k++ {
						if node_form[ZWG_Kint[i]].Attribute[g] == block_message_body[j].Keywords[k].Attribute {
							/*
								if i == 1 {
									fmt.Println(block_message_body[j].Keywords[k])
									fmt.Println(node_form[ZWG_Kint[i]].Attribute[g])
									fmt.Println(node_form[ZWG_Kint[i]].Attribute_symbol[g])
								}
							*/
							if node_form[ZWG_Kint[i]].Attribute_symbol[g] == block_message_body[j].Keywords[k].Data_symbol { //响应特征L
								node_form[ZWG_Kint[i]].Object_approval = 1
							} else {
								node_form[ZWG_Kint[i]].Object_opposition = 1
								block_message_body[j].Conflag2 = 1 //冲突标记
							}
							if block_message_body[j].Conflag2 == 1 && node_form[ZWG_Kint[i]].Attribute_symbol[g] == block_message_body[j].Keywords[k].Data_symbol {
								node_form[WG_N].N_Conf_TS += float64(node_form[ZWG_Kint[i]].Object_approval) - float64(node_form[ZWG_Kint[i]].Object_opposition)/float64(each_area_RN)
							} else if block_message_body[j].Conflag2 == 0 {
								node_form[WG_N].N_Conf_TS += float64(node_form[ZWG_Kint[i]].Object_approval)
							}
						}
					}
				}
				block_message_body[j].Conf_TS += node_form[WG_N].N_Conf_TS / float64(len(node_form[ZWG_Kint[i]].Data_symbol))
				if Conf_TS < block_message_body[j].Conf_TS {
					Conf_TS = block_message_body[j].Conf_TS
				}
				node_form[WG_N].N_Conf_TS = 0
			}
		}
	}
	//fmt.Println(Conf_TS)
	//交易共识阈值判断
	for j := 0; j < Nmessage_XY; j++ {
		//fmt.Println(block_message_body[j].Conf_TS)
		//fmt.Println(block_message_body[j].Conflag1)
		//判断交易响应值是否达到共识要求
		if block_message_body[j].Conf_TS >= b1*Conf_TS {
			block_message_body[j].Condition = 2
			//fmt.Println(block_message_body[j].L_Hash)
		}
	}
	//更新交易个数以及交易集个数
	for j := 0; j < Nmessage_XY; j++ {
		if block_message_body[j].Condition == 2 {
			Nmessage_GS += 1
			counttx++
		}
	}
	dataset_sum_GS = Nmessage_GS / dataset_amount
	remainder := Nmessage_GS % dataset_amount
	if remainder > 0 {
		dataset_sum_GS += 1 // 如果有余数，则将结果加1
	}
	classification_2 = dataset_sum_GS //共识投票周期
	//fmt.Println(Nmessage_GS)
	//fmt.Println(dataset_sum_GS)
	//更新已共识交易集
	a := 0
	for i := 0; i < dataset_sum_GS; i++ {
		//清空原交易
		block_message[i].block_message_body = [dataset_amount]Ebody{}
		block_message[i].TS_CT = [dataset_amount]Ebody{}
		block_message[i].TS_FC = [dataset_amount]Ebody{}
		//将Condition = 2(待共识状态)交易更新在交易集中 新交易集更新为待共识交易集
	}
	b := 0
	c := 0
	for i := 0; i < dataset_sum_GS; i++ {
		for j := 0; j < dataset_amount; j++ {
			for k := b; k < Nmessage_XY; k++ {
				if block_message_body[j].Condition == 2 {
					block_message[i].block_message_body[j] = block_message_body[k]
					block_message_body_XY[c] = block_message[i].block_message_body[j]
					if j > 0 {
						block_message[i].block_message_body[j].F_Hash = block_message[i].block_message_body[j-1].L_Hash
						//fmt.Println(block_message[i].block_message_body[j].F_Hash)
					}
					//fmt.Println(block_message[i].block_message_body[j].Keywords)
					//fmt.Println(block_message_body_XY[c].L_Hash)
					//fmt.Println(b)
					b = k + 1
					c = c + 1
					if c == Nmessage_GS {
						break
					}
					break
				}
			}
			if j == dataset_amount-1 { //网关标识符以及交易集存储地址为其中最后一个交易所述的网关编号
				block_message[i].ID = block_message[i].block_message_body[j].region
				block_message[i].NAddress = fmt.Sprintf("%d", block_message[i].block_message_body[j].region)
			}
			if a == Nmessage_GS {
				break
			}
		}
		if a == Nmessage_GS {
			break
		}
		block_message[i].Number = i
		t := time.Now() //获取当前时间
		block_message[i].NTime_Stamp = t.String()
		block_message[i].NCondition = 2 //将新交易集状态更新为待共识交易集
		block_message[i].MHash = Eventbody.CalculateHash1(block_message[i].N_SData)
		//fmt.Println(block_message[i].MHash)
	}
	//fmt.Println(c)
	//fmt.Println(b)
	b = 0

	//清空原有交易 将同步交易集信息到block_message_body_XY交易放入block_message_body
	for i := 0; i < len(block_message_body); i++ {
		block_message_body[i] = Ebody{}
	}
	for i := 0; i < Nmessage_GS; i++ {
		block_message_body[i] = block_message_body_XY[i]
		//fmt.Println(block_message_body[i].L_Hash)
	}
	//清空block_message_body_XY
	for i := 0; i < Nmessage_GS; i++ {
		block_message_body_XY[i] = Ebody{}
		//fmt.Println(block_message_body[i].L_Hash)
	}
	//将已经成功达成共识的交易进行上链 首先赋值到chain_message
	for i := 0; i < dataset_sum_GS; i++ {
		chain_message = append(chain_message, block_message[i])
	}
	//fmt.Println(chain_message)
	for i := 0; i < Nmessage_GS; i++ {
		for j := 0; j < 1; j++ {
			for k := LN_number; k >= LN_number && k < LN_number+EY_number; k++ {
				//fmt.Println(LightNode_form[k].DT)
				//fmt.Println(block_message_body[i].Keywords[j].DT)
				//fmt.Printf("\n")
				if block_message_body[i].Keywords[j].DT == LightNode_form[k].DT {
					EY_NmessageGS_number += 1
					//fmt.Println(block_message_body[i].L_Hash)
				}
			}
		}
	}
	fmt.Println("完成权重投票")
	fmt.Printf("本轮恶意节点产生交易个数：%d\n", EY_Nmessage_number)
	fmt.Printf("本轮正常交易个数：%d\n", Nmessage-EY_Nmessage_number)
	fmt.Printf("本轮通过响应交易中含有恶意交易数据的个数：%d\n", EY_NmessageXY_number)
	fmt.Printf("本轮通过共识交易中含有恶意交易数据的个数：%d\n", EY_NmessageGS_number)
	fmt.Printf("本轮通过共识交易个数：%d\n", Nmessage_GS)

	ZB_EY_NmessageGS_number = float64(EY_NmessageGS_number) / float64(Nmessage_GS)
	fmt.Printf("本轮通过共识交易中含有恶意节点产出数据占比：%f\n", ZB_EY_NmessageGS_number)
	fmt.Printf("本轮通过共识交易中不含恶意节点产出数据占比：%f\n", 1-ZB_EY_NmessageGS_number)
	var Rejection float64 = float64(EY_Nmessage_number-EY_NmessageGS_number) / float64(EY_Nmessage_number)
	fmt.Printf("本轮通过共识交易中恶意交易剔除率：%f\n", Rejection)

	again_number += 1
	Nmessage_GS_number = (OK() + Nmessage_GS_number) / float64(again_number) //共识交易中数据特征一致率
	EY_Nmessage_number1 += EY_Nmessage_number
	EY_NmessageXY_number1 += EY_NmessageXY_number
	EY_NmessageGS_number1 += EY_NmessageGS_number
	chian_message_number1 += Nmessage_GS
	fmt.Printf("\n")
	fmt.Printf("总共恶意节点产生交易个数：%d\n", EY_Nmessage_number1)
	fmt.Printf("总共正常交易个数：%d\n", NNmessage-EY_Nmessage_number1)
	fmt.Printf("总共通过响应交易中含有恶意交易数据的个数：%d\n", EY_NmessageXY_number1)
	fmt.Printf("总共通过共识交易中含有恶意交易数据的个数：%d\n", EY_NmessageGS_number1)
	fmt.Printf("总共通过共识交易个数：%d\n", chian_message_number1)
	fmt.Printf("总共共识交易中恶意交易剔除率：%f\n", Rejection)
	fmt.Printf("总共共识交易中数据特征一致率:%f\n", Nmessage_GS_number)
	return chain_message
}

//
func OK() float64 {
	Nmessage_GS_OK := 0.0
	WD_L := 0
	WD_M := 0
	WD_H := 0
	WD_TOP := 0
	WD_OK := 0.0
	SD_L := 0
	SD_M := 0
	SD_H := 0
	SD_TOP := 0
	SD_OK := 0.0
	KG_L := 0
	KG_M := 0
	KG_H := 0
	KG_TOP := 0
	KG_OK := 0.0
	KY_L := 0
	KY_H := 0
	KY_TOP := 0
	KY_OK := 0.0

	for i := 0; i < Nmessage_GS; i++ {
		for j := 0; j < each_area_RN; j++ {
			switch block_message_body[i].Keywords[j].Attribute {
			case "WD":
				//fmt.Println(block_message_body[i].Keywords[j].Data_symbol)
				switch block_message_body[i].Keywords[j].Data_symbol {
				case "L":
					WD_L += 1
				case "M":
					WD_M += 1
				case "H":
					WD_H += 1
				}
			case "SD":
				switch block_message_body[i].Keywords[j].Data_symbol {
				case "L":
					SD_L += 1
				case "M":
					SD_M += 1
				case "H":
					SD_H += 1
				}
			case "KG":
				switch block_message_body[i].Keywords[j].Data_symbol {
				case "L":
					KG_L += 1
				case "M":
					KG_M += 1
				case "H":
					KG_H += 1
				}
			case "KY":
				switch block_message_body[i].Keywords[j].Data_symbol {
				case "L":
					KY_L += 1
				case "H":
					KY_H += 1
				}
			}
		}
	}
	WD_TOP = WD_L
	if WD_M > WD_TOP {
		WD_TOP = WD_M
	}
	if WD_H > WD_TOP {
		WD_TOP = WD_H
	}
	SD_TOP = SD_L
	if SD_M > SD_TOP {
		SD_TOP = SD_M
	}
	if SD_H > SD_TOP {
		SD_TOP = SD_H
	}
	KG_TOP = KG_L
	if KG_M > KG_TOP {
		KG_TOP = KG_M
	}
	if KG_H > KG_TOP {
		KG_TOP = KG_H
	}
	KY_TOP = KY_L
	if KY_H > KY_TOP {
		KY_TOP = KY_H
	}
	WD_OK = float64(WD_TOP) / float64(WD_L+WD_M+WD_H)
	SD_OK = float64(SD_TOP) / float64(SD_L+SD_M+SD_H)
	KG_OK = float64(KG_TOP) / float64(KG_L+KG_M+KG_H)
	KY_OK = float64(KY_TOP) / float64(KY_L+KY_H)
	//fmt.Printf("%f %f %f %f\n", WD_OK, SD_OK, KG_OK, KY_OK)
	Nmessage_GS_OK = (WD_OK + SD_OK + KG_OK + KY_OK) / 4.0
	fmt.Printf("本轮共识交易中数据平均正确率:%f\n", Nmessage_GS_OK)
	return Nmessage_GS_OK
}

// 自定义排序函数
func customSort(arr []int, indexes []int) {
	for i := 0; i < len(arr)-1; i++ {
		for j := 0; j < len(arr)-1-i; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
				indexes[j], indexes[j+1] = indexes[j+1], indexes[j]
			}
		}
	}
}

//延时函数
func sigmoid(k int) float64 {
	time_score = 1 / (1 + math.Exp(float64(-k+5)))
	time_score = time_score * time_score * 500
	return time_score
}

// 生成随机不重复数组
func generateRandomArray(length, min, max int) []int {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())
	// 创建一个映射来检查生成的随机数是否重复
	used := make(map[int]bool)
	// 创建一个切片用于存储随机数
	randomArray := make([]int, 0, length)
	// 生成随机数并添加到切片中，直到切片长度为 length
	for len(randomArray) < length {
		// 生成随机数
		randomNum := rand.Intn(max-min+1) + min
		// 检查随机数是否重复，若不重复则添加到切片中
		if !used[randomNum] {
			randomArray = append(randomArray, randomNum)
			used[randomNum] = true
		}
	}
	return randomArray
}

//定义函数calculateHash 生成Hash值
func calculateHash(s string) string {
	h := sha256.New()                 //创建一个基于SHA256算法的hash.Hash接口的对象
	h.Write([]byte(s))                //输入数据
	hashed := h.Sum(nil)              //计算哈希值
	return hex.EncodeToString(hashed) //将字符串编码为16进制格式,返回字符串
}

//定义函数calculateBlockHash 生成区块Hash值
func calculateBlockHash(block Block) string {
	var temp string
	for i := 0; i < block.Nmessage; i++ {
		temp = temp + strconv.Itoa(block.Nmessage)
	}
	record := string(block.Index) + block.Time_Stamp + temp + block.PrevHash //区块内其他信息作为字符串   strconv.Itoa 整形变字符串
	return calculateHash(record)                                             //调用函数calculateHash 生成Hash值
}

//定义函数isHashValid 验证Hash的前导0个数是否与困难度系数一致
func isHashValid(hash string, difficulty int) bool {
	// 返回difficulty个0串联的字符串prefix
	prefix := strings.Repeat("0", difficulty)
	//若hash是以prefix开头，则返回true，否则返回false
	return strings.HasPrefix(hash, prefix)
}

//定义函数generateBlock 生成新区块
func generateBlock(oldBlock Block, Nmessage int, address string) Block {
	var newBlock Block //新区块

	t := time.Now()                     //获取当前时间
	newBlock.Index = oldBlock.Index + 1 //区块的增加，index也加一
	newBlock.Time_Stamp = t.String()    //时间戳
	newBlock.Nmessage = Nmessage        //区块链内交易数量
	newBlock.PrevHash = oldBlock.Hash   //新区块的PrevHash存储上一个区块的Hash
	newBlock.Validator_key = address    //验证者地址
	for i := 0; ; i++ {                 //通过循环改变 Nonce
		hex := fmt.Sprintf("%x", i) //获得字符串形式 Nonce
		newBlock.Nonce = hex        //Nonce赋值给newBlock.Nonce
		// 判断Hash的前导0个数，是否与难度系数一致
		if !isHashValid(calculateBlockHash(newBlock), difficulty-adv_difficulty) { //difficulty-adv_difficulty 表示主节点会在计算哈希时降低难度值
			//fmt.Println(calculateBlockHash(newBlock), " do more work!") //继续挖矿中
			time.Sleep(time.Second) //暂停一秒
			continue
		} else {
			//a += 1
			//fmt.Printf("第%d区块上链(计算哈希赋值)：", a)
			//fmt.Println(calculateBlockHash(newBlock), " work done!") //挖矿成功
			newBlock.Hash = calculateBlockHash(newBlock) //重新计算hash并赋值
			time.Sleep(time.Microsecond)                 //暂停一微秒
			break
		}
	}
	return newBlock //返回新区块
}

//定义函数General_Index_calculation 普通消息指标计算
func General_Index_calculation(el float64) (float64, float64, float64) {
	//指标计算
	//TPS吞吐量  定义：上链的总交易数/其区块生成时间  生成时间包括交易+共识机制执行+区块
	var result_time_2 float64 = 0
	//交易与区块进行模拟
	result_time_2 = result_time_2 + (float64(Nmessage)*0.001)*float64(General_safe_number_temp) //交易广播
	for j := 0; j < len(block_make_point); j++ {
		for k := 0; k < len(node_form); k++ {
			if node_form[k].number == block_make_point[j] {
				result_time_2 = result_time_2 + node_form[k].waste_time*(float64(dataset_amount)) //交易集构建
			}
		}
	}
	if (len(Main_chain)+len(Side_chain)-2)%check_point == 0 { //父节点抽样与响应
		result_time_2 = result_time_2 + float64(node_sum*(each_area_RN-1)-dataset_amount*classification_1)*0.01*float64(node_sum*(each_area_RN-1)-dataset_amount*classification_1)*0.01
		//fmt.Println(1)
	}

	if (len(Main_chain)+len(Side_chain)-2)%dataset_sum == 0 { //共识投票
		result_time_2 = result_time_2 + float64((node_sum*(each_area_RN-1)-dataset_amount*classification_1)/(classification_2+1))*0.01*float64((node_sum*(each_area_RN-1)-dataset_amount*classification_1)/(classification_2+1))*0.01
	}

	for m := 0; m < len(chain_message); m++ {
		chian_message_number = chian_message_number + len(chain_message[m].block_message_body)
	}
	var result_tps float64
	result_tps = float64(chian_message_number) / (result_time_2 + el) * WG_number / each_area_RN

	//delay 时间延迟 定义：交易发出到确认的时间 平均值
	//result_delay := float64(tra_time.Sub(begain).Nanoseconds()) / 1000000000 //计算将交易量用于生成区块的时间 单位秒
	var result_delay float64 = 0
	// 交易与区块进行模拟
	result_delay = result_delay + (float64(Nmessage)*0.1)*float64(General_safe_number_temp) + //交易广播
		float64(dataset_amount)*0.1*float64(check_id_time) + float64((LN_number))*0.01 + float64((WG_number))*0.01
	for j := 0; j < len(block_make_point); j++ {
		for k := 0; k < len(node_form); k++ {
			if node_form[k].number == block_make_point[j] {
				result_delay = result_delay + node_form[k].waste_time*(float64(dataset_amount)) //交易集构建
			}
		}
	}
	if (len(Main_chain)+len(Side_chain)-2)%check_point == 0 { //父节点抽样与响应
		result_delay = result_delay + float64((dataset_sum*WG_k)*(dataset_sum*WG_k))*0.001 + float64(dataset_sum*WG_m)*0.001
	}

	if (len(Main_chain)+len(Side_chain)-2)%dataset_sum == 0 { //共识投票
		result_delay = result_delay + float64(node_sum*len(ZWG_Kint))*dataset_amount*0.001
	}
	result_delay = (result_delay + el)
	//容错性  定义：保证共识安全的前提下，增加恶意节点数量，查看TPS与delay
	//见前面

	//通信代价 定义为节点间的通信次数
	var resul_exchange float64 = 0
	resul_exchange = float64(LN_number)*float64(WG_number) + //交易及交易集构建与广播
		float64(Nmessage)*float64(dataset_amount)/2 + float64((dataset_sum)*(dataset_sum-1)) +
		float64((dataset_amount)*(dataset_amount-1))*float64(General_safe_number_temp)
	if (len(Side_chain)-1)%check_point == 0 { //父节点抽样与响应
		resul_exchange = resul_exchange + float64((dataset_sum*WG_k)*(dataset_sum*WG_k)) + float64(dataset_sum*WG_m)*0.01
	}
	if (len(Main_chain)+len(Side_chain)-2)%dataset_sum == 0 { //共识投票
		resul_exchange = resul_exchange + float64(node_sum*len(ZWG_Kint))*dataset_amount
	}
	resul_exchange = math.Sqrt(resul_exchange / float64(WG_number)) //平均通信开销

	return result_tps, result_delay, resul_exchange
}

//定义函数printf_result 实现结果打印与保存
func printf_result(calculation_time int) {
	//计算平均结果
	var result_tps_avg float64 = 0
	var result_delay_avg float64 = 0
	var resul_exchange_avg float64 = 0

	for i := start_cal; i < block_number; i++ {
		result_tps_avg = result_tps_avg + result_tps[i]
	}

	for i := start_cal; i < block_number; i++ {
		result_delay_avg = result_delay_avg + result_delay[i]
	}
	if start_cal > 0 {
		result_delay_avg = result_delay_avg / (block_number - float64(start_cal) - 1)
	} else {
		result_delay_avg = result_delay_avg / block_number
	}

	for i := start_cal; i < block_number; i++ {
		resul_exchange_avg = resul_exchange_avg + resul_exchange[i]
	}
	if start_cal > 0 {
		resul_exchange_avg = resul_exchange_avg / (block_number - float64(start_cal) - 1)
	} else {
		resul_exchange_avg = resul_exchange_avg
	}

	//打印
	/*
		fmt.Println("最终平均结果：")
		fmt.Println("TPS：")
		fmt.Println(result_tps_avg)

		fmt.Println("Delay：")
		fmt.Println(result_delay_avg)

		fmt.Println("consumption：")
		fmt.Println(resul_exchange_avg)

		fmt.Println("\n普通消息-最终共识成功次数")
		fmt.Println(General_safe_number)
		fmt.Printf("\n")
		//结果保存
	*/
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

	writer.WriteString("consumption:") //写入consumption
	writer.WriteString("\n")
	writer.WriteString(strconv.FormatFloat(resul_exchange_avg, 'E', -1, 64))
	writer.WriteString("\n")
	writer.Flush()

	writer.WriteString("Datarate：") //写入数据正确率
	writer.WriteString("\n")
	writer.WriteString(strconv.FormatFloat(Nmessage_GS_number, 'E', -1, 64))
	writer.WriteString("\n")
	writer.Flush()
}

//定义函数Global_variable_initialization 完成下一次全局变量初始化
func Global_variable_initialization() {
	node_sum = WG_number + LN_number + EY_number //节点数量
	node_yz_sum = math.Floor(((WG_number + LN_number + EY_number) * 0.6))
	block_message_body = make([]Ebody, Nmessage, Nmessage)      //交易体信息
	block_message_body_XY = make([]Ebody, Nmessage, Nmessage)   //响应后交易集个数
	block_message = make([]Message, Nmessage, Nmessage)         //区块交易集合
	WG_block_message = make([]WGMessage, dataset_sum, Nmessage) //交易集
	chain_message = make([]Message, 0, Nmessage)                //上链交易集
	Nmessage_XY = 0                                             //响应之后交易的个数
	Nmessage_GS = 0                                             //共识之后交易的个数
	dataset_sum_XY = 0                                          //响应之后交易集的个数
	dataset_sum_GS = 0                                          //共识之后交易集的个数
	EY_Nmessage_number = 0                                      //恶意节点产生的恶意交易个数
	EY_NmessageGS_number = 0                                    //通过共识交易中的恶意交易个数
	EY_NmessageXY_number = 0                                    //待验证交易中的恶意交易个数
	T_Exchange = 0.0                                            //交易达成票数
	class_result_2 = make([]float64, K, K)
	classification_2 = 0 //共识投票周期 与classification_1差距不能过大//待验证交易中的恶意交易个数
	ZWG_Kint = make([]int, dataset_sum_XY)
}
