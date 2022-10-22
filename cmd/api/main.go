package main

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"twitch_chat_analysis/cmd/model"
)

func main() {
	r := gin.Default()

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, "worked")
	})

	r.POST("/message", func(c *gin.Context) {
		var messageData model.MessageData
		c.BindJSON(&messageData)
		err := SendMessageToRabbitMqQueue(messageData)
		if err != nil {
			c.JSON(400, err)
		} else {
			c.JSON(200, "worked")
		}
	})

	r.GET("/message/list/:sender/:receiver", func(c *gin.Context) {
		SenderMessageList(c)
	})

	r.Run()
}

func SendMessageToRabbitMqQueue(messageData model.MessageData) error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"message_queue", // name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(messageData)
	if err != nil {
		return err
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	if err != nil {
		return err
	}

	err = SaveMessageToRedis(body, messageData.Sender, messageData.Receiver)
	if err != nil {
		return err
	}

	return nil
}

func SaveMessageToRedis(messageData []byte, sender, receiver string) error {
	var ctx = context.Background()

	var redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	key := sender + ":" + receiver
	_, err := redisClient.Set(ctx, key, messageData, 0).Result()
	if err != nil {
		return err
	}

	return nil

}

func SenderMessageList(c *gin.Context) {
	var ctx = context.Background()

	var redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	key := c.Param("sender") + ":" + c.Param("receiver")

	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		c.JSON(400, err)
	}

	c.JSON(200, val)
}
