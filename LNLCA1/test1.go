package main

import (
	"LNLCA1/Eventbody"
	"fmt"
	"math"
	"math/rand"
	"time"
)

//临近父节点抽样与响应 WG_number WG_m
//临近父节点抽样与响应 WG_number WG_m
func near_parents1() {
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
		newArray[i] = int(math.Sqrt(math.Pow(float64(node_form[arr[i]].position[0]-ALLNode_x), 2) + math.Pow(float64(node_form[arr[i]].position[1]-ALLNode_y), 2)))
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
		//父节点根据最终特征 对交易进行响应
		for j := 0; j < Nmessage; j++ {
			for g := 0; g < len(node_form[kparents[i]].Attribute); g++ {
				for k := 0; k < each_area_RN; k++ {
					if node_form[kparents[i]].Attribute[g] == block_message_body[j].Keywords[k].Attribute {
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
	fmt.Printf("更新交易个数以及交易集个数\n")
	fmt.Println(Nmessage_XY)
	fmt.Println(dataset_sum_XY)
	fmt.Printf("\n")
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
	fmt.Println(a)
	fmt.Println(b)
	b = 0

	//清空原有交易 将同步交易集信息到block_message_body_XY交易放入block_message_body
	for i := 0; i < len(block_message_body); i++ {
		block_message_body[i] = Ebody{}
	}
	for i := 0; i < Nmessage_XY; i++ {
		block_message_body[i] = block_message_body_XY[i]
		//fmt.Println(block_message_body[i].L_Hash)
	}
	fmt.Printf("父节点响应完毕\n")
}
