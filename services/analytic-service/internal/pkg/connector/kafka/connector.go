package kafka

import (
	"log"
	"time"

	"analytic-service/config"

	"github.com/IBM/sarama"
)

func MustConsumerGroup() sarama.ConsumerGroup {
	cfg := sarama.NewConfig()

	cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}

	cfg.Consumer.Group.Rebalance.Retry.Max = 4
	cfg.Consumer.Group.Rebalance.Retry.Backoff = 500 * time.Millisecond

	cfg.Consumer.Group.Session.Timeout = 30 * time.Second
	cfg.Consumer.Group.Heartbeat.Interval = 5 * time.Second

	cfg.Consumer.Offsets.AutoCommit.Enable = false
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	cfg.Consumer.Fetch.Min = 1
	cfg.Consumer.Fetch.Default = 1024 * 1024
	cfg.Consumer.Fetch.Max = 10 * 1024 * 1024
	cfg.Consumer.MaxProcessingTime = 60 * time.Second

	group, err := sarama.NewConsumerGroup(config.Instance().Kafka.Brokers, config.Instance().Kafka.ConsumerGroup, cfg)
	if err != nil {
		log.Fatalf(err.Error())
		return nil
	}

	return group
}

func MustSyncProducer() sarama.SyncProducer {
	saramaConfig := sarama.NewConfig()

	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Return.Successes = true

	saramaConfig.Net.DialTimeout = 5 * time.Second
	saramaConfig.Net.WriteTimeout = 5 * time.Second

	saramaConfig.Producer.Retry.Max = 10
	saramaConfig.Producer.Retry.Backoff = 100 * time.Millisecond
	saramaConfig.Producer.Idempotent = true
	saramaConfig.Producer.Partitioner = sarama.NewRoundRobinPartitioner

	saramaConfig.Producer.Timeout = 10 * time.Second
	saramaConfig.Net.MaxOpenRequests = 1

	saramaConfig.Metadata.Retry.Max = 5
	saramaConfig.Metadata.Retry.Backoff = 500 * time.Millisecond

	producer, err := sarama.NewSyncProducer(config.Instance().Kafka.Brokers, saramaConfig)
	if err != nil {
		log.Fatalf(err.Error())
		return nil
	}

	return producer
}
