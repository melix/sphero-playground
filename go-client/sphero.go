package main

import (
	"fmt"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/sphero"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type SpheroParams struct {
	speed uint8
	angle uint16
	color SpheroColor
}

type SpheroColor struct {
	r uint8
	g uint8
	b uint8
}

func kafkaConsumer(params *SpheroParams) {

	broker := "192.168.86.40:32770"
	group := "test"
	topics := []string{"test"}
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		// Avoid connecting to IPv6 brokers:
		// This is needed for the ErrAllBrokersDown show-case below
		// when using localhost brokers on OSX, since the OSX resolver
		// will return the IPv6 addresses first.
		// You typically don't need to specify this configuration property.
		"broker.address.family": "v4",
		"group.id":              group,
		"session.timeout.ms":    6000,
		"auto.offset.reset":     "earliest"})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Consumer %v\n", c)

	err = c.SubscribeTopics(topics, nil)

	run := true

	for run == true {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				msg := string(e.Value)
				key := string(e.Key)
				fmt.Printf("%% %s on %s:\n%s\n",
					key, e.TopicPartition, msg)
				if key == "color" {
					rgbString := strings.Split(msg, ",")
					r, _ := strconv.Atoi(rgbString[0])
					g, _ := strconv.Atoi(rgbString[1])
					b, _ := strconv.Atoi(rgbString[2])
					params.color.r = uint8(r)
					params.color.g = uint8(g)
					params.color.b = uint8(b)
				}
				if key == "speed" {
					speed, _ := strconv.Atoi(msg)
					params.speed = uint8(speed)
				}
				if key == "angle" {
					angle, _ := strconv.Atoi(msg)
					params.angle = uint16(angle)
				}
				if e.Headers != nil {
					fmt.Printf("%% Headers: %v\n", e.Headers)
				}
			case kafka.Error:
				// Errors should generally be considered
				// informational, the client will try to
				// automatically recover.
				// But in this example we choose to terminate
				// the application if all brokers are down.
				fmt.Fprintf(os.Stderr, "%% Error: %v: %v\n", e.Code(), e)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				fmt.Printf("Ignored %v\n", e)
			}
		}
	}

	fmt.Printf("Closing consumer\n")
	c.Close()
}

func spheroMain() {
	adaptor := sphero.NewAdaptor("/dev/rfcomm0")
	driver := sphero.NewSpheroDriver(adaptor)
	driver.ConfigureCollisionDetection(sphero.CollisionConfig{
		Method: 0x1,
		Xt:     0x01,
		Yt:     0x01,
		Xs:     0x01,
		Ys:     0x01,
		Dead:   0x60,
	})

	params := SpheroParams{
		speed: 20,
		angle: 0,
		color: SpheroColor{
			r: 128,
			g: 129,
			b: 65,
		},
	}
	driver.On(sphero.Collision, func(s interface{}) {
		driver.SetRGB(uint8(gobot.Rand(256)), uint8(gobot.Rand(256)), uint8(gobot.Rand(256)))
		params.angle = uint16(int(params.angle) + ( 90 + gobot.Rand(90)) % 360)
	})

	work := func() {
		gobot.Every(100*time.Millisecond, func() {
			driver.SetRGB(params.color.r, params.color.g, params.color.b)
			driver.Roll(params.speed, params.angle)
		})
	}

	robot := gobot.NewRobot("sphero",
		[]gobot.Connection{adaptor},
		[]gobot.Device{driver},
		work,
	)

	go kafkaConsumer(&params)

	fmt.Printf("Starting robot")
	robot.Start()
}

func main() {
	spheroMain()
}