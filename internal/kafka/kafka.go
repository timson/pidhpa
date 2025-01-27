package kafka

import (
	"context"
	"errors"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kversion"
	"github.com/twmb/franz-go/pkg/sasl"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

var ErrConsumerGroupNotStable = errors.New("consumer group is not stable")
var ErrNoKafkaClient = errors.New("no connection to Kafka")
var ErrTopicNotFound = errors.New("topic not found")
var ErrGroupNotFound = errors.New("group not found")

func GetKafkaLag(ctx context.Context, kafkaClient *kadm.Client, group string, topic string) (int64, error) {
	if kafkaClient != nil {
		lags, err := kafkaClient.Lag(ctx, group)
		if err != nil {
			return 0, err
		}
		lag, found := lags[group]
		if !found {
			return 0, ErrGroupNotFound
		}
		if lag.State != "Stable" {
			return 0, ErrConsumerGroupNotStable
		}
		lagsByTopic := lag.Lag.TotalByTopic()
		topicLag, topicFound := lagsByTopic[topic]
		if !topicFound {
			return 0, ErrTopicNotFound
		}
		return topicLag.Lag, nil
	}
	return 0, ErrNoKafkaClient
}

func NewKafkaClient(brokers []string, useSASL bool, saslMechanism string, username string, password string) (*kadm.Client, error) {
	options := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.MaxVersions(kversion.V2_4_0()),
	}
	if useSASL == true {
		var sm sasl.Mechanism
		if saslMechanism == "plain" {
			sm = plain.Auth{User: username, Pass: password}.AsMechanism()
		} else if saslMechanism == "scram_sha256" {
			sm = scram.Auth{User: username, Pass: password}.AsSha256Mechanism()
		} else if saslMechanism == "scram_sha512" {
			sm = scram.Auth{User: username, Pass: password}.AsSha512Mechanism()
		}
		options = append(options, kgo.SASL(sm))
	}
	kafkaClient, err := kgo.NewClient(options...)
	if err != nil {
		return nil, err
	}
	return kadm.NewClient(kafkaClient), nil
}
