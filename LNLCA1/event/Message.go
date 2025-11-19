package event

import (
	"bytes"
	"crypto"
	"encoding/json"
	"time"
)


type Eventbody struct {               //交易结构
	Transaction   [][]byte            // application transactions
	Timestamp     int64               //创建交易的时间戳
	Content       []string            //交易具体内容
	ECreator      []byte              //创造者的公钥（数字签名）
	EHash         int                 //交易哈希值
	Index         int 				  //在创建者创建的交易集序列中的索引
	//未达到验证阈值的历史交易State(0)，达到验证阈值交易State(1)和待验证新交易State(2)
}

type Message struct {                 //交易集结构
	Body          Eventbody           //交易本身
	Timestamp     int64               //创建交易集的时间戳
	MCreator      []byte              //创造者的公钥（数字签名）
	Number        int                 //交易数量
	MHash         int                 //新建交易集自身哈希值
	Parents       []string			  //交易集父节点哈希值
	Index         int 				  //在创建者创建的交易集序列中的索引
	Transaction   [][]byte            //多笔单个交易state(0,1,2)
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
	return crypto.SHA256(hashBytes), nil
}


type Event struct {                   //交易结构
	Body          Eventbody
	Signature     string              //creator's digital signature of body
	State         int                 //单个交易state(0,1,2)
	//未达到验证阈值的历史交易State(0)，达到验证阈值交易State(1)和待验证新交易State(2)
}

//新建交易
func NewEvent(transaction [][]byte,
	content  []string,
	eCreator []byte,
	eHash   int,
	index   int,
	) *Event {

	body := Eventbody{
		Transaction:  transaction,
		Timestamp:    time.Now().Unix(),
		Content:      content,
		ECreator:     eCreator,
		EHash:        eHash,
		Index:        index,
	}
	return & Event{
		Body: body,
	}
}

