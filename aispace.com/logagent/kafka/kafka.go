package kafka

import (
	"fmt"
	"time"

	"Golong/aispace.com/logagent/taillog"
	"github.com/Shopify/sarama"
)

var client sarama.SyncProducer

//Init 初始化kafka连接的client
func Init(hosts []string, logchansize int) (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner // 新选出一个partition
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel返回
	// 连接kafka
	client, err = sarama.NewSyncProducer(hosts, config)
	if err != nil {
		fmt.Println("producer closed, err:", err)
		return
	}
	//根据配置大小初始化LogChan 通道
	taillog.LogChan = make(chan *taillog.LogTopic, logchansize)
	//放到后台进行接收日志
	go SendMsg()
	return
}

//SendMsg 向topic中发送数据
func SendMsg() {
	// 从channel 中读取日志结构体构造一个消息
	for {
		select {
		case v := <-taillog.LogChan:
			msg := &sarama.ProducerMessage{}
			msg.Topic = v.Topic
			msg.Value = sarama.StringEncoder(v.Line)
			pid, offset, err := client.SendMessage(msg)
			if err != nil {
				fmt.Println("send msg failed, err:", err)
				return
			}
			fmt.Printf("pid:%v offset:%v\n", pid, offset)
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}
