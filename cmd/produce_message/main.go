package main

import (
	"context"
	"fmt"
	"time"

	"protokafka/producer"
	"protokafka/proto/kafka/driver_shifts"

	"github.com/Shopify/sarama"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Beat event producer conf
	stack         = "dev"
	schemaVersion = "1"
	serviceName   = "core_business"

	// Send message config
	topic      = "core_business.driver_shifts.v001"
	entityName = "core_driver_performance"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("error: ", err)
	} else {
		fmt.Println("message sent")
	}
}

func run() error {
	ctx := context.Background()

	kb := []string{
		"kafka-0.kafka-headless:9092",
	}

	eventProducer, err := producer.NewBeatEventProducer(kb, stack, schemaVersion, serviceName, sarama.NewConfig())
	if err != nil {
		return err
	}
	defer eventProducer.Close()

	today := time.Now()

	msg := driver_shifts.DriverShifts{
		DriverShifts: []*driver_shifts.DriverShift{
			{
				CityId:             1,
				DriverShiftId:      220,
				ShiftId:            190,
				DriverId:           2,
				VehicleAssigned:    false,
				Type:               2,
				ShiftStartDatetime: timestamppb.New(today),
				ShiftEndDatetime:   timestamppb.New(today.Add(time.Hour * time.Duration(8))),
				DriverFirstName:    "Paco",
				DriverLastName:     "El chato",
			},
			{
				CityId:             1,
				DriverShiftId:      221,
				ShiftId:            191,
				DriverId:           3,
				VehicleAssigned:    false,
				Type:               1,
				ShiftStartDatetime: timestamppb.New(today),
				ShiftEndDatetime:   timestamppb.New(today.Add(time.Hour * time.Duration(8))),
				DriverFirstName:    "Don",
				DriverLastName:     "Gato",
			},
		},
	}

	err = eventProducer.SendMessage(ctx, topic, entityName, &msg)
	if err != nil {
		return err
	}

	return nil
}
