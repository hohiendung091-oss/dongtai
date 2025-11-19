package Eventbody

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type Eventbody struct { //交易结构
	Data       int    //交易数据（Transaction_account）
	Keywords   string //交易关键词，由各物联网节点标识符DT、节点属性与其特性标记L/M/H整合组成
	Condition  int    //待验证状态Condition（0），待共识状态Condition（1），通过共识状态Condition（2）
	Time_Stamp string //时间戳，记录新交易发起以及上传至网关的时间点）
	EHash      string //交易哈希值
	Address    int    //数据存储地址，该数据交易中的所有数据信息在网关中存储的物理地址

	DT        string //节点标识符 即节点编号
	Attribute string //节点属性
	Time      string //数据采集时间戳
	Type      string //数据类型
	SData     string //交易节点数据，包括节点标识符（DT）、节点属性（Attribute）、数据采集时间戳（Time）、数据类型（Type）和数据值（Data）

	ECreator string //创造者的公钥（数字签名）
	Index    int    //在创建者创建的交易集序列中的索引
	//Area     int    //交易所属网关
	//Conflag     int    //在交易共识阶段中，对多个待共识交易集之间存在冲突的数据交易进行标记
}

// Transaction contains the payload of an Event as well as the information that
// ties it to other Events. The private fields are for local computations only.

// Marshal returns the JSON encoding of an TransactionBody
func (e *Eventbody) Marshal() ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b) //will write to b
	if err := enc.Encode(e); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Unmarshal converts a JSON encoded Transaction to an TransactionBody
func (e *Eventbody) Unmarshal(data []byte) error {
	b := bytes.NewBuffer(data)
	dec := json.NewDecoder(b) //will read from b
	if err := dec.Decode(e); err != nil {
		return err
	}
	return nil
}

// Hash returns the SHA256 hash of the JSON encoded TransactionBody.
func (e *Eventbody) Hash() ([]byte, error) {
	hashBytes, err := e.Marshal()
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	return h.Sum(hashBytes), nil
}

//定义函数calculateHash 生成Hash值
func CalculateHash1(s string) string {
	h := sha256.New()                 //创建一个基于SHA256算法的hash.Hash接口的对象
	h.Write([]byte(s))                //输入数据
	hashed := h.Sum(nil)              //计算哈希值
	return hex.EncodeToString(hashed) //将字符串编码为16进制格式,返回字符串
}
