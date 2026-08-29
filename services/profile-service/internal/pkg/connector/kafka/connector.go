package kafka

import (
	"log"

	"profile-service/config"

	"github.com/IBM/sarama"
)

func MustConsumerGroup() sarama.ConsumerGroup {
	cfg := sarama.NewConfig()
	cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	cfg.Consumer.Offsets.AutoCommit.Enable = false
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	group, err := sarama.NewConsumerGroup(config.Instance().Kafka.Brokers, config.Instance().Kafka.ConsumerGroup, cfg)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return group
}
