// Package websocket provides WebSocket capability provider for gorp framework.
// This file implements cluster mode using Redis Pub/Sub for cross-node broadcast.
//
// WebSocket 包提供 gorp 框架的 WebSocket 能力 provider。
// 本文件使用 Redis Pub/Sub 实现集群模式的跨节点广播。
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/redis/go-redis/v9"
)

// ClusterMessage represents a message broadcast across nodes.
//
// ClusterMessage 表示跨节点广播的消息。
type ClusterMessage struct {
	Type      string `json:"type"`      // "broadcast", "room", "user"
	RoomID    string `json:"room_id"`   // for room broadcast
	UserID    string `json:"user_id"`   // for user-specific message
	Message   string `json:"message"`   // text message
	Binary    []byte `json:"binary"`    // binary message (base64)
	IsBinary  bool   `json:"is_binary"` // whether this is a binary message
	SenderID  string `json:"sender_id"` // node ID of sender
	Timestamp int64  `json:"timestamp"` // unix timestamp
}

// ClusterServer implements WebSocketClusterServer with Redis Pub/Sub.
//
// ClusterServer 使用 Redis Pub/Sub 实现 WebSocketClusterServer。
type ClusterServer struct {
	*Server // embed single-node server
	config  *transportcontract.WebSocketClusterConfig
	redis   *redis.Client
	nodeID  string
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	rooms   sync.Map // roomID -> map[transportcontract.WebSocketConn]bool
}

// NewClusterServer creates a new cluster-enabled WebSocket server.
//
// NewClusterServer 创建新的支持集群的 WebSocket 服务器。
func NewClusterServer(wsConfig *transportcontract.WebSocketConfig, clusterConfig *transportcontract.WebSocketClusterConfig) (*ClusterServer, error) {
	if clusterConfig == nil {
		clusterConfig = &transportcontract.WebSocketClusterConfig{Enabled: false}
	}

	server := &ClusterServer{
		Server: NewServerWithConfig(wsConfig),
		config: clusterConfig,
	}

	if clusterConfig.Enabled {
		// Generate node ID if not provided
		// 如果未提供则生成节点 ID
		if clusterConfig.NodeID == "" {
			server.nodeID = generateNodeID()
		} else {
			server.nodeID = clusterConfig.NodeID
		}

		// Initialize Redis client
		// 初始化 Redis 客户端
		server.redis = redis.NewClient(&redis.Options{
			Addr:     clusterConfig.RedisAddr,
			Password: clusterConfig.RedisPassword,
			DB:       clusterConfig.RedisDB,
		})

		// Test Redis connection
		// 测试 Redis 连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := server.redis.Ping(ctx).Err(); err != nil {
			cancel()
			return nil, fmt.Errorf("connect to redis: %w", err)
		}
		cancel()

		// Setup context for goroutines
		// 设置 goroutine 的上下文
		server.ctx, server.cancel = context.WithCancel(context.Background())

		// Start Pub/Sub subscription
		// 启动 Pub/Sub 订阅
		server.startSubscription()

		// Start heartbeat for node health tracking
		// 启动节点健康心跳
		if clusterConfig.EnableGlobalCount {
			server.startHeartbeat()
		}
	}

	return server, nil
}

// startSubscription starts Redis Pub/Sub subscription for cross-node messages.
//
// startSubscription 启动 Redis Pub/Sub 订阅以接收跨节点消息。
func (s *ClusterServer) startSubscription() {
	channelPrefix := s.config.ChannelPrefix
	if channelPrefix == "" {
		channelPrefix = "gorp:ws"
	}

	channel := fmt.Sprintf("%s:broadcast", channelPrefix)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		pubsub := s.redis.Subscribe(s.ctx, channel)
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case <-s.ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				s.handleClusterMessage(msg.Payload)
			}
		}
	}()
}

// handleClusterMessage handles incoming cluster message from Redis.
//
// handleClusterMessage 处理从 Redis 接收的集群消息。
func (s *ClusterServer) handleClusterMessage(payload string) {
	var msg ClusterMessage
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		return
	}

	// Ignore messages from self
	// 忽略来自自己的消息
	if msg.SenderID == s.nodeID {
		return
	}

	switch msg.Type {
	case "broadcast":
		s.broadcastLocal(msg)
	case "room":
		s.broadcastToRoomLocal(msg.RoomID, msg)
	case "user":
		s.sendToUserLocal(msg.UserID, msg)
	}
}

// broadcastLocal broadcasts message to local connections.
//
// broadcastLocal 向本地连接广播消息。
func (s *ClusterServer) broadcastLocal(msg ClusterMessage) {
	broadcaster := s.Server.NewBroadcaster()
	if msg.IsBinary {
		broadcaster.BroadcastBinary(msg.Binary)
	} else {
		broadcaster.BroadcastString(msg.Message)
	}
}

// broadcastToRoomLocal broadcasts message to local room members.
//
// broadcastToRoomLocal 向本地房间成员广播消息。
func (s *ClusterServer) broadcastToRoomLocal(roomID string, msg ClusterMessage) {
	value, ok := s.rooms.Load(roomID)
	if !ok {
		return
	}

	members := value.(*sync.Map)
	members.Range(func(key, value any) bool {
		conn := key.(transportcontract.WebSocketConn)
		if msg.IsBinary {
			conn.WriteBinary(msg.Binary)
		} else {
			conn.WriteString(msg.Message)
		}
		return true
	})
}

// sendToUserLocal sends message to a specific user on this node.
//
// sendToUserLocal 向本节点的特定用户发送消息。
func (s *ClusterServer) sendToUserLocal(userID string, msg ClusterMessage) {
	// Find user connection by context
	// 通过上下文查找用户连接
	for _, conn := range s.Server.Connections() {
		ctx := conn.Context()
		if ctx != nil {
			if uid, ok := ctx.Value("user_id").(string); ok && uid == userID {
				if msg.IsBinary {
					conn.WriteBinary(msg.Binary)
				} else {
					conn.WriteString(msg.Message)
				}
				return
			}
		}
	}
}

// startHeartbeat starts periodic heartbeat for node health tracking.
//
// startHeartbeat 启动周期性心跳以跟踪节点健康状态。
func (s *ClusterServer) startHeartbeat() {
	countKeyPrefix := s.config.CountKeyPrefix
	if countKeyPrefix == "" {
		countKeyPrefix = "gorp:ws:count"
	}

	interval := s.config.HeartbeatInterval
	if interval <= 0 {
		interval = 30
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		key := fmt.Sprintf("%s:node:%s", countKeyPrefix, s.nodeID)

		for {
			select {
			case <-s.ctx.Done():
				// Remove node key on shutdown with timeout
				// 关闭时移除节点键，使用带超时的 context
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				s.redis.Del(ctx, key)
				cancel()
				return
			case <-ticker.C:
				// Update node count with expiry
				// 更新节点连接数并设置过期时间
				count := s.Server.Count()
				s.redis.Set(context.Background(), key, count, time.Duration(interval*3)*time.Second)
			}
		}
	}()
}

// Upgrade upgrades HTTP connection to WebSocket and tracks in cluster.
//
// Upgrade 将 HTTP 连接升级为 WebSocket 并在集群中跟踪。
func (s *ClusterServer) Upgrade(w http.ResponseWriter, r *http.Request, handler transportcontract.WebSocketHandler) (transportcontract.WebSocketConn, error) {
	conn, err := s.Server.Upgrade(w, r, handler)
	if err != nil {
		return nil, err
	}

	// Update cluster count if enabled
	// 如果启用则更新集群连接数
	if s.config.Enabled && s.config.EnableGlobalCount {
		countKeyPrefix := s.config.CountKeyPrefix
		if countKeyPrefix == "" {
			countKeyPrefix = "gorp:ws:count"
		}
		key := fmt.Sprintf("%s:node:%s", countKeyPrefix, s.nodeID)
		s.redis.Incr(context.Background(), key)
	}

	return conn, nil
}

// JoinRoom adds a connection to a room.
//
// JoinRoom 将连接加入房间。
func (s *ClusterServer) JoinRoom(roomID string, conn transportcontract.WebSocketConn) {
	value, _ := s.rooms.LoadOrStore(roomID, &sync.Map{})
	members := value.(*sync.Map)
	members.Store(conn, true)
}

// LeaveRoom removes a connection from a room.
//
// LeaveRoom 将连接移出房间。
func (s *ClusterServer) LeaveRoom(roomID string, conn transportcontract.WebSocketConn) {
	if value, ok := s.rooms.Load(roomID); ok {
		members := value.(*sync.Map)
		members.Delete(conn)
	}
}

// BroadcastGlobal broadcasts message to all nodes in the cluster.
//
// BroadcastGlobal 向集群所有节点广播消息。
func (s *ClusterServer) BroadcastGlobal(message string) error {
	if !s.config.Enabled {
		// Fallback to local broadcast if cluster not enabled
		// 如果未启用集群则回退到本地广播
		return s.Server.NewBroadcaster().BroadcastString(message)
	}

	msg := ClusterMessage{
		Type:      "broadcast",
		Message:   message,
		IsBinary:  false,
		SenderID:  s.nodeID,
		Timestamp: time.Now().Unix(),
	}

	return s.publishClusterMessage(msg)
}

// BroadcastGlobalBinary broadcasts binary message to all nodes in the cluster.
//
// BroadcastGlobalBinary 向集群所有节点广播二进制消息。
func (s *ClusterServer) BroadcastGlobalBinary(data []byte) error {
	if !s.config.Enabled {
		return s.Server.NewBroadcaster().BroadcastBinary(data)
	}

	msg := ClusterMessage{
		Type:      "broadcast",
		Binary:    data,
		IsBinary:  true,
		SenderID:  s.nodeID,
		Timestamp: time.Now().Unix(),
	}

	return s.publishClusterMessage(msg)
}

// BroadcastToRoom broadcasts message to a room across all nodes.
//
// BroadcastToRoom 向房间广播消息（跨节点）。
func (s *ClusterServer) BroadcastToRoom(roomID string, message string) error {
	if !s.config.Enabled {
		s.broadcastToRoomLocal(roomID, ClusterMessage{Message: message})
		return nil
	}

	msg := ClusterMessage{
		Type:      "room",
		RoomID:    roomID,
		Message:   message,
		IsBinary:  false,
		SenderID:  s.nodeID,
		Timestamp: time.Now().Unix(),
	}

	return s.publishClusterMessage(msg)
}

// SendToUser sends message to a specific user across all nodes.
//
// SendToUser 向特定用户发送消息（跨节点）。
func (s *ClusterServer) SendToUser(userID string, message string) error {
	if !s.config.Enabled {
		s.sendToUserLocal(userID, ClusterMessage{Message: message})
		return nil
	}

	msg := ClusterMessage{
		Type:      "user",
		UserID:    userID,
		Message:   message,
		IsBinary:  false,
		SenderID:  s.nodeID,
		Timestamp: time.Now().Unix(),
	}

	return s.publishClusterMessage(msg)
}

// publishClusterMessage publishes message to Redis channel.
//
// publishClusterMessage 将消息发布到 Redis 通道。
func (s *ClusterServer) publishClusterMessage(msg ClusterMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	channelPrefix := s.config.ChannelPrefix
	if channelPrefix == "" {
		channelPrefix = "gorp:ws"
	}

	channel := fmt.Sprintf("%s:broadcast", channelPrefix)
	return s.redis.Publish(s.ctx, channel, data).Err()
}

// GlobalCount returns total connection count across all nodes.
//
// GlobalCount 返回所有节点的总连接数。
func (s *ClusterServer) GlobalCount() int {
	if !s.config.Enabled || !s.config.EnableGlobalCount {
		return s.Server.Count()
	}

	countKeyPrefix := s.config.CountKeyPrefix
	if countKeyPrefix == "" {
		countKeyPrefix = "gorp:ws:count"
	}

	pattern := fmt.Sprintf("%s:node:*", countKeyPrefix)
	keys, err := s.redis.Keys(s.ctx, pattern).Result()
	if err != nil {
		return s.Server.Count()
	}

	total := 0
	for _, key := range keys {
		val, err := s.redis.Get(s.ctx, key).Int()
		if err != nil {
			continue
		}
		total += val
	}

	return total
}

// NodeCount returns connection count for a specific node.
//
// NodeCount 返回指定节点的连接数。
func (s *ClusterServer) NodeCount(nodeID string) int {
	if !s.config.Enabled || !s.config.EnableGlobalCount {
		if nodeID == s.nodeID {
			return s.Server.Count()
		}
		return 0
	}

	countKeyPrefix := s.config.CountKeyPrefix
	if countKeyPrefix == "" {
		countKeyPrefix = "gorp:ws:count"
	}

	key := fmt.Sprintf("%s:node:%s", countKeyPrefix, nodeID)
	val, err := s.redis.Get(s.ctx, key).Int()
	if err != nil {
		return 0
	}
	return val
}

// ListNodes returns all active node IDs.
//
// ListNodes 返回所有活跃节点 ID。
func (s *ClusterServer) ListNodes() []string {
	if !s.config.Enabled || !s.config.EnableGlobalCount {
		return []string{s.nodeID}
	}

	countKeyPrefix := s.config.CountKeyPrefix
	if countKeyPrefix == "" {
		countKeyPrefix = "gorp:ws:count"
	}

	pattern := fmt.Sprintf("%s:node:*", countKeyPrefix)
	keys, err := s.redis.Keys(s.ctx, pattern).Result()
	if err != nil {
		return []string{s.nodeID}
	}

	nodes := make([]string, 0, len(keys))
	prefix := fmt.Sprintf("%s:node:", countKeyPrefix)
	for _, key := range keys {
		nodeID := key[len(prefix):]
		nodes = append(nodes, nodeID)
	}

	return nodes
}

// GetNodeID returns this node's ID.
//
// GetNodeID 返回本节点 ID。
func (s *ClusterServer) GetNodeID() string {
	return s.nodeID
}

// Shutdown gracefully closes all connections and stops cluster goroutines.
//
// Shutdown 优雅关闭所有连接并停止集群 goroutine。
func (s *ClusterServer) Shutdown(ctx context.Context) error {
	// Stop cluster goroutines
	// 停止集群 goroutine
	if s.cancel != nil {
		s.cancel()
		s.wg.Wait()
	}

	// Close Redis connection
	// 关闭 Redis 连接
	if s.redis != nil {
		s.redis.Close()
	}

	// Shutdown local server
	// 关闭本地服务器
	return s.Server.Shutdown(ctx)
}

// generateNodeID generates a unique node ID.
//
// generateNodeID 生成唯一的节点 ID。
func generateNodeID() string {
	return fmt.Sprintf("node-%d", time.Now().UnixNano())
}
