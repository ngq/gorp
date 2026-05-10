package kafka

import (
	"github.com/IBM/sarama"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// buildSaramaConfig builds sarama.Config from MessageQueueConfig.
// 构建 sarama.Config 配置。
func buildSaramaConfig(cfg *integrationcontract.MessageQueueConfig) *sarama.Config {
	saramaCfg := sarama.NewConfig()

	// Client ID
	if cfg.KafkaClientID != "" {
		saramaCfg.ClientID = cfg.KafkaClientID
	}

	// Version
	if cfg.KafkaVersion != "" {
		// sarama.ParseVersion 不存在，使用 V2_8_0_0 常量
		saramaCfg.Version = sarama.V2_8_0_0
	} else {
		saramaCfg.Version = sarama.V2_8_0_0
	}

	// Producer configuration
	saramaCfg.Producer.RequiredAcks = parseRequiredAcks(cfg.KafkaRequiredACKs)
	saramaCfg.Producer.Partitioner = parsePartitioner(cfg.KafkaPartitioner)
	saramaCfg.Producer.MaxMessageBytes = cfg.KafkaMaxMessageBytes
	saramaCfg.Producer.Flush.Frequency = cfg.KafkaFlushFrequency
	saramaCfg.Producer.Compression = parseCompression(cfg.KafkaCompression)
	saramaCfg.Producer.Return.Successes = true // Required for SyncProducer

	// Consumer configuration
	saramaCfg.Consumer.Return.Errors = true

	// TLS configuration
	if cfg.KafkaEnableTLS {
		saramaCfg.Net.TLS.Enable = true
		// If cert files provided, use TLS config
		// For simplicity, we only enable TLS without custom certs here
		// Users can configure via NativeMQClient() if needed
	}

	return saramaCfg
}

// parseRequiredAcks converts int to sarama.RequiredAcks.
// 转换 RequiredAcks 配置。
func parseRequiredAcks(acks int) sarama.RequiredAcks {
	switch acks {
	case 0:
		return sarama.NoResponse
	case 1:
		return sarama.WaitForLocal
	case -1:
		return sarama.WaitForAll
	default:
		return sarama.WaitForAll
	}
}

// parsePartitioner converts string to sarama.PartitionerConstructor.
// 转换 Partitioner 配置。
func parsePartitioner(partitioner string) sarama.PartitionerConstructor {
	switch partitioner {
	case "hash":
		return sarama.NewHashPartitioner
	case "random":
		return sarama.NewRandomPartitioner
	case "round-robin":
		return sarama.NewRoundRobinPartitioner
	default:
		return sarama.NewHashPartitioner
	}
}

// parseCompression converts string to sarama.CompressionCodec.
// 转换 Compression 配置。
func parseCompression(compression string) sarama.CompressionCodec {
	switch compression {
	case "gzip":
		return sarama.CompressionGZIP
	case "snappy":
		return sarama.CompressionSnappy
	case "lz4":
		return sarama.CompressionLZ4
	case "zstd":
		return sarama.CompressionZSTD
	case "none":
		return sarama.CompressionNone
	default:
		return sarama.CompressionNone
	}
}