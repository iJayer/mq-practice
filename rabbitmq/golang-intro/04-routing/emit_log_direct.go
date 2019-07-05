/*
 * 说明：日志生产者
 * 作者：jayer
 * 时间：2019-06-02 10:53 PM
 * 更新：
 */

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ijayer/mq-practice/utils"
	"github.com/streadway/amqp"
)

func main() {
	log.Printf("connecting to rabbitmq[%v]\n", utils.Host)
	// dial with rabbitmq
	conn, err := amqp.Dial(utils.Host)
	utils.FatalOnError(err, fmt.Sprintf("failed to dial with rabbitmq[%v]", utils.Host))
	defer conn.Close()
	log.Printf("connected")

	// create channel
	ch, err := conn.Channel()
	utils.FatalOnError(err, "failed to create channel")
	defer ch.Close()

	// declare exchange
	xType := "direct"
	xName := "logs_direct"
	err = ch.ExchangeDeclare(xName, xType, true, false, false, false, nil)
	utils.FatalOnError(err, "failed to declare an exchange with type(direct)")

	body := utils.BodyFrom(os.Args)

	// publish msg
	err = ch.Publish(
		xName,
		utils.SeverityFrom(os.Args),
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	utils.FatalOnError(err, "failed to publish msg")

	log.Printf("[X] sent: %v\n", body)
}
