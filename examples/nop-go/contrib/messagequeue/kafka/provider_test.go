package kafka

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()

	require.Equal(t, "messagequeue.kafka", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{
		integrationcontract.MessageQueueKey,
		integrationcontract.MessagePublisherKey,
		integrationcontract.MessageSubscriberKey,
	}, p.Provides())
}

func TestProviderNewQueue(t *testing.T) {
	// Skip if no Kafka available - this is a unit test for contract only
	// Integration tests require real Kafka instance
	t.Skip("requires Kafka instance for integration test")
}

func TestBuildSaramaConfig(t *testing.T) {
	cfg := &integrationcontract.MessageQueueConfig{
		Type:                 "kafka",
		KafkaBrokers:         []string{"localhost:9092"},
		KafkaClientID:        "test-client",
		KafkaVersion:         "2.8.0",
		KafkaRequiredACKs:    -1,
		KafkaPartitioner:     "hash",
		KafkaCompression:     "gzip",
		KafkaMaxMessageBytes: 1000000,
	}

	saramaCfg := buildSaramaConfig(cfg)

	require.Equal(t, "test-client", saramaCfg.ClientID)
	require.Equal(t, sarama.WaitForAll, saramaCfg.Producer.RequiredAcks)
	require.Equal(t, sarama.CompressionGZIP, saramaCfg.Producer.Compression)
	require.True(t, saramaCfg.Producer.Return.Successes)
}

func TestParseRequiredAcks(t *testing.T) {
	require.Equal(t, sarama.NoResponse, parseRequiredAcks(0))
	require.Equal(t, sarama.WaitForLocal, parseRequiredAcks(1))
	require.Equal(t, sarama.WaitForAll, parseRequiredAcks(-1))
}

func TestParsePartitioner(t *testing.T) {
	require.NotNil(t, parsePartitioner("hash"))
	require.NotNil(t, parsePartitioner("random"))
	require.NotNil(t, parsePartitioner("round-robin"))
	require.NotNil(t, parsePartitioner("unknown")) // defaults to hash
}

func TestParseCompression(t *testing.T) {
	require.Equal(t, sarama.CompressionGZIP, parseCompression("gzip"))
	require.Equal(t, sarama.CompressionSnappy, parseCompression("snappy"))
	require.Equal(t, sarama.CompressionLZ4, parseCompression("lz4"))
	require.Equal(t, sarama.CompressionZSTD, parseCompression("zstd"))
	require.Equal(t, sarama.CompressionNone, parseCompression("none"))
	require.Equal(t, sarama.CompressionNone, parseCompression("unknown"))
}
