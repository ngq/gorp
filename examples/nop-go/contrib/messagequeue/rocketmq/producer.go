// Package rocketmq provides RocketMQ Publisher implementation.
// This file implements the MessagePublisher contract using rocketmq-client-go SDK.
//
// 本包提供 RocketMQ Publisher 实现。
// 本文件使用 rocketmq-client-go SDK 实现 MessagePublisher 契约。
package rocketmq

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// rocketmqPublisher implements MessagePublisher using rocketmq-client-go.
// Wraps the Queue to access the underlying producer.
//
// rocketmqPublisher 使用 rocketmq-client-go 实现 MessagePublisher。
// 包装 Queue 以访问底层 producer。
type rocketmqPublisher struct {
	queue *Queue
}

// Publish sends a message to a topic.
// Implements integrationcontract.MessagePublisher.Publish.
// Supports custom headers via PublishOption.
//
// Publish 将消息发送到 topic。
// 实现 integrationcontract.MessagePublisher.Publish。
// 通过 PublishOption 支持自定义 headers。
func (p *rocketmqPublisher) Publish(ctx context.Context, topic string, message []byte, options ...integrationcontract.PublishOption) error {
	if p == nil || p.queue == nil || p.queue.producer == nil {
		return errors.New("messagequeue.rocketmq: producer not initialized")
	}

	// Apply publish options
	cfg := &integrationcontract.PublishConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	// Create message
	msg := &primitive.Message{
		Topic: topic,
		Body:  message,
	}

	// Apply headers if provided
	if len(cfg.Headers) > 0 {
		for k, v := range cfg.Headers {
			msg.WithProperty(k, v)
		}
	}

	// Send synchronously and check result
	result, err := p.queue.producer.SendSync(ctx, msg)
	if err != nil {
		return err
	}
	if result.Status != primitive.SendOK {
		return fmt.Errorf("messagequeue.rocketmq: send failed with status %d", result.Status)
	}
	return nil
}

// PublishWithDelay sends a delayed message using DelayTimeLevel.
// RocketMQ supports 18 delay levels:
// 1: 1s, 2: 5s, 3: 10s, 4: 30s, 5: 1m, 6: 2m, 7: 3m, 8: 4m, 9: 5m, 10: 6m,
// 11: 7m, 12: 8m, 13: 9m, 14: 10m, 15: 20m, 16: 30m, 17: 1h, 18: 2h
//
// PublishWithDelay 发送延迟消息，使用 DelayTimeLevel。
// RocketMQ 支持 18 个延迟等级。
func (p *rocketmqPublisher) PublishWithDelay(ctx context.Context, topic string, message []byte, delay time.Duration) error {
	if p == nil || p.queue == nil || p.queue.producer == nil {
		return errors.New("messagequeue.rocketmq: producer not initialized")
	}

	// Convert delay duration to RocketMQ delay level (1-18)
	delayLevel := parseDelayLevel(delay)
	msg := &primitive.Message{
		Topic: topic,
		Body:  message,
	}
	msg.WithDelayTimeLevel(delayLevel)

	// Send synchronously and check result
	result, err := p.queue.producer.SendSync(ctx, msg)
	if err != nil {
		return err
	}
	if result.Status != primitive.SendOK {
		return fmt.Errorf("messagequeue.rocketmq: send failed with status %d", result.Status)
	}
	return nil
}

// PublishWithPriority sends a message with priority.
// RocketMQ does not have native priority, but we use tag for priority-like routing.
// This allows consumers to filter messages by priority tag.
//
// PublishWithPriority 发送带优先级的消息。
// RocketMQ 无原生优先级，但使用 tag 实现类似优先级的路由。
// 这允许消费者通过优先级 tag 过滤消息。
func (p *rocketmqPublisher) PublishWithPriority(ctx context.Context, topic string, message []byte, priority int) error {
	if p == nil || p.queue == nil || p.queue.producer == nil {
		return errors.New("messagequeue.rocketmq: producer not initialized")
	}

	msg := &primitive.Message{
		Topic: topic,
		Body:  message,
	}
	// Use tag for priority-like routing
	msg.WithTag(fmt.Sprintf("priority-%d", priority))

	// Send synchronously and check result
	result, err := p.queue.producer.SendSync(ctx, msg)
	if err != nil {
		return err
	}
	if result.Status != primitive.SendOK {
		return fmt.Errorf("messagequeue.rocketmq: send failed with status %d", result.Status)
	}
	return nil
}

// Send sends a message treating queue as topic with tag.
// Implements integrationcontract.MessagePublisher.Send.
// In RocketMQ, queue format is "topic:tag" where tag is optional.
//
// Send 将消息发送到 queue（格式为 topic:tag）。
// 实现 integrationcontract.MessagePublisher.Send。
// 在 RocketMQ 中，queue 格式为 "topic:tag"，tag 可选。
func (p *rocketmqPublisher) Send(ctx context.Context, queue string, message []byte, options ...integrationcontract.PublishOption) error {
	if p == nil || p.queue == nil || p.queue.producer == nil {
		return errors.New("messagequeue.rocketmq: producer not initialized")
	}

	// Parse queue as topic:tag format
	parts := strings.SplitN(queue, ":", 2)
	topic := parts[0]
	tag := ""
	if len(parts) > 1 {
		tag = parts[1]
	}

	// Apply publish options
	cfg := &integrationcontract.PublishConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	// Create message
	msg := &primitive.Message{
		Topic: topic,
		Body:  message,
	}
	if tag != "" {
		msg.WithTag(tag)
	}

	// Send synchronously and check result
	result, err := p.queue.producer.SendSync(ctx, msg)
	if err != nil {
		return err
	}
	if result.Status != primitive.SendOK {
		return fmt.Errorf("messagequeue.rocketmq: send failed with status %d", result.Status)
	}
	return nil
}

// Underlying returns the underlying rocketmq.Producer.
// This allows users to access native SDK capabilities directly.
//
// Underlying 返回底层 rocketmq.Producer。
// 这允许用户直接访问原生 SDK 能力。
func (p *rocketmqPublisher) Underlying() any {
	if p == nil || p.queue == nil {
		return nil
	}
	return p.queue.producer
}

// As attempts to cast the underlying producer to the target type.
// Uses the internal native.As helper for type casting.
//
// As 尝试将底层 producer 转换为目标类型。
// 使用内部 native.As 辅助函数进行类型转换。
func (p *rocketmqPublisher) As(target any) bool {
	if p == nil || p.queue == nil || p.queue.producer == nil {
		return false
	}
	return As(p.queue.producer, target)
}

// NativePublisher implements NativePublisherProvider interface.
// Returns the underlying rocketmq.Producer for native SDK access.
//
// NativePublisher 实现 NativePublisherProvider 接口。
// 返回底层 rocketmq.Producer 用于原生 SDK 访问。
func (p *rocketmqPublisher) NativePublisher() any {
	return p.Underlying()
}

// parseDelayLevel converts delay duration to RocketMQ delay level (1-18).
// RocketMQ supports predefined delay levels, not arbitrary durations.
// The function maps the requested delay to the closest available level.
//
// parseDelayLevel 将延迟时间转换为 RocketMQ 延迟等级（1-18）。
// RocketMQ 支持预定义的延迟等级，不支持任意延迟时间。
// 该函数将请求的延迟映射到最接近的可用等级。
func parseDelayLevel(delay time.Duration) int {
	// RocketMQ delay levels:
	// 1: 1s, 2: 5s, 3: 10s, 4: 30s, 5: 1m, 6: 2m, 7: 3m, 8: 4m, 9: 5m, 10: 6m,
	// 11: 7m, 12: 8m, 13: 9m, 14: 10m, 15: 20m, 16: 30m, 17: 1h, 18: 2h
	switch {
	case delay <= time.Second:
		return 1 // 1s
	case delay <= 5*time.Second:
		return 2 // 5s
	case delay <= 10*time.Second:
		return 3 // 10s
	case delay <= 30*time.Second:
		return 4 // 30s
	case delay <= time.Minute:
		return 5 // 1m
	case delay <= 2*time.Minute:
		return 6 // 2m
	case delay <= 3*time.Minute:
		return 7 // 3m
	case delay <= 4*time.Minute:
		return 8 // 4m
	case delay <= 5*time.Minute:
		return 9 // 5m
	case delay <= 6*time.Minute:
		return 10 // 6m
	case delay <= 7*time.Minute:
		return 11 // 7m
	case delay <= 8*time.Minute:
		return 12 // 8m
	case delay <= 9*time.Minute:
		return 13 // 9m
	case delay <= 10*time.Minute:
		return 14 // 10m
	case delay <= 20*time.Minute:
		return 15 // 20m
	case delay <= 30*time.Minute:
		return 16 // 30m
	case delay <= time.Hour:
		return 17 // 1h
	default:
		return 18 // 2h
	}
}
